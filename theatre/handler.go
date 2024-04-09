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
	usageData = "/application/logs/usage_data.io"
)

func CreateMovie(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 5<<30) // 5GB
	file, header, err := c.Request.FormFile("File")
	if err != nil {
		c.String(http.StatusBadRequest, "Error retrieving file: %s", err.Error())
		return
	}
	savePath := "/Media/go_cinema/Movies"

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
	// remove the last } from the json object
	jsonMovie = jsonMovie[:len(jsonMovie)-1]
	jsonMovie = append(jsonMovie, []byte(`,"Type":"Movie"}`)...)
	UpdateUsageData(jsonMovie)
}

func GetUsageData(c *gin.Context) {
	file, err := os.Open(usageData)
	if err != nil {
		fmt.Println("Something went wrong reading the data: ", err)
	}

	var item map[string]interface{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&item)
	if err != nil {
		fmt.Println("Something went wrong decoding the data: ", err)
	}

	c.JSON(http.StatusOK, item)
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

func CreateSerie(c *gin.Context) {
	var serie Series
	db := c.MustGet("db").(*gorm.DB)
	if err := c.ShouldBindJSON(&serie); err != nil {
		fmt.Printf("Error binding json: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	basedir := "/Media/go_cinema/Series/"

	err := os.MkdirAll(basedir+serie.Title, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating directory: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	serie.BaseDir = basedir + serie.Title

	serie.CurrentIndex = 0

	result := db.Create(&serie)
	if result.Error != nil {
		c.String(http.StatusInternalServerError, "Error creating serie record: %s", result.Error.Error())
		return
	}
}

func ListSeries(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var series []Series
	db.Find(&series)
	c.JSON(http.StatusOK, series)
}

func GetSerie(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")

	var serie Series
	if err := db.First(&serie, id).Error; err != nil {
		c.String(http.StatusNotFound, "Serie not found: %s", err.Error())
		return
	}
	c.JSON(http.StatusOK, serie)
}

func DeleteSerie(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")

	var serie Series
	if err := db.First(&serie, id).Error; err != nil {
		c.String(http.StatusNotFound, "Serie not found: %s", err.Error())
		return
	}

	err := os.RemoveAll(serie.BaseDir)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error deleting directory: %s", err.Error())
		return
	}

	result := db.Delete(&serie)
	if result.Error != nil {
		c.String(http.StatusInternalServerError, "Error deleting serie record: %s", result.Error.Error())
		return
	}
}

func EditSerie(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")

	var serie Series
	if err := db.First(&serie, id).Error; err != nil {
		c.String(http.StatusNotFound, "Serie not found: %s", err.Error())
		return
	}

	var serieReq SeriesRequest

	if err := c.ShouldBindJSON(&serieReq); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if serieReq.Title != "" {
		serie.Title = serieReq.Title
	}

	if serieReq.Description != "" {
		serie.Description = serieReq.Description
	}

	result := db.Save(&serie)
	if result.Error != nil {
		c.String(http.StatusInternalServerError, "Error updating serie record: %s", result.Error.Error())
		return
	}
}

func AppendEpisodeToSeries(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 5<<30) // 5GB
	file, header, err := c.Request.FormFile("File")

	var serie Series
	if err := db.Preload("Episodes").First(&serie, id).Error; err != nil {
		c.String(http.StatusNotFound, "Serie not found: %s", err.Error())
		return
	}

	if err != nil {
		c.String(http.StatusBadRequest, "Error retrieving file: %s", err.Error())
		return
	}

	destinationFile, err := os.Create(serie.BaseDir + "/" + header.Filename)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error creating file: %s", err.Error())
		return
	}

	defer destinationFile.Close()

	if _, err := io.Copy(destinationFile, file); err != nil {
		c.String(http.StatusInternalServerError, "Error saving file: %s", err.Error())
		return
	}

	currentIndex := 0
	if serie.Episodes != nil {
		currentIndex = len(serie.Episodes)
	}

	episode := Episode{
		Path:         serie.BaseDir + "/" + header.Filename,
		ResumeAt:     "00:00",
		EpisodeIndex: currentIndex + 1,
		SeriesID:     serie.ID,
	}

	result := db.Create(&episode)
	if result.Error != nil {
		c.String(http.StatusInternalServerError, "Error creating episode record: %s", result.Error.Error())
		return
	}

	c.JSON(http.StatusCreated, episode)
}

func GetSerieEpisodes(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")

	var serie Series
	if err := db.First(&serie, id).Error; err != nil {
		c.String(http.StatusNotFound, "Serie not found: %s", err.Error())
		return
	}

	var episodes []Episode
	db.Model(&serie).Association("Episodes").Find(&episodes)

	MergeSortEpisodesByIndex(&episodes)

	c.JSON(http.StatusOK, episodes)
}

func HandleLastAccessForEpisode(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")

	var episode Episode
	if err := db.First(&episode, id).Error; err != nil {
		c.String(http.StatusNotFound, "Episode not found: %s", err.Error())
		return
	}

	episode.ResumeAt = c.Query("time")
	result := db.Save(&episode)
	if result.Error != nil {
		c.String(http.StatusInternalServerError, "Error updating episode record: %s", result.Error.Error())
		return
	}

	var serie Series
	if err := db.First(&serie, episode.SeriesID).Error; err != nil {
		c.String(http.StatusNotFound, "Serie not found: %s", err.Error())
		return
	}

	serie.Episodes = []Episode{episode}

	jsonSerie, _ := json.Marshal(serie)
	// append a type to the json object to differentiate between movie and serie
	jsonSerie = jsonSerie[:len(jsonSerie)-1]
	jsonSerie = append(jsonSerie, []byte(`,"Type":"Serie"}`)...)
	UpdateUsageData(jsonSerie)
}

func HandleSetSeriesIndex(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")
	indexStr := c.Query("index")
	index, err := strconv.ParseUint(indexStr, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid index value: %s", err.Error())
		return
	}

	var serie Series
	if err := db.First(&serie, id).Error; err != nil {
		c.String(http.StatusNotFound, "Serie not found: %s", err.Error())
		return
	}

	serie.CurrentIndex = uint(index)
	result := db.Save(&serie)
	if result.Error != nil {
		c.String(http.StatusInternalServerError, "Error updating serie record: %s", result.Error.Error())
		return
	}
}

func HandleGetLastEpisodeIndex(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")

	var serie Series
	if err := db.First(&serie, id).Error; err != nil {
		c.String(http.StatusNotFound, "Serie not found: %s", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"index": serie.CurrentIndex})
}

func MergeSortEpisodesByIndex(episodes *[]Episode) {
	if len(*episodes) <= 1 {
		return
	}

	mid := len(*episodes) / 2
	left := (*episodes)[:mid]
	right := (*episodes)[mid:]

	MergeSortEpisodesByIndex(&left)
	MergeSortEpisodesByIndex(&right)

	*episodes = MergeEpisodes(left, right)
}

func MergeEpisodes(left, right []Episode) []Episode {
	size, i, j := len(left)+len(right), 0, 0
	merged := make([]Episode, size)

	for k := 0; k < size; k++ {
		if i > len(left)-1 && j <= len(right)-1 {
			merged[k] = right[j]
			j++
			continue
		}

		if j > len(right)-1 && i <= len(left)-1 {
			merged[k] = left[i]
			i++
			continue
		}

		if left[i].EpisodeIndex < right[j].EpisodeIndex {
			merged[k] = left[i]
			i++
		} else {
			merged[k] = right[j]
			j++
		}
	}

	return merged
}
