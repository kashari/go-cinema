package videostream

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultBufferSize = 4 * 1024 * 1024 // 4MB default buffer size
	maxAge            = 86400           // Cache control max age in seconds (24 hours)
)

var (
	// Error definitions
	ErrInvalidRange      = errors.New("invalid range")
	ErrUnsupportedFormat = errors.New("unsupported video format")
	ErrFileNotFound      = errors.New("file not found")
	ErrInvalidFilePath   = errors.New("invalid file path")

	// Buffer pool for reusing memory
	bufferPool sync.Pool

	// Supported video types
	supportedFormats = map[string]bool{
		".mp4":  true,
		".webm": true,
		".mov":  true,
		".avi":  true,
		".mkv":  true,
		".flv":  true,
	}

	// File cache to reduce disk I/O
	fileCache     = make(map[string]*fileCacheEntry)
	fileCacheMu   sync.RWMutex
	cleanupTicker *time.Ticker
)

type fileCacheEntry struct {
	info     os.FileInfo
	lastUsed time.Time
	hitCount int
}

// Config holds configurable options for the video streaming handler
type Config struct {
	BufferSize      int           // Buffer size for reading file chunks
	EnableFileCache bool          // Enable file info caching
	CacheTTL        time.Duration // Time to live for cached file info
	CleanupInterval time.Duration // Interval for cache cleanup
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		BufferSize:      defaultBufferSize,
		EnableFileCache: true,
		CacheTTL:        10 * time.Minute,
		CleanupInterval: 30 * time.Second,
	}
}

func init() {
	// Initialize buffer pool with the default buffer size
	bufferPool = sync.Pool{
		New: func() interface{} {
			buffer := make([]byte, defaultBufferSize)
			return &buffer
		},
	}

	// Start cleanup routine for file cache
	cleanupTicker = time.NewTicker(30 * time.Second)
	go cleanupCache()
}

// cleanupCache periodically removes old entries from the file cache
func cleanupCache() {
	for range cleanupTicker.C {
		now := time.Now()
		fileCacheMu.Lock()

		for path, entry := range fileCache {
			// Remove entries older than cache TTL
			log.Printf("Checking cache for %s, last used: %s\n", path, entry.lastUsed)
			if now.Sub(entry.lastUsed) > 10*time.Minute {
				delete(fileCache, path)
			}
		}

		fileCacheMu.Unlock()
	}
}

func NetHttp(w http.ResponseWriter, r *http.Request, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()

	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "video/mp4"
	}

	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Accept-Ranges", "bytes")
		http.ServeContent(w, r, filepath.Base(filePath), fileInfo.ModTime(), file)
		return nil
	}

	if !strings.HasPrefix(rangeHeader, "bytes=") {
		http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
		return fmt.Errorf("invalid range format")
	}

	rangeValue := strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.Split(rangeValue, "-")
	if len(parts) != 2 {
		http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
		return fmt.Errorf("invalid range format")
	}

	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || start < 0 {
		http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
		return fmt.Errorf("invalid start range")
	}

	var end int64
	if parts[1] == "" {
		end = fileSize - 1
	} else {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil || end >= fileSize {
			end = fileSize - 1
		}
	}

	if start > end || start >= fileSize {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
		http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
		return fmt.Errorf("invalid range values")
	}

	contentLength := end - start + 1

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	w.WriteHeader(http.StatusPartialContent)

	if r.Method == http.MethodHead {
		return nil
	}

	_, err = file.Seek(start, io.SeekStart)
	if err != nil {
		return err
	}
	http.ServeContent(w, r, filepath.Base(filePath), fileInfo.ModTime(), io.NewSectionReader(file, start, contentLength))

	return nil
}

func StreamVideo(w http.ResponseWriter, r *http.Request, filePath string) error {
	config := DefaultConfig()

	// Clean the path to prevent security issues
	filePath = filepath.Clean(filePath)

	// Check if file exists and get file info
	var fileInfo os.FileInfo
	var err error

	if config.EnableFileCache {
		fileInfo, err = getCachedFileInfo(filePath)
	} else {
		fileInfo, err = os.Stat(filePath)
	}

	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
			return ErrFileNotFound
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return err
		}
	}

	// Check if it's a directory
	if fileInfo.IsDir() {
		http.Error(w, "Cannot stream a directory", http.StatusBadRequest)
		return fmt.Errorf("path is a directory: %s", filePath)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if !supportedFormats[ext] {
		http.Error(w, "Unsupported video format", http.StatusUnsupportedMediaType)
		return ErrUnsupportedFormat
	}

	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return err
	}
	defer file.Close()

	// Get file size
	fileSize := fileInfo.Size()

	// Set content type based on file extension
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "video/mp4" // Default to mp4 if can't determine type
	}

	// Set common headers regardless of range request
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))

	// Check for range header
	rangeHeader := r.Header.Get("Range")

	// No range header means serve the whole file
	if rangeHeader == "" {
		w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))

		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return nil
		}

		// Stream the file in chunks instead of using ServeContent
		w.WriteHeader(http.StatusOK)
		return chunkedFileCopy(w, file, fileSize, config.BufferSize)
	}

	// Parse the range header
	ranges, err := parseRange(rangeHeader, fileSize)
	if err != nil {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
		http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
		return err
	}

	if len(ranges) > 1 {
		// We don't support multiple ranges
		http.Error(w, "Multiple ranges not supported", http.StatusNotImplemented)
		return fmt.Errorf("multiple ranges not supported")
	}

	// Get the range
	start, end := ranges[0][0], ranges[0][1]

	// Set headers for partial content
	w.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	w.WriteHeader(http.StatusPartialContent)

	// If it's a HEAD request, don't send the body
	if r.Method == http.MethodHead {
		return nil
	}

	// Seek to position
	_, err = file.Seek(start, io.SeekStart)
	if err != nil {
		return err
	}

	// Stream the range with timeout-resilient function
	return chunkedRangeFileCopy(w, file, start, end, config.BufferSize)
}

func parseRange(rangeHeader string, fileSize int64) ([][2]int64, error) {
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return nil, ErrInvalidRange
	}

	rangeHeader = strings.TrimPrefix(rangeHeader, "bytes=")
	ranges := strings.Split(rangeHeader, ",")
	parsedRanges := make([][2]int64, 0, len(ranges))

	for _, r := range ranges {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}

		parts := strings.Split(r, "-")
		if len(parts) != 2 {
			return nil, ErrInvalidRange
		}

		var start, end int64
		var err error

		if parts[0] == "" {
			end = fileSize - 1

			suffixLength, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return nil, ErrInvalidRange
			}

			start = fileSize - suffixLength
			if start < 0 {
				start = 0
			}
		} else {
			start, err = strconv.ParseInt(parts[0], 10, 64)
			if err != nil || start < 0 {
				return nil, ErrInvalidRange
			}

			if parts[1] == "" {
				end = fileSize - 1
			} else {
				end, err = strconv.ParseInt(parts[1], 10, 64)
				if err != nil || end < start || end >= fileSize {
					return nil, ErrInvalidRange
				}
			}
		}

		if start >= fileSize {
			return nil, ErrInvalidRange
		}

		if end >= fileSize {
			end = fileSize - 1
		}

		parsedRanges = append(parsedRanges, [2]int64{start, end})
	}

	if len(parsedRanges) == 0 {
		return nil, ErrInvalidRange
	}

	return parsedRanges, nil
}

// chunkedFileCopy streams an entire file in chunks to be more resilient to timeouts
func chunkedFileCopy(w http.ResponseWriter, file *os.File, fileSize int64, bufferSize int) error {
	// Get buffer from pool
	bufferPtr := bufferPool.Get().(*[]byte)
	buffer := *bufferPtr
	defer bufferPool.Put(bufferPtr)

	// Use a smaller buffer size if needed
	if bufferSize > defaultBufferSize {
		buffer = make([]byte, bufferSize)
	}

	// Track progress
	var totalWritten int64 = 0

	// Use flush writer for better network performance
	fw := bufio.NewWriterSize(w, bufferSize)
	defer fw.Flush()

	for totalWritten < fileSize {
		// Read from file
		n, readErr := file.Read(buffer)

		if n > 0 {
			// Write to response
			written, writeErr := fw.Write(buffer[:n])
			if writeErr != nil {
				return fmt.Errorf("write error: %w", writeErr)
			}

			totalWritten += int64(written)

			// Flush periodically to prevent buffer overflow
			if totalWritten%(int64(bufferSize)*4) == 0 {
				if err := fw.Flush(); err != nil {
					return fmt.Errorf("flush error: %w", err)
				}
			}
		}

		// Handle end of file
		if readErr == io.EOF {
			break
		}

		// Handle other errors
		if readErr != nil {
			return fmt.Errorf("read error: %w", readErr)
		}
	}

	// Final flush
	if err := fw.Flush(); err != nil {
		return fmt.Errorf("final flush error: %w", err)
	}

	return nil
}

// chunkedRangeFileCopy streams a specific range in chunks to be more resilient to timeouts
func chunkedRangeFileCopy(w http.ResponseWriter, file *os.File, start, end int64, bufferSize int) error {
	// Get buffer from pool
	bufferPtr := bufferPool.Get().(*[]byte)
	buffer := *bufferPtr
	defer bufferPool.Put(bufferPtr)

	// Track remaining bytes to write
	remaining := end - start + 1

	// Use flush writer for better network performance
	fw := bufio.NewWriterSize(w, bufferSize)
	defer fw.Flush()

	// Track progress for logging
	var totalWritten int64 = 0
	expectedBytes := remaining
	lastLogTime := time.Now()

	for remaining > 0 {
		// Adjust read size for last chunk
		readSize := min(remaining, int64(len(buffer)))

		// Read from file
		n, readErr := io.ReadAtLeast(file, buffer[:readSize], int(readSize))

		if n > 0 {
			// Write to response
			written, writeErr := fw.Write(buffer[:n])
			if writeErr != nil {
				log.Printf("Error streaming content: bytes written %d, expected %d, error: %v",
					totalWritten+int64(written), expectedBytes, writeErr)
				return writeErr
			}

			remaining -= int64(written)
			totalWritten += int64(written)

			if time.Since(lastLogTime) > 5*time.Second {
				log.Printf("Streaming progress: %.2f%% (%d/%d bytes)",
					float64(totalWritten)/float64(expectedBytes)*100, totalWritten, expectedBytes)
				lastLogTime = time.Now()
			}

			if totalWritten%(int64(bufferSize)*4) == 0 {
				if err := fw.Flush(); err != nil {
					return err
				}
			}
		}

		if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
			return readErr
		}

		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF || n == 0 {
			break
		}
	}

	if err := fw.Flush(); err != nil {
		return err
	}

	if totalWritten != expectedBytes {
		log.Printf("Warning: Content length mismatch. Wrote %d bytes, expected %d",
			totalWritten, expectedBytes)
	}

	return nil
}

// getCachedFileInfo gets file info with caching
func getCachedFileInfo(path string) (os.FileInfo, error) {
	// Check cache first
	fileCacheMu.RLock()
	entry, exists := fileCache[path]
	fileCacheMu.RUnlock()

	now := time.Now()

	if exists {
		// Update cache metrics
		fileCacheMu.Lock()
		entry.lastUsed = now
		entry.hitCount++
		fileCacheMu.Unlock()
		return entry.info, nil
	}

	// Not in cache, get from disk
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// Add to cache
	fileCacheMu.Lock()
	fileCache[path] = &fileCacheEntry{
		info:     info,
		lastUsed: now,
		hitCount: 1,
	}
	fileCacheMu.Unlock()

	return info, nil
}

// StopCleanup stops the background cleanup goroutine
func StopCleanup() {
	if cleanupTicker != nil {
		cleanupTicker.Stop()
	}
}
