package theatre

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	BufferSize = 64 * 1024

	MaxCacheSize = 200 * 1024 * 1024

	MaxCachedChunkSize = 256 * 1024

	CacheTimeout = 10 * time.Minute

	PreloadAmount = 1024 * 1024
)

type ChunkCache struct {
	mu        sync.RWMutex
	cache     map[string]*CacheEntry
	size      int64
	maxSize   int64
	hitCount  int64
	missCount int64
}

type CacheEntry struct {
	data      []byte
	timestamp time.Time
	hits      int
}

var chunkCache = &ChunkCache{
	cache:   make(map[string]*CacheEntry),
	maxSize: MaxCacheSize,
}

type VideoServer struct {
	AllowedTypes  map[string]string
	BufferPool    *sync.Pool
	FileInfoCache sync.Map
	PreloadCache  sync.Map
}

type FileInfo struct {
	Size    int64
	ModTime time.Time
	expires time.Time
}

func NewVideoServer() *VideoServer {
	vs := &VideoServer{
		AllowedTypes: map[string]string{
			".mp4":  "video/mp4",
			".webm": "video/webm",
			".ogg":  "video/ogg",
			".mov":  "video/quicktime",
			".mkv":  "video/x-matroska",
			".avi":  "video/x-msvideo",
			".flv":  "video/x-flv",
		},
		BufferPool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, BufferSize)
			},
		},
	}

	go vs.periodicCacheCleanup()
	go vs.monitorPerformance()

	return vs
}

func (vs *VideoServer) periodicCacheCleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		vs.FileInfoCache.Range(func(key, value interface{}) bool {
			info := value.(FileInfo)
			if time.Now().After(info.expires) {
				vs.FileInfoCache.Delete(key)
			}
			return true
		})

		chunkCache.mu.Lock()
		now := time.Now()
		for key, entry := range chunkCache.cache {
			if now.Sub(entry.timestamp) > CacheTimeout {
				chunkCache.size -= int64(len(entry.data))
				delete(chunkCache.cache, key)
			}
		}
		chunkCache.mu.Unlock()

		runtime.GC()
	}
}

func (vs *VideoServer) monitorPerformance() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		chunkCache.mu.RLock()
		hits := chunkCache.hitCount
		misses := chunkCache.missCount
		total := hits + misses
		chunkCache.mu.RUnlock()

		if total > 0 {
			hitRatio := float64(hits) / float64(total) * 100
			log.Printf("Cache performance: hit ratio %.2f%% (%d hits, %d misses), cache size: %d bytes",
				hitRatio, hits, misses, chunkCache.size)
		}
	}
}

func (vs *VideoServer) GetFileInfo(path string) (FileInfo, error) {
	if cachedInfo, found := vs.FileInfoCache.Load(path); found {
		info := cachedInfo.(FileInfo)
		if time.Now().Before(info.expires) {
			return info, nil
		}
	}

	stat, err := os.Stat(path)
	if err != nil {
		return FileInfo{}, err
	}

	info := FileInfo{
		Size:    stat.Size(),
		ModTime: stat.ModTime(),
		expires: time.Now().Add(30 * time.Second),
	}

	vs.FileInfoCache.Store(path, info)
	return info, nil
}

func (vs *VideoServer) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		filePath := c.Query("file")
		if filePath == "" {
			c.String(http.StatusBadRequest, "Missing file parameter")
			return
		}

		if strings.Contains(filePath, "..") {
			c.String(http.StatusForbidden, "Invalid file path")
			return
		}

		fourLastCharactersInPath := filePath[len(filePath)-4:]

		isAvatar := fourLastCharactersInPath == "_mkv"
		var contentType string
		var allowed bool

		if !isAvatar {
			ext := strings.ToLower(filepath.Ext(filePath))
			log.Printf("Requested file: %s, extension: %s \n", filePath, ext)
			contentType, allowed = vs.AllowedTypes[ext]
			if !allowed {
				c.String(http.StatusForbidden, "File type not allowed")
				return
			}
		}

		fileInfo, err := vs.GetFileInfo(filePath)
		if err != nil {
			c.String(http.StatusNotFound, "File not found or inaccessible")
			return
		}

		rangeHeader := c.Request.Header.Get("Range")
		if rangeHeader == "" {
			vs.serveFullFile(c, filePath, fileInfo.Size, contentType)
			return
		}

		start, end, err := parseTheRange(rangeHeader, fileInfo.Size)
		if err != nil {
			c.String(http.StatusRequestedRangeNotSatisfiable, err.Error())
			return
		}

		if fileInfo.Size > PreloadAmount*2 && end < fileInfo.Size-PreloadAmount {
			go vs.preloadRange(filePath, end+1, end+PreloadAmount)
		}

		vs.serveRange(c, filePath, start, end, fileInfo.Size, contentType)
	}
}

func (vs *VideoServer) preloadRange(filePath string, start, end int64) {
	cacheKey := fmt.Sprintf("%s:%d-%d", filePath, start, end)

	chunkCache.mu.RLock()
	_, found := chunkCache.cache[cacheKey]
	chunkCache.mu.RUnlock()

	if found {
		return
	}

	if _, loaded := vs.PreloadCache.LoadOrStore(cacheKey, true); loaded {
		return
	}
	defer vs.PreloadCache.Delete(cacheKey)

	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return
	}

	contentLength := end - start + 1
	buffer := make([]byte, contentLength)
	n, err := io.ReadFull(file, buffer)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return
	}

	if int64(n) < contentLength {
		buffer = buffer[:n]
	}

	chunkCache.mu.Lock()
	defer chunkCache.mu.Unlock()

	for chunkCache.size+int64(len(buffer)) > chunkCache.maxSize {
		var leastUsedKey string
		var leastScore int = -1

		for k, v := range chunkCache.cache {
			age := time.Since(v.timestamp).Seconds()
			score := v.hits - int(age/60)

			if leastScore == -1 || score < leastScore {
				leastScore = score
				leastUsedKey = k
			}
		}

		if leastUsedKey == "" {
			break
		}

		chunkCache.size -= int64(len(chunkCache.cache[leastUsedKey].data))
		delete(chunkCache.cache, leastUsedKey)
	}

	chunkCache.cache[cacheKey] = &CacheEntry{
		data:      buffer,
		timestamp: time.Now(),
		hits:      0,
	}
	chunkCache.size += int64(len(buffer))
}

func (vs *VideoServer) serveFullFile(c *gin.Context, filePath string, fileSize int64, contentType string) {
	c.Header("Content-Type", contentType)
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Length", strconv.FormatInt(fileSize, 10))
	c.Header("Cache-Control", "public, max-age=86400")

	if c.Request.Method == "HEAD" {
		c.Status(http.StatusOK)
		return
	}

	if gin.Mode() != gin.DebugMode {
		file, err := os.Open(filePath)
		if err == nil {
			defer file.Close()

			fileInfo, err := file.Stat()
			if err == nil {
				http.ServeContent(c.Writer, c.Request, filepath.Base(filePath), fileInfo.ModTime(), file)
				return
			}
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error opening file")
		return
	}
	defer file.Close()

	buffer := vs.BufferPool.Get().([]byte)
	defer vs.BufferPool.Put(&buffer)

	c.Status(http.StatusOK)
	written, err := io.CopyBuffer(c.Writer, file, buffer)

	if err != nil && err != io.EOF {
		log.Printf("Error streaming file %s: %v (wrote %d bytes of %d)",
			filePath, err, written, fileSize)
	}
}

func (vs *VideoServer) serveRange(c *gin.Context, filePath string, start, end, fileSize int64, contentType string) {
	contentLength := end - start + 1

	c.Header("Content-Type", contentType)
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	c.Header("Content-Length", strconv.FormatInt(contentLength, 10))

	if c.Request.Method == "HEAD" {
		c.Status(http.StatusPartialContent)
		return
	}

	if contentLength <= MaxCachedChunkSize {
		cacheKey := fmt.Sprintf("%s:%d-%d", filePath, start, end)

		chunkCache.mu.RLock()
		entry, found := chunkCache.cache[cacheKey]
		if found {
			entry.hits++
			entry.timestamp = time.Now()
			chunkCache.hitCount++
			data := entry.data
			chunkCache.mu.RUnlock()

			c.Status(http.StatusPartialContent)
			c.Writer.Write(data)
			return
		}
		chunkCache.mu.RUnlock()
		chunkCache.mu.Lock()
		chunkCache.missCount++
		chunkCache.mu.Unlock()
	}

	file, err := os.Open(filePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error opening file")
		return
	}
	defer file.Close()

	if _, err := file.Seek(start, io.SeekStart); err != nil {
		c.String(http.StatusInternalServerError, "Error seeking in file")
		return
	}

	if contentLength <= MaxCachedChunkSize {
		buffer := make([]byte, contentLength)
		n, err := io.ReadFull(file, buffer)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			c.String(http.StatusInternalServerError, "Error reading file")
			return
		}

		if int64(n) < contentLength {
			buffer = buffer[:n]
		}

		cacheKey := fmt.Sprintf("%s:%d-%d", filePath, start, end)

		go func(key string, data []byte) {
			chunkCache.mu.Lock()
			defer chunkCache.mu.Unlock()

			for chunkCache.size+int64(len(data)) > chunkCache.maxSize && len(chunkCache.cache) > 0 {
				var oldestKey string
				var oldestTime time.Time
				first := true

				for k, v := range chunkCache.cache {
					if first || v.timestamp.Before(oldestTime) {
						oldestKey = k
						oldestTime = v.timestamp
						first = false
					}
				}

				if oldestKey != "" {
					chunkCache.size -= int64(len(chunkCache.cache[oldestKey].data))
					delete(chunkCache.cache, oldestKey)
				}
			}

			if chunkCache.size+int64(len(data)) <= chunkCache.maxSize {
				chunkCache.cache[key] = &CacheEntry{
					data:      data,
					timestamp: time.Now(),
					hits:      1,
				}
				chunkCache.size += int64(len(data))
			}
		}(cacheKey, buffer)

		c.Status(http.StatusPartialContent)
		c.Writer.Write(buffer)
		return
	}

	c.Status(http.StatusPartialContent)

	buffer := vs.BufferPool.Get().([]byte)
	defer vs.BufferPool.Put(&buffer)

	limitedReader := io.LimitReader(file, contentLength)

	written, err := io.CopyBuffer(c.Writer, limitedReader, buffer)

	if err != nil {
		log.Printf("Error streaming range %d-%d of %s: %v (wrote %d bytes of %d)",
			start, end, filePath, err, written, contentLength)
	}
}

func parseTheRange(rangeHeader string, fileSize int64) (int64, int64, error) {
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return 0, 0, errors.New("invalid range format")
	}

	rangeStr := strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return 0, 0, errors.New("invalid range format")
	}

	var start, end int64
	var err error

	if parts[0] == "" {
		suffixLength, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return 0, 0, errors.New("invalid range suffix")
		}

		start = fileSize - suffixLength
		if start < 0 {
			start = 0
		}
		end = fileSize - 1
	} else {
		start, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return 0, 0, errors.New("invalid range start")
		}

		if start >= fileSize {
			return 0, 0, errors.New("range start beyond file size")
		}

		if parts[1] == "" {
			end = fileSize - 1
		} else {
			end, err = strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return 0, 0, errors.New("invalid range end")
			}

			if end >= fileSize {
				end = fileSize - 1
			}
		}
	}

	if start > end {
		return 0, 0, errors.New("invalid range: start position after end")
	}

	return start, end, nil
}
