package filehandler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	log.Println("Listing files in directory", f.Root)
	dir, err := os.Open(f.Root)
	if err != nil {
		log.Println("Error opening directory", err)
		return []FileRow{}
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		log.Println("Error reading directory", err)
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
	log.Println("Deleting file", fileName)
	err := os.Remove(f.Root + "/" + fileName)
	if err != nil {
		log.Println("Error deleting file", err)
		return err
	}
	return nil
}

func (f *FileHandler) GetFile(fileName string) (*os.File, error) {
	log.Println("Opening file", fileName)
	file, err := os.Open(f.Root + "/" + fileName)
	if err != nil {
		log.Println("Error opening file", err)
		return nil, err
	}
	return file, nil
}

func (f *FileHandler) DownloadFromInternet(url string) error {
	log.Println("Downloading file from", url)
	response, err := http.Get(url)
	if err != nil {
		log.Println("Error downloading file", err)
		return err
	}
	defer response.Body.Close()

	fileName := strings.Split(url, "/")[len(strings.Split(url, "/"))-1]

	out, err := os.Create(f.Root + "/" + fileName)
	if err != nil {
		log.Println("Error creating file", err)
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
			log.Println("Error reading response body", err)
			return err
		}

		_, err = out.Write(buffer[:n])
		if err != nil {
			log.Println("Error writing to file", err)
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

	log.Println("File downloaded successfully")
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
	log.Println("Serving video file", name)
	file, err := os.Open(f.Root + "/" + name)
	if err != nil {
		log.Println("Error opening video file", err)
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
