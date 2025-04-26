package filehandler

import (
	"encoding/json"
	"fmt"
	logger "go-cinema/file-logger"
	"io"
	"net/http"
	"os"
	"strings"

	"golang.org/x/sync/syncmap"
)

var (
	downloadProgress = syncmap.Map{}
	usageData        = "/application/logs/usage_data.io"
)

type FileHandler struct {
	// Path to the directory to upload files
	Root string
}

type FileRow struct {
	Name string
	Size string
}

func MergeSortByNames(files []FileRow) []FileRow {
	if len(files) <= 1 {
		return files
	}

	mid := len(files) / 2
	left := MergeSortByNames(files[:mid])
	right := MergeSortByNames(files[mid:])

	return MergeNames(left, right)
}

func MergeNames(left, right []FileRow) []FileRow {
	var result []FileRow
	l, r := 0, 0

	for l < len(left) && r < len(right) {
		if left[l].Name < right[r].Name {
			result = append(result, left[l])
			l++
		} else {
			result = append(result, right[r])
			r++
		}
	}

	result = append(result, left[l:]...)
	result = append(result, right[r:]...)

	return result
}

func (f *FileHandler) ListFiles() []FileRow {
	logger.Info("Listing files in directory", f.Root)
	dir, err := os.Open(f.Root)
	if err != nil {
		logger.Error("Error opening directory", err)
		return []FileRow{}
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		logger.Error("Error reading directory", err)
		return []FileRow{}
	}

	var fileNames []FileRow

	for _, file := range files {
		if file.Mode().IsRegular() {
			size := float64(file.Size()) * 0.000001
			sizeString := fmt.Sprintf("%.2f MB", size)
			fileNames = append(fileNames, FileRow{Name: file.Name(), Size: sizeString})
		}
	}

	return MergeSortByNames(fileNames)
}

func (f *FileHandler) DeleteFile(fileName string) error {
	logger.Info("Deleting file", fileName)
	err := os.Remove(f.Root + "/" + fileName)
	if err != nil {
		logger.Error("Error deleting file", err)
		return err
	}
	return nil
}

func (f *FileHandler) GetFile(fileName string) (*os.File, error) {
	logger.Info("Opening file", fileName)
	file, err := os.Open(f.Root + "/" + fileName)
	if err != nil {
		logger.Error("Error opening file", err)
		return nil, err
	}
	return file, nil
}

func (f *FileHandler) DownloadFromInternet(url string) error {
	logger.Info("Downloading file from", url)
	response, err := http.Get(url)
	if err != nil {
		logger.Error("Error downloading file", err)
		return err
	}
	defer response.Body.Close()

	fileName := strings.Split(url, "/")[len(strings.Split(url, "/"))-1]

	out, err := os.Create(f.Root + "/" + fileName)
	if err != nil {
		logger.Error("Error creating file", err)
		return err
	}
	defer out.Close()

	contentLength := response.ContentLength

	buffer := make([]byte, 1024)

	var downloadedBytes int64

	var percentage float64

	for {
		n, err := response.Body.Read(buffer)
		if err != nil && err != io.EOF {
			logger.Error("Error reading response body", err)
			return err
		}

		_, err = out.Write(buffer[:n])
		if err != nil {
			logger.Error("Error writing to file", err)
			return err
		}

		downloadedBytes += int64(n)

		percentage = float64(downloadedBytes) / float64(contentLength) * 100
		downloadProgress.Store(url, percentage)

		if err == io.EOF {
			break
		}
	}

	downloadProgress.Delete(url)

	logger.Info("File downloaded successfully")
	return nil
}

func UpdateUsageData(data []byte) {
	file, err := os.Create(usageData)
	if err != nil {
		fmt.Println("Something went wrong updating the data: ", err)
	}

	file.Write(data)
	fmt.Println("Last access watching updated.")
}

func GetUsageData() map[string]string {
	file, err := os.Open(usageData)
	if err != nil {
		fmt.Println("Something went wrong reading the data: ", err)
	}

	data := make(map[string]string)
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		fmt.Println("Something went wrong decoding the data: ", err)
	}

	return data
}

func (f *FileHandler) PercentagePollerOnFile(url string) int64 {
	value, _ := downloadProgress.Load(url)
	return int64(value.(float64))
}

func (f *FileHandler) ServeVideoFile(name string) (*os.File, error) {
	logger.Info("Serving video file", name)
	file, err := os.Open(f.Root + "/" + name)
	if err != nil {
		logger.Error("Error opening video file", err)
		return nil, err
	}
	return file, nil
}

func GetFileSize(file *os.File) int64 {
	info, err := file.Stat()
	if err != nil {
		return 0
	}
	return info.Size()
}

func (f *FileHandler) ServeFile(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("file")
	if fileName == "" {
		http.Error(w, "File name is required", http.StatusBadRequest)
		return
	}

	file, err := f.GetFile(fileName)
	if err != nil {
		http.Error(w, "Error opening file", http.StatusInternalServerError)
		return
	}

	defer file.Close()
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", GetFileSize(file)))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Range", fmt.Sprintf("bytes 0-%d/%d", GetFileSize(file)-1, GetFileSize(file)))
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Accept-Encoding", "identity")
	w.Header().Set("Content-Encoding", "identity")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	w.Header().Set("Feature-Policy", "geolocation 'none'; microphone 'none'; camera 'none'")
	w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Max-Age", "86400")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Error writing file to response", http.StatusInternalServerError)
		return
	}

	logger.Info("File served successfully")
	return
}

func (f *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error getting file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	out, err := os.Create(f.Root + "/" + r.FormValue("filename"))
	if err != nil {
		http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Error writing file to disk", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded successfully"))
	logger.Info("File uploaded successfully")
}

func (f *FileHandler) ServeContentViaHttp(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("file")
	if fileName == "" {
		http.Error(w, "File name is required", http.StatusBadRequest)
		return
	}

	file, err := f.GetFile(fileName)
	if err != nil {
		http.Error(w, "Error opening file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "inline; filename="+fileName)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", GetFileSize(file)))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Accept-Encoding", "identity")
	w.Header().Set("Content-Encoding", "identity")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	w.Header().Set("Feature-Policy", "geolocation 'none'; microphone 'none'; camera 'none'")
	w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Max-Age", "86400")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Range", fmt.Sprintf("bytes 0-%d/%d", GetFileSize(file)-1, GetFileSize(file)))
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	stat, err := os.Stat(file.Name())
	if err != nil {
		http.Error(w, "Error getting file info", http.StatusInternalServerError)
		return
	}

	mod_time := stat.ModTime()
	if mod_time.IsZero() {
		http.Error(w, "Error getting file info", http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, fileName, mod_time, file)

	logger.Info("File served successfully")
	return
}

type Dlna interface {
	ServeVideoFile(name string) (*os.File, error)
	ServeFile(w http.ResponseWriter, r *http.Request)
	UploadFile(w http.ResponseWriter, r *http.Request)
	ServeContentViaHttp(w http.ResponseWriter, r *http.Request)
	DownloadFromInternet(url string) error
	ListFiles() []FileRow
	DeleteFile(fileName string) error
	GetFile(fileName string) (*os.File, error)
	PercentagePollerOnFile(url string) int64
	UpdateUsageData(data []byte)
	GetUsageData() map[string]string
}

func NewFileHandler(root string) *FileHandler {
	return &FileHandler{
		Root: root,
	}
}

func (f *FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		f.ServeFile(w, r)
	case http.MethodPost:
		f.UploadFile(w, r)
	case http.MethodDelete:
		fileName := r.URL.Query().Get("file")
		if fileName == "" {
			http.Error(w, "File name is required", http.StatusBadRequest)
			return
		}
		err := f.DeleteFile(fileName)
		if err != nil {
			http.Error(w, "Error deleting file", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("File deleted successfully"))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
