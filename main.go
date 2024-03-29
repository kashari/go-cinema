package main

import (
	filehandler "dlna/io"
	"dlna/middlewares"
	"io"
	"net/http"
	"os"

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
