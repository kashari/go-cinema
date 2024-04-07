package theatre

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	usageData = "/Users/klajdi/Desktop/usage_data.io"
)

func CreateMovie(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 5<<30) // 5GB
	file, header, err := c.Request.FormFile("File")
	if err != nil {
		c.String(http.StatusBadRequest, "Error retrieving file: %s", err.Error())
		return
	}
	savePath := "/Users/klajdi/Downloads/go_cinema"

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

	movie := Movie{
		Title:       c.Request.FormValue("Title"),
		Path:        savePath + "/" + header.Filename,
		Description: c.Request.FormValue("Description"),
		ResumeAt:    "00:00",
	}

	result := db.Create(&movie)
	if result.Error != nil {
		c.String(http.StatusInternalServerError, "Error creating movie record: %s", result.Error.Error())
		return
	}

	c.JSON(http.StatusCreated, movie)
}

func EditMovie(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")

	var movie Movie
	if err := db.First(&movie, id).Error; err != nil {
		c.String(http.StatusNotFound, "Movie not found: %s", err.Error())
		return
	}

	var movieReq MovieRequest

	if err := c.ShouldBindJSON(&movieReq); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if movieReq.Title != "" {
		movie.Title = movieReq.Title
	}

	if movieReq.Description != "" {
		movie.Description = movieReq.Description
	}

	result := db.Save(&movie)
	if result.Error != nil {
		c.String(http.StatusInternalServerError, "Error updating movie record: %s", result.Error.Error())
		return
	}
}

func DeleteMovie(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")

	var movie Movie
	if err := db.First(&movie, id).Error; err != nil {
		c.String(http.StatusNotFound, "Movie not found: %s", err.Error())
		return
	}

	result := db.Delete(&movie)
	if result.Error != nil {
		c.String(http.StatusInternalServerError, "Error deleting movie record: %s", result.Error.Error())
		return
	}
	os.Remove(movie.Path)
}

func GetMovies(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var movies []Movie
	db.Find(&movies)
	c.JSON(http.StatusOK, movies)
}

func GetMovie(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")

	var movie Movie
	if err := db.First(&movie, id).Error; err != nil {
		c.String(http.StatusNotFound, "Movie not found: %s", err.Error())
		return
	}
	c.JSON(http.StatusOK, movie)
}

func VideoServerHandler(c *gin.Context) {
	fileName := c.Query("file")
	log.Println("Serving video file", fileName)
	file, err := ServeVideo(fileName)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error opening file: %s", err.Error())
		return
	}
	defer file.Close()

	fileSize := GetFileSize(file)
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

func UpdateUsageData(data []byte) {
	file, err := os.Create(usageData)
	if err != nil {
		fmt.Println("Something went wrong updating the data: ", err)
	}

	file.Write(data)
	fmt.Println("Last access watching updated.")
}

func HandleLastAccessForMovie(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")

	var movie Movie
	if err := db.First(&movie, id).Error; err != nil {
		c.String(http.StatusNotFound, "Movie not found: %s", err.Error())
		return
	}

	movie.ResumeAt = c.Query("time")
	result := db.Save(&movie)
	if result.Error != nil {
		c.String(http.StatusInternalServerError, "Error updating movie record: %s", result.Error.Error())
		return
	}

	jsonMovie, _ := json.Marshal(movie)
	UpdateUsageData(jsonMovie)
}

func GetUsageData(c *gin.Context) {
	file, err := os.Open(usageData)
	if err != nil {
		fmt.Println("Something went wrong reading the data: ", err)
	}

	var movie Movie
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&movie)
	if err != nil {
		fmt.Println("Something went wrong decoding the data: ", err)
	}

	c.JSON(http.StatusOK, movie)
}

func GetFile(fileName string) (*os.File, error) {
	log.Println("Opening file", fileName)
	file, err := os.Open(fileName)
	if err != nil {
		log.Println("Error opening file", err)
		return nil, err
	}
	return file, nil
}

func Pop(words *[]string) string {
	f := len(*words)
	rv := (*words)[f-1]
	*words = (*words)[:f-1]
	return rv
}

func HandleDownloadFile(c *gin.Context) {
	filePath := c.Query("file")
	words := strings.Split(filePath, "/")
	fileName := Pop(&words)

	log.Println(fileName)
	file, err := GetFile(filePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error opening file: %s", err.Error())
		return
	}
	defer file.Close()
	c.FileAttachment(filePath, fileName)
}
