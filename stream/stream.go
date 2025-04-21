package stream

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	unixEpochTime = time.Unix(0, 0)
	errSeeker     = errors.New("seeker can't seek")
	errNoOverlap  = errors.New("invalid range: failed to overlap")
)

// The algorithm uses at most sniffLen bytes to make its decision.
const sniffLen = 512

// bufferPool manages reusable byte buffers to reduce GC pressure
var bufferPool = sync.Pool{
	New: func() interface{} {
		buffer := make([]byte, 32*1024) // 32KB buffers
		return &buffer
	},
}

func isZeroTime(t time.Time) bool {
	return t.IsZero() || t.Equal(unixEpochTime)
}

func setLastModified(w http.ResponseWriter, modtime time.Time) {
	if !isZeroTime(modtime) {
		w.Header().Set("Last-Modified", modtime.UTC().Format(http.TimeFormat))
	}
}

type godebug struct {
	value string
}

func (g *godebug) New(value string) *godebug {
	return &godebug{value: value}
}

func (g *godebug) Value() string {
	return g.value
}

func (g *godebug) IncNonDefault() {}

var httpservecontentkeepheaders = (&godebug{}).New("httpservecontentkeepheaders")

func Has(h http.Header, key string) bool {
	_, ok := h[key]
	return ok
}

func serveError(w http.ResponseWriter, text string, code int) {
	h := w.Header()

	nonDefault := false
	// Fast path for common case - avoid string iteration
	if httpservecontentkeepheaders.Value() != "1" {
		h.Del("Cache-Control")
		h.Del("Content-Encoding")
		h.Del("Etag")
		h.Del("Last-Modified")
	} else {
		for _, k := range []string{
			"Cache-Control",
			"Content-Encoding",
			"Etag",
			"Last-Modified",
		} {
			if Has(h, k) {
				nonDefault = true
				break
			}
		}
	}

	if nonDefault {
		httpservecontentkeepheaders.IncNonDefault()
	}

	http.Error(w, text, code)
}

// Fast path for ETag scanning with fewer allocations
func scanETag(s string) (etag string, remain string) {
	s = textproto.TrimString(s)
	start := 0
	if len(s) >= 2 && s[0] == 'W' && s[1] == '/' {
		start = 2
	}
	if len(s[start:]) < 2 || s[start] != '"' {
		return "", ""
	}

	for i := start + 1; i < len(s); i++ {
		c := s[i]
		switch {
		case c == 0x21 || c >= 0x23 && c <= 0x7E || c >= 0x80:
			// Valid character
		case c == '"':
			return s[:i+1], s[i+1:]
		default:
			return "", ""
		}
	}
	return "", ""
}

// Optimized ETag matching with fewer allocations
func etagStrongMatch(a, b string) bool {
	return a == b && a != "" && a[0] == '"'
}

func etagWeakMatch(a, b string) bool {
	aPrefix := strings.HasPrefix(a, "W/")
	bPrefix := strings.HasPrefix(b, "W/")

	if aPrefix {
		a = a[2:]
	}
	if bPrefix {
		b = b[2:]
	}

	return a == b
}

type condResult int

const (
	condNone condResult = iota
	condTrue
	condFalse
)

func cond(result bool) condResult {
	if result {
		return condTrue
	}
	return condFalse
}

func checkIfMatch(w http.ResponseWriter, r *http.Request) condResult {
	im := r.Header.Get("If-Match")
	if im == "" {
		return condNone
	}

	etag := w.Header().Get("Etag")
	if etag == "" {
		return condFalse
	}

	if im == "*" {
		return condTrue
	}

	for {
		im = textproto.TrimString(im)
		if len(im) == 0 {
			break
		}
		if im[0] == ',' {
			im = im[1:]
			continue
		}
		if im[0] == '*' {
			return condTrue
		}
		tag, remain := scanETag(im)
		if tag == "" {
			break
		}
		if etagStrongMatch(tag, etag) {
			return condTrue
		}
		im = remain
	}

	return condFalse
}

func checkIfUnmodifiedSince(r *http.Request, modtime time.Time) condResult {
	ius := r.Header.Get("If-Unmodified-Since")
	if ius == "" || isZeroTime(modtime) {
		return condNone
	}
	t, err := http.ParseTime(ius)
	if err != nil {
		return condNone
	}

	modtime = modtime.Truncate(time.Second)
	if ret := modtime.Compare(t); ret <= 0 {
		return condTrue
	}
	return condFalse
}

func checkIfNoneMatch(w http.ResponseWriter, r *http.Request) condResult {
	inm := r.Header.Get("If-None-Match")
	if inm == "" {
		return condNone
	}

	etag := w.Header().Get("Etag")
	if etag == "" {
		return condTrue
	}

	if inm == "*" {
		return condFalse
	}

	buf := inm
	for {
		buf = textproto.TrimString(buf)
		if len(buf) == 0 {
			break
		}
		if buf[0] == ',' {
			buf = buf[1:]
			continue
		}
		if buf[0] == '*' {
			return condFalse
		}
		tag, remain := scanETag(buf)
		if tag == "" {
			break
		}
		if etagWeakMatch(tag, etag) {
			return condFalse
		}
		buf = remain
	}
	return condTrue
}

func checkIfModifiedSince(r *http.Request, modtime time.Time) condResult {
	if r.Method != "GET" && r.Method != "HEAD" {
		return condNone
	}
	ims := r.Header.Get("If-Modified-Since")
	if ims == "" || isZeroTime(modtime) {
		return condNone
	}
	t, err := http.ParseTime(ims)
	if err != nil {
		return condNone
	}

	modtime = modtime.Truncate(time.Second)
	if ret := modtime.Compare(t); ret <= 0 {
		return condFalse
	}
	return condTrue
}

func checkIfRange(w http.ResponseWriter, r *http.Request, modtime time.Time) condResult {
	if r.Method != "GET" && r.Method != "HEAD" {
		return condNone
	}
	ir := r.Header.Get("If-Range")
	if ir == "" {
		return condNone
	}
	etag, _ := scanETag(ir)
	if etag != "" {
		return cond(etagStrongMatch(etag, w.Header().Get("Etag")))
	}

	if modtime.IsZero() {
		return condFalse
	}
	t, err := http.ParseTime(ir)
	if err != nil {
		return condFalse
	}
	if t.Unix() == modtime.Unix() {
		return condTrue
	}
	return condFalse
}

func writeNotModified(w http.ResponseWriter) {
	h := w.Header()
	delete(h, "Content-Type")
	delete(h, "Content-Length")
	delete(h, "Content-Encoding")
	if h.Get("Etag") != "" {
		delete(h, "Last-Modified")
	}
	w.WriteHeader(http.StatusNotModified)
}

func checkPreconditions(w http.ResponseWriter, r *http.Request, modtime time.Time) (done bool, rangeHeader string) {
	ch := checkIfMatch(w, r)
	if ch == condNone {
		ch = checkIfUnmodifiedSince(r, modtime)
	}
	if ch == condFalse {
		w.WriteHeader(http.StatusPreconditionFailed)
		return true, ""
	}

	switch checkIfNoneMatch(w, r) {
	case condFalse:
		if r.Method == "GET" || r.Method == "HEAD" {
			writeNotModified(w)
			return true, ""
		} else {
			w.WriteHeader(http.StatusPreconditionFailed)
			return true, ""
		}
	case condNone:
		if checkIfModifiedSince(r, modtime) == condFalse {
			writeNotModified(w)
			return true, ""
		}
	}

	rangeHeader = r.Header.Get("Range")
	if rangeHeader != "" && checkIfRange(w, r, modtime) == condFalse {
		rangeHeader = ""
	}
	return false, rangeHeader
}

type httpRange struct {
	start, length int64
}

func (r httpRange) contentRange(size int64) string {
	return fmt.Sprintf("bytes %d-%d/%d", r.start, r.start+r.length-1, size)
}

func (r httpRange) mimeHeader(contentType string, size int64) textproto.MIMEHeader {
	return textproto.MIMEHeader{
		"Content-Range": {r.contentRange(size)},
		"Content-Type":  {contentType},
	}
}

func sumRangesSize(ranges []httpRange) (size int64) {
	for _, ra := range ranges {
		size += ra.length
	}
	return
}

type countingWriter int64

func (w *countingWriter) Write(p []byte) (n int, err error) {
	*w += countingWriter(len(p))
	return len(p), nil
}

func rangesMIMESize(ranges []httpRange, contentType string, contentSize int64) (encSize int64) {
	var w countingWriter
	mw := multipart.NewWriter(&w)
	for _, ra := range ranges {
		mw.CreatePart(ra.mimeHeader(contentType, contentSize))
		encSize += ra.length
	}
	mw.Close()
	encSize += int64(w)
	return
}

// Optimized range parsing with fewer allocations
func parseRange(s string, size int64) ([]httpRange, error) {
	if s == "" {
		return nil, nil
	}
	const b = "bytes="
	if !strings.HasPrefix(s, b) {
		return nil, errors.New("invalid range")
	}

	// Pre-allocate for common case (1-2 ranges)
	ranges := make([]httpRange, 0, 2)
	noOverlap := false

	parts := strings.Split(s[len(b):], ",")
	for _, ra := range parts {
		ra = textproto.TrimString(ra)
		if ra == "" {
			continue
		}
		start, end, ok := strings.Cut(ra, "-")
		if !ok {
			return nil, errors.New("invalid range")
		}
		start, end = textproto.TrimString(start), textproto.TrimString(end)
		var r httpRange
		if start == "" {
			// Format: -N (last N bytes)
			if end == "" || end[0] == '-' {
				return nil, errors.New("invalid range")
			}
			i, err := strconv.ParseInt(end, 10, 64)
			if i < 0 || err != nil {
				return nil, errors.New("invalid range")
			}
			if i > size {
				i = size
			}
			r.start = size - i
			r.length = size - r.start
		} else {
			// Format: M-N or M-
			i, err := strconv.ParseInt(start, 10, 64)
			if err != nil || i < 0 {
				return nil, errors.New("invalid range")
			}
			if i >= size {
				noOverlap = true
				continue
			}
			r.start = i
			if end == "" {
				r.length = size - r.start
			} else {
				i, err := strconv.ParseInt(end, 10, 64)
				if err != nil || r.start > i {
					return nil, errors.New("invalid range")
				}
				if i >= size {
					i = size - 1
				}
				r.length = i - r.start + 1
			}
		}
		ranges = append(ranges, r)
	}
	if noOverlap && len(ranges) == 0 {
		return nil, errNoOverlap
	}
	return ranges, nil
}

func serveContent(w http.ResponseWriter, req *http.Request, name string, modtime time.Time, sizeFunc func() (int64, error), content io.ReadSeeker) {
	setLastModified(w, modtime)
	done, rangeReq := checkPreconditions(w, req, modtime)
	if done {
		return
	}

	code := http.StatusOK
	ctypes, haveType := w.Header()["Content-Type"]
	var ctype string
	if !haveType {
		ctype = mime.TypeByExtension(filepath.Ext(name))
		if ctype == "" {
			var buf [sniffLen]byte
			n, _ := io.ReadFull(content, buf[:])
			ctype = http.DetectContentType(buf[:n])
			_, err := content.Seek(0, io.SeekStart)
			if err != nil {
				serveError(w, "seeker can't seek", http.StatusInternalServerError)
				return
			}
		}
		w.Header().Set("Content-Type", ctype)
	} else if len(ctypes) > 0 {
		ctype = ctypes[0]
	}

	size, err := sizeFunc()
	if err != nil {
		serveError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if size < 0 {
		serveError(w, "negative content size computed", http.StatusInternalServerError)
		return
	}

	sendSize := size
	ranges, err := parseRange(rangeReq, size)
	switch err {
	case nil:
		// Valid ranges
	case errNoOverlap:
		if size == 0 {
			ranges = nil
			break
		}
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", size))
		fallthrough
	default:
		serveError(w, err.Error(), http.StatusRequestedRangeNotSatisfiable)
		return
	}

	if sumRangesSize(ranges) > size {
		ranges = nil
	}

	// Set common headers
	w.Header().Set("Accept-Ranges", "bytes")

	if req.Method == "HEAD" {
		// For HEAD requests, we only need to set headers
		switch {
		case len(ranges) == 1:
			ra := ranges[0]
			w.Header().Set("Content-Range", ra.contentRange(size))
			w.Header().Set("Content-Length", strconv.FormatInt(ra.length, 10))
			w.WriteHeader(http.StatusPartialContent)
		case len(ranges) > 1:
			sendSize = rangesMIMESize(ranges, ctype, size)
			w.Header().Set("Content-Type", "multipart/byteranges; boundary="+multipart.NewWriter(io.Discard).Boundary())
			w.Header().Set("Content-Length", strconv.FormatInt(sendSize, 10))
			w.WriteHeader(http.StatusPartialContent)
		default:
			w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
			w.WriteHeader(code)
		}
		return
	}

	// For GET requests, we need to send content
	switch {
	case len(ranges) == 1:
		// Single range request
		ra := ranges[0]
		if _, err := content.Seek(ra.start, io.SeekStart); err != nil {
			serveError(w, err.Error(), http.StatusRequestedRangeNotSatisfiable)
			return
		}
		sendSize = ra.length
		code = http.StatusPartialContent
		w.Header().Set("Content-Range", ra.contentRange(size))
		w.Header().Set("Content-Length", strconv.FormatInt(sendSize, 10))
		w.WriteHeader(code)

		// Use buffered copy for better performance
		bufPtr := bufferPool.Get().(*[]byte)
		buf := *bufPtr
		defer bufferPool.Put(bufPtr)

		_, err = io.CopyBuffer(w, io.LimitReader(content, sendSize), buf)
		if err != nil {
			// Cannot write response after headers are sent
			return
		}

	case len(ranges) > 1:
		// Multiple range request
		sendSize = rangesMIMESize(ranges, ctype, size)
		code = http.StatusPartialContent

		pr, pw := io.Pipe()
		mw := multipart.NewWriter(pw)
		w.Header().Set("Content-Type", "multipart/byteranges; boundary="+mw.Boundary())
		w.Header().Set("Content-Length", strconv.FormatInt(sendSize, 10))
		w.WriteHeader(code)

		var wg sync.WaitGroup
		errc := make(chan error, 1)

		// Start a goroutine to write the response
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer pw.Close()

			for _, ra := range ranges {
				part, err := mw.CreatePart(ra.mimeHeader(ctype, size))
				if err != nil {
					errc <- err
					return
				}

				if _, err := content.Seek(ra.start, io.SeekStart); err != nil {
					errc <- err
					return
				}

				// Get buffer from pool
				bufPtr := bufferPool.Get().(*[]byte)
				buf := *bufPtr

				_, err = io.CopyBuffer(part, io.LimitReader(content, ra.length), buf)

				// Return buffer to pool
				bufferPool.Put(bufPtr)

				if err != nil {
					errc <- err
					return
				}
			}

			if err := mw.Close(); err != nil {
				errc <- err
			}
		}()

		// Copy from pipe to response writer with buffer
		bufPtr := bufferPool.Get().(*[]byte)
		buf := *bufPtr
		_, copyErr := io.CopyBuffer(w, pr, buf)
		bufferPool.Put(bufPtr)

		// Close the read side of the pipe
		pr.Close()

		// Wait for writer goroutine to finish
		wg.Wait()

		// Check for errors
		select {
		case err := <-errc:
			if err != nil && copyErr == nil {
				// Log the error (can't report to client at this point)
				return
			}
		default:
		}

	default:
		// No range or invalid range
		w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
		w.WriteHeader(code)

		// Use buffered copy for better performance
		bufPtr := bufferPool.Get().(*[]byte)
		buf := *bufPtr
		_, err = io.CopyBuffer(w, content, buf)
		bufferPool.Put(bufPtr)

		if err != nil {
			// Cannot write response after headers are sent
			return
		}
	}
}

func staticServeContent(w http.ResponseWriter, req *http.Request, name string, modtime time.Time, content io.ReadSeeker) {
	sizeFunc := func() (int64, error) {
		size, err := content.Seek(0, io.SeekEnd)
		if err != nil {
			return 0, errSeeker
		}
		_, err = content.Seek(0, io.SeekStart)
		if err != nil {
			return 0, errSeeker
		}
		return size, nil
	}
	serveContent(w, req, name, modtime, sizeFunc, content)
}

// Stream serves content to HTTP clients with optimized performance
func Stream(w http.ResponseWriter, req *http.Request, name string, modtime time.Time, content io.ReadSeeker) {
	staticServeContent(w, req, name, modtime, content)
}
