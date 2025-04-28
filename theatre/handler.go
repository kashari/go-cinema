package theatre

import (
	"encoding/json"
	"fmt"
	entity "go-cinema/entities"
	repo "go-cinema/repository"
	videostream "go-cinema/video"
	"io"
	"net/http"
	"os"
	"strconv"

	gjallarhorn "github.com/kashari/gjallarhorn/engine"
	"github.com/kashari/golog"
	"gorm.io/gorm"
)

var (
	usageData = "/home/mkashari/go-cinema/usage-data.io"
)

func CreateMovie(c *gjallarhorn.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 5<<30) // 5GB
	file, header, err := c.Request.FormFile("File")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Error retrieving file: %s", err.Error()))
		return
	}
	savePath := "/home/mkashari/UMS"

	destinationFile, err := os.Create(savePath + "/" + header.Filename)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error creating file: %s", err.Error()))
		return
	}

	defer destinationFile.Close()

	if _, err := io.Copy(destinationFile, file); err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error saving file: %s", err.Error()))
		return
	}

	movie := entity.Movie{
		Title:       c.Request.FormValue("Title"),
		Path:        savePath + "/" + header.Filename,
		Description: c.Request.FormValue("Description"),
		ResumeAt:    "00:00",
	}

	err = repo.MovieRepository.Save(&movie)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error creating movie record: %s", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, movie)
}

func CreateMovieSpecial(c *gjallarhorn.Context) {
	var movie entity.Movie
	if err := c.BindJSON(&movie); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Error binding json: %s", err.Error()))
		return
	}

	movie.ResumeAt = "00:00"

	err := repo.MovieRepository.Save(&movie)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error creating movie record: %s", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, movie)
}

func EditMovie(c *gjallarhorn.Context) {
	id := c.Param("id")
	movie_id, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Invalid movie ID: %s", err.Error()))
		return
	}

	movie, err := repo.MovieRepository.FindByID(uint(movie_id))
	if err != nil {
		c.String(http.StatusNotFound, fmt.Sprintf("Movie not found: %s", err.Error()))
		return
	}
	var movieReq entity.MovieRequest

	if err := c.BindJSON(&movieReq); err != nil {
		c.JSON(400, fmt.Sprintf("Error binding json: %s", err.Error()))
		return
	}

	if movieReq.Title != "" {
		movie.Title = movieReq.Title
	}

	if movieReq.Description != "" {
		movie.Description = movieReq.Description
	}

	err = repo.MovieRepository.Save(movie)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error updating movie record: %s", err.Error()))
		return
	}

	c.JSON(http.StatusOK, movie)
}

func DeleteMovie(c *gjallarhorn.Context) {
	id, err := c.ParamInt("id")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Invalid movie ID: %s", err.Error()))
		return
	}

	movie, err := repo.MovieRepository.FindByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, fmt.Sprintf("Movie not found: %s", err.Error()))
		return
	}

	err = repo.MovieRepository.DeleteByID(uint(id))
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error deleting movie record: %s", err.Error()))
		return
	}

	os.Remove(movie.Path)

	c.String(http.StatusOK, "Movie deleted successfully")
}

func GetMovies(c *gjallarhorn.Context) {
	movies, err := repo.MovieRepository.FindAll()
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error retrieving movies: %s", err.Error()))
		return
	}
	c.JSON(http.StatusOK, movies.ToSlice())
}

func GetMovie(c *gjallarhorn.Context) {
	id, err := c.ParamInt("id")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Invalid movie ID: %s", err.Error()))
		return
	}

	movie, err := repo.MovieRepository.FindByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, fmt.Sprintf("Movie not found: %s", err.Error()))
		return
	}

	c.JSON(http.StatusOK, movie)
}

func VideoStreamer(c *gjallarhorn.Context) {
	fileName := c.QueryParam("file")
	golog.Info("Serving video file:", fileName)

	err := videostream.StreamVideo(c.Writer, c.Request, fileName)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error opening video file: %s", err.Error()))
		return
	}

}

func VideoServerHandler(c *gjallarhorn.Context) {
	fileName := c.QueryParam("file")
	golog.Info("Serving video file:", fileName)

	file, err := entity.ServeVideo(fileName)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error opening video file: %s", err.Error()))
		return
	}
	defer file.Close()

	fileSize := entity.GetFileSize(file)
	handleRangeRequests(c.Writer, c.Request, file, fileSize)
}

func handleRangeRequests(w http.ResponseWriter, r *http.Request, file *os.File, fileSize int64) {
	rangeHeader := r.Header.Get("Range")

	if rangeHeader == "" {
		w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
		fileInfo, err := file.Stat()
		if err != nil {
			golog.Error("Error getting file info", err)
			return
		}
		http.ServeContent(w, r, file.Name(), fileInfo.ModTime(), file)
		return
	}

	fileInfo, err := file.Stat()
	if err != nil {
		golog.Error("Error getting file info", err)
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

func HandleLastAccessForMovie(c *gjallarhorn.Context) {
	id, err := c.ParamInt("id")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Invalid movie ID: %s", err.Error()))
		return
	}

	movie, err := repo.MovieRepository.FindByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, fmt.Sprintf("Movie not found: %s", err.Error()))
		return
	}

	movie.ResumeAt = c.QueryParam("time")
	err = repo.MovieRepository.Save(movie)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error updating movie record: %s", err.Error()))
		return
	}

	jsonMovie, _ := json.Marshal(movie)
	// remove the last } from the json object
	jsonMovie = jsonMovie[:len(jsonMovie)-1]
	jsonMovie = append(jsonMovie, []byte(`,"Type":"Movie"}`)...)
	UpdateUsageData(jsonMovie)
}

func GetUsageData(c *gjallarhorn.Context) {
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
	golog.Info("Opening file", fileName)
	file, err := os.Open(fileName)
	if err != nil {
		golog.Error("Error opening file", err)
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

func CreateSerie(c *gjallarhorn.Context) {
	var serie entity.Series
	if err := c.BindJSON(&serie); err != nil {
		fmt.Printf("Error binding json: %s", err.Error())
		c.JSON(http.StatusBadRequest, fmt.Sprintf("Error binding json: %s", err.Error()))
		return
	}

	basedir := "/home/mkashari/UMS/Series/"

	err := os.MkdirAll(basedir+serie.Title, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating directory: %s", err.Error())
		c.JSON(http.StatusInternalServerError, fmt.Sprintf("Error creating directory: %s", err.Error()))
		return
	}

	serie.BaseDir = basedir + serie.Title

	serie.CurrentIndex = 0

	err = repo.SeriesRepository.Save(&serie)
	if err != nil {
		fmt.Printf("Error creating serie record: %s", err.Error())
		c.JSON(http.StatusInternalServerError, fmt.Sprintf("Error creating serie record: %s", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, serie)
}

func ListSeries(c *gjallarhorn.Context) {
	series, err := repo.SeriesRepository.FindAll()
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error retrieving series: %s", err.Error()))
		return
	}

	c.JSON(http.StatusOK, series.ToSlice())
}

func GetSerie(c *gjallarhorn.Context) {
	id, err := c.ParamInt("id")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Invalid serie ID: %s", err.Error()))
		return
	}

	serie, err := repo.SeriesRepository.FindByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, fmt.Sprintf("Serie not found: %s", err.Error()))
		return
	}

	c.JSON(http.StatusOK, serie)
}

func DeleteSerie(c *gjallarhorn.Context) {
	id, err := c.ParamInt("id")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Invalid serie ID: %s", err.Error()))
		return
	}

	serie, err := repo.SeriesRepository.FindByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, fmt.Sprintf("Serie not found: %s", err.Error()))
		return
	}

	err = os.RemoveAll(serie.BaseDir)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error deleting serie directory: %s", err.Error()))
		return
	}

	err = repo.SeriesRepository.DeleteByID(uint(id))
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error deleting serie record: %s", err.Error()))
		return
	}
	c.String(http.StatusOK, "Serie deleted successfully")
}

func EditSerie(c *gjallarhorn.Context) {
	id, err := c.ParamInt("id")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Invalid serie ID: %s", err.Error()))
		return
	}

	serie, err := repo.SeriesRepository.FindByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, fmt.Sprintf("Serie not found: %s", err.Error()))
		return
	}

	var serieReq entity.SeriesRequest

	if err := c.BindJSON(&serieReq); err != nil {
		c.JSON(400, fmt.Sprintf("Error binding json: %s", err.Error()))
		return
	}

	if serieReq.Title != "" {
		serie.Title = serieReq.Title
	}

	if serieReq.Description != "" {
		serie.Description = serieReq.Description
	}

	err = repo.SeriesRepository.Save(serie)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error updating serie record: %s", err.Error()))
		return
	}
	c.JSON(http.StatusOK, serie)
}

func AppendEpisodeToSeries(c *gjallarhorn.Context) {
	id, err := c.ParamInt("id")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Invalid serie ID: %s", err.Error()))
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 20<<30) // 5GB
	file, header, err := c.Request.FormFile("File")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Error retrieving file: %s", err.Error()))
		return
	}

	serie, err := repo.SeriesRepository.FindByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, fmt.Sprintf("Serie not found: %s", err.Error()))
		return
	}

	// get all episodes by querying the database
	query := func(db *gorm.DB) *gorm.DB {
		return db.Where("series_id = ?", id)
	}

	episodes, err := repo.EpisodeRepository.FindByQuery(query)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error retrieving episodes: %s", err.Error()))
		return
	}

	destinationFile, err := os.Create(serie.BaseDir + "/" + header.Filename)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error creating file: %s", err.Error()))
		return
	}

	defer destinationFile.Close()

	if _, err := io.Copy(destinationFile, file); err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error saving file: %s", err.Error()))
		return
	}

	currentIndex := episodes.Size()

	episode := entity.Episode{
		Path:         serie.BaseDir + "/" + header.Filename,
		ResumeAt:     "00:00",
		EpisodeIndex: currentIndex + 1,
		SeriesID:     serie.ID,
	}

	err = repo.EpisodeRepository.Save(&episode)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error creating episode record: %s", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, episode)
}

func AppendEpisodeToSeriesSpecial(c *gjallarhorn.Context) {
	id, err := c.ParamInt("id")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Invalid serie ID: %s", err.Error()))
		return
	}
	filename := c.QueryParam("filename")

	serie, err := repo.SeriesRepository.FindByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, fmt.Sprintf("Serie not found: %s", err.Error()))
		return
	}

	// get all episodes by querying the database
	query := func(db *gorm.DB) *gorm.DB {
		return db.Where("series_id = ?", id)
	}

	episodes, err := repo.EpisodeRepository.FindByQuery(query)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error retrieving episodes: %s", err.Error()))
		return
	}

	currentIndex := episodes.Size()

	episode := entity.Episode{
		Path:         serie.BaseDir + "/" + filename,
		ResumeAt:     "00:00",
		EpisodeIndex: currentIndex + 1,
		SeriesID:     serie.ID,
	}

	err = repo.EpisodeRepository.Save(&episode)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error creating episode record: %s", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, episode)
}

func GetSerieEpisodes(c *gjallarhorn.Context) {
	id, err := c.ParamInt("id")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Invalid serie ID: %s", err.Error()))
		return
	}

	// get all episodes by querying the database
	query := func(db *gorm.DB) *gorm.DB {
		return db.Where("series_id = ?", id)
	}

	episodes, err := repo.EpisodeRepository.FindByQuery(query)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error retrieving episodes: %s", err.Error()))
		return
	}

	episodesList := episodes.ToSlice()
	if len(episodesList) == 0 {
		c.String(http.StatusNotFound, "No episodes found for this series")
		return
	}

	MergeSortEpisodesByIndex(&episodesList)

	c.JSON(http.StatusOK, episodesList)
}

func HandleLastAccessForEpisode(c *gjallarhorn.Context) {
	id, err := c.ParamInt("id")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Invalid episode ID: %s", err.Error()))
		return
	}

	episode, err := repo.EpisodeRepository.FindByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, fmt.Sprintf("Episode not found: %s", err.Error()))
		return
	}

	episode.ResumeAt = c.QueryParam("time")
	err = repo.EpisodeRepository.Save(episode)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error updating episode record: %s", err.Error()))
		return
	}

	serie, err := repo.SeriesRepository.FindByID(episode.SeriesID)
	if err != nil {
		c.String(http.StatusNotFound, fmt.Sprintf("Serie not found: %s", err.Error()))
		return
	}

	serie.Episodes = []entity.Episode{*episode}

	jsonSerie, _ := json.Marshal(serie)
	// append a type to the json object to differentiate between movie and serie
	jsonSerie = jsonSerie[:len(jsonSerie)-1]
	jsonSerie = append(jsonSerie, []byte(`,"Type":"Serie"}`)...)
	UpdateUsageData(jsonSerie)
}

func HandleSetSeriesIndex(c *gjallarhorn.Context) {
	id, err := c.ParamInt("id")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Invalid serie ID: %s", err.Error()))
		return
	}
	indexStr := c.QueryParam("index")
	index, err := strconv.ParseUint(indexStr, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Invalid index: %s", err.Error()))
		return
	}

	serie, err := repo.SeriesRepository.FindByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, fmt.Sprintf("Serie not found: %s", err.Error()))
		return
	}

	serie.CurrentIndex = uint(index)
	err = repo.SeriesRepository.Save(serie)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error updating serie record: %s", err.Error()))
		return
	}
	c.JSON(http.StatusOK, serie)
}

func HandleGetLastEpisodeIndex(c *gjallarhorn.Context) {
	id, err := c.ParamInt("id")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Invalid serie ID: %s", err.Error()))
		return
	}

	serie, err := repo.SeriesRepository.FindByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, fmt.Sprintf("Serie not found: %s", err.Error()))
		return
	}

	c.JSON(http.StatusOK, serie.CurrentIndex)
}

func MergeSortEpisodesByIndex(episodes *[]entity.Episode) {
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

func MergeEpisodes(left, right []entity.Episode) []entity.Episode {
	size, i, j := len(left)+len(right), 0, 0
	merged := make([]entity.Episode, size)

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
