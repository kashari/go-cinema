package main

import (
	filehandler "dlna/io"
	"dlna/middlewares"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.DisableConsoleColor()
	r := gin.Default()
	r.Use(middlewares.CORSMiddleware())
	r.Use(middlewares.Logger())

	r.POST("/upload", uploadHandler)

	r.GET("/files", listHandler)

	r.GET("/files/:file", downloadHandler)

	r.DELETE("/files/:file", deleteHandler)

	r.GET("/download", downloadFromInternetHandler)

	r.GET("/percentage", percentageHandler)

	r.GET("/video/:file", videoServerHandler)

	r.Run()
}

func uploadHandler(c *gin.Context) {

	// Set a reasonable maximum file size for uploads
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<30) // 2GB

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "Error retrieving file: %s", err.Error())
		return
	}

	savePath := "/Media"

	destinationFile, err := os.Create(savePath + "/" + header.Filename)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error creating file: %s", err.Error())
		return
	}
	defer destinationFile.Close()

	if _, err := io.Copy(destinationFile, file); err != nil {
		c.String(http.StatusInternalServerError, "Error saving file: %s", err.Error())
		return
	}

	c.String(http.StatusOK, "File uploaded successfully to %s\n", savePath)
}

func listHandler(c *gin.Context) {
	fh := filehandler.FileHandler{Root: "/Media"}
	files := fh.ListFiles()
	c.JSON(http.StatusOK, gin.H{"files": files})
}

func deleteHandler(c *gin.Context) {
	fileName := c.Param("file")
	fh := filehandler.FileHandler{Root: "/Media"}
	err := fh.DeleteFile(fileName)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error deleting file: %s", err.Error())
		return
	}
	c.String(http.StatusOK, "File deleted successfully")
}

func downloadHandler(c *gin.Context) {
	fileName := c.Param("file")
	fh := filehandler.FileHandler{Root: "/Media"}
	file, err := fh.GetFile(fileName)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error opening file: %s", err.Error())
		return
	}
	defer file.Close()
	c.FileAttachment(fh.Root+"/"+fileName, fileName)
}

func downloadFromInternetHandler(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.String(http.StatusBadRequest, "Missing URL parameter")
		return
	}

	fh := filehandler.FileHandler{Root: "/Media"}
	err := fh.DownloadFromInternet(url)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error downloading file: %s", err.Error())
		return
	}
	c.String(http.StatusOK, "File downloaded successfully")
}

func percentageHandler(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.String(http.StatusBadRequest, "Missing URL parameter")
		return
	}
	filehandler := filehandler.FileHandler{Root: "/Media"}
	percentage := filehandler.PercentagePollerOnFile(url)
	c.JSON(http.StatusOK, gin.H{"percentage": percentage})
}

func videoServerHandler(c *gin.Context) {
    fileName := c.Param("file")
    fh := filehandler.FileHandler{Root: "/Media"}
    file, err := fh.ServeVideoFile(fileName)
    if err != nil {
        c.String(http.StatusInternalServerError, "Error opening file: %s", err.Error())
        return
    }
    defer file.Close()

    fileSize := filehandler.GetFileSize(file)
    handleRangeRequests(c.Writer, c.Request, file, fileSize) 
}

func handleRangeRequests(w http.ResponseWriter, r *http.Request, file *os.File, fileSize int64) {
	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
		fileInfo, err := file.Stat()
		if err != nil {
			log.Println("Error getting file info", err)
			return
		}
		http.ServeContent(w, r, file.Name(), fileInfo.ModTime(), file)
		return
	}

	fileInfo, err := file.Stat()
	if err != nil {
		log.Println("Error getting file info", err)
		return
	}

	http.ServeContent(w, r, file.Name(), fileInfo.ModTime(), file)
}

