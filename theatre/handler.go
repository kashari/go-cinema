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

	"github.com/kashari/golog"
	"gorm.io/gorm"
)

var (
	usageData = "/home/mkashari/go-cinema/usage-data.io"
)

func stringToUint(str string) (uint, error) {
	id, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// CreateMovie handles movie file upload and saves it to a specified path
func CreateMovieLegacy(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request /movies/create handler, method: {}", r.Method)
	// Limit the size of the request body to 5GB (5 << 30 bytes)
	r.Body = http.MaxBytesReader(w, r.Body, 5<<30)

	// Parse the form to access file data and other fields
	err := r.ParseMultipartForm(10 << 20) // Limit 10MB for form data
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing form: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// Retrieve the file from the form
	file, header, err := r.FormFile("File")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving file: %s", err.Error()), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Set the path to save the file
	savePath := "/home/mkashari/UMS"
	destinationFile, err := os.Create(savePath + "/" + header.Filename)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating file: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	defer destinationFile.Close()

	// Copy the uploaded file content to the destination file
	if _, err := io.Copy(destinationFile, file); err != nil {
		http.Error(w, fmt.Sprintf("Error saving file: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	movie := entity.Movie{
		Title:       r.FormValue("Title"),
		Path:        savePath + "/" + header.Filename,
		Description: r.FormValue("Description"),
		ResumeAt:    "00:00",
	}

	err = repo.MovieRepository.Save(&movie)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating movie record: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Respond with the created movie in JSON format
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(movie)
}

func CreateMovie(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request /movies/create handler, method: {}", r.Method)

	if r.Method != http.MethodPost {
		golog.Error("Invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 5<<30) // 5GB
	file, header, err := r.FormFile("File")
	if err != nil {
		golog.Error("Error retrieving file: {}", err)
		http.Error(w, fmt.Sprintf("Error retrieving file: %s", err.Error()), http.StatusBadRequest)
		return
	}

	savePath := "/home/mkashari/UMS"

	destinationFile, err := os.Create(savePath + "/" + header.Filename)
	if err != nil {
		golog.Error("Error creating file: {}", err)
		http.Error(w, fmt.Sprintf("Error creating file: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	defer destinationFile.Close()

	if _, err := io.Copy(destinationFile, file); err != nil {
		golog.Error("Error saving file: {}", err)
		http.Error(w, fmt.Sprintf("Error saving file: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	movie := entity.Movie{
		Title:       r.FormValue("Title"),
		Path:        savePath + "/" + header.Filename,
		Description: r.FormValue("Description"),
		ResumeAt:    "00:00",
	}

	err = repo.MovieRepository.Save(&movie)
	if err != nil {
		golog.Error("Error creating movie record: {}", err)
		http.Error(w, fmt.Sprintf("Error creating movie record: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(movie)
}

func CreateMovieSpecial(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request /movie_special handler, method: {}", r.Method)
	if r.Method != http.MethodPost {
		golog.Error("Invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var movie entity.Movie
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&movie)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	movie.ResumeAt = "00:00"

	err = repo.MovieRepository.Save(&movie)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating movie record: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(movie)
}

func EditMovie(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request /movies/:id/update handler, method: {}", r.Method)

	if r.Method != http.MethodPut {
		golog.Error("Invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Get the ID parameter from the request context
	idParam := GetParam(r.Context(), "id")
	if idParam == "" {
		http.Error(w, "Movie ID is required", http.StatusBadRequest)
		return
	}

	id, err := stringToUint(idParam)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid movie ID: %s", err.Error()), http.StatusBadRequest)
		return
	}

	movie, err := repo.MovieRepository.FindByID(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Movie not found: %s", err.Error()), http.StatusNotFound)
		return
	}

	var movieReq entity.MovieRequest
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&movieReq)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
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
		golog.Error("Error updating movie record: {}", err)
		http.Error(w, fmt.Sprintf("Error updating movie record: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(movie)
}

func DeleteMovie(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request /movies/:id/delete handler, method: {}", r.Method)

	if r.Method != http.MethodDelete {
		golog.Error("Invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	idParam := GetParam(r.Context(), "id")
	if idParam == "" {
		http.Error(w, "Movie ID is required", http.StatusBadRequest)
		return
	}

	id, err := stringToUint(idParam)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid movie ID: %s", err.Error()), http.StatusBadRequest)
		return
	}

	movie, err := repo.MovieRepository.FindByID(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Movie not found: %s", err.Error()), http.StatusNotFound)
		return
	}

	err = repo.MovieRepository.DeleteByID(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting movie record: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	os.Remove(movie.Path)

	// return a string indicating success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode("Movie deleted successfully")
}

func GetMovies(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request /movies handler, method: {}", r.Method)
	if r.Method != http.MethodGet {
		golog.Error("Invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	movies, err := repo.MovieRepository.FindAll()
	if err != nil {
		golog.Error("Error retrieving movies: {}", err)
		http.Error(w, fmt.Sprintf("Error retrieving movies: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	moviesList := movies.ToSlice()
	if len(moviesList) == 0 {
		http.Error(w, "No movies found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(moviesList)
}

func GetMovie(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request /movies/:id handler, method: {}", r.Method)

	if r.Method != http.MethodGet {
		golog.Error("Invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	idParam := GetParam(r.Context(), "id")
	if idParam == "" {
		http.Error(w, "Movie ID is required", http.StatusBadRequest)
		return
	}

	id, err := stringToUint(idParam)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid movie ID: %s", err.Error()), http.StatusBadRequest)
		return
	}

	movie, err := repo.MovieRepository.FindByID(uint(id))
	if err != nil {
		golog.Error("Movie not found: {}", err)
		http.Error(w, fmt.Sprintf("Movie not found: %s", err.Error()), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(movie)
}

func VideoStreamer(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request /video handler, method: {}", r.Method)
	fileName := r.URL.Query().Get("file")

	err := videostream.StreamVideo(w, r, fileName)
	if err != nil {
		golog.Error("Error streaming video file: {}", err)
		http.Error(w, fmt.Sprintf("Error streaming video file: %s", err.Error()), http.StatusInternalServerError)
		return
	}
}

func VideoServerHandler(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request Method: {}", r.Method)
	golog.Info("Request URL: {}", r.URL.String())
	golog.Info("Request Headers: {}", r.Header)
	golog.Info("Query Parameters: {}", r.URL.Query())
	golog.Info("Request /video/legacy handler, method: {}", r.Method)

	if r.Method != http.MethodGet {
		golog.Error("Invalid request method: {}", r.Method)
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	fileName := r.URL.Query().Get("file")
	if fileName == "" {
		golog.Error("File parameter is missing")
		http.Error(w, "File parameter is missing", http.StatusBadRequest)
		return
	}

	golog.Info("Serving video file: {}", fileName)

	// Ensure ServeVideo doesn't return a nil file
	file, err := entity.ServeVideo(fileName)
	if err != nil {
		golog.Error("Error opening file: {}", err)
		http.Error(w, fmt.Sprintf("Error opening file: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if file == nil {
		golog.Error("File returned is nil")
		http.Error(w, "File not found or unable to serve video", http.StatusNotFound)
		return
	}

	// Ensure the file is properly closed after handling
	defer file.Close()

	fileSize := entity.GetFileSize(file)
	if fileSize == 0 {
		golog.Error("File size is 0")
		http.Error(w, "Error with file size", http.StatusInternalServerError)
		return
	}

	// Handle range requests if applicable
	handleRangeRequests(w, r, file, fileSize)
}

func handleRangeRequests(w http.ResponseWriter, r *http.Request, file *os.File, fileSize int64) {
	rangeHeader := r.Header.Get("Range")

	if rangeHeader == "" {
		w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
		fileInfo, err := file.Stat()
		if err != nil {
			golog.Error("Error getting file info: {}", err)
			http.Error(w, "Error getting file information", http.StatusInternalServerError)
			return
		}

		// Serve the full file if no range is specified
		http.ServeContent(w, r, file.Name(), fileInfo.ModTime(), file)
		return
	}

	// Handling byte-range requests for partial content
	fileInfo, err := file.Stat()
	if err != nil {
		golog.Error("Error getting file info: {}", err)
		http.Error(w, "Error getting file information", http.StatusInternalServerError)
		return
	}

	// Serve content for the range request
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

func HandleLastAccessForMovie(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request /last-access/:id handler, method: {}", r.Method)
	if r.Method != http.MethodPost {
		golog.Error("Invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	idParam := GetParam(r.Context(), "id")
	if idParam == "" {
		http.Error(w, "Movie ID is required", http.StatusBadRequest)
		return
	}

	id, err := stringToUint(idParam)
	if err != nil {
		golog.Error("Invalid movie ID: {}", err)
		http.Error(w, fmt.Sprintf("Invalid movie ID: %s", err.Error()), http.StatusBadRequest)
		return
	}

	movie, err := repo.MovieRepository.FindByID(uint(id))
	if err != nil {
		golog.Error("Movie not found: {}", err)
		http.Error(w, fmt.Sprintf("Movie not found: %s", err.Error()), http.StatusNotFound)
		return
	}

	movie.ResumeAt = r.URL.Query().Get("time")
	err = repo.MovieRepository.Save(movie)
	if err != nil {
		golog.Error("Error updating movie record: {}", err)
		http.Error(w, fmt.Sprintf("Error updating movie record: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	jsonMovie, _ := json.Marshal(movie)
	// remove the last } from the json object
	jsonMovie = jsonMovie[:len(jsonMovie)-1]
	jsonMovie = append(jsonMovie, []byte(`,"Type":"Movie"}`)...)
	UpdateUsageData(jsonMovie)
}

func GetUsageData(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request /left-at handler, method: {}", r.Method)
	if r.Method != http.MethodGet {
		golog.Error("Invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	file, err := os.Open(usageData)
	if err != nil {
		fmt.Println("Something went wrong reading the data: ", err)
	}

	defer file.Close()

	var item map[string]interface{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&item)
	if err != nil {
		fmt.Println("Something went wrong decoding the data: ", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(item)
}

func GetFile(fileName string) (*os.File, error) {
	golog.Info("Opening file: ", fileName)
	file, err := os.Open(fileName)
	if err != nil {
		golog.Error("Error opening file: ", err)
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

func CreateSerie(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request /series/create handler, method: {}", r.Method)
	if r.Method != http.MethodPost {
		golog.Error("Invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var serie entity.Series
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&serie)
	if err != nil {
		golog.Error("Invalid JSON format: {}", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	basedir := "/home/mkashari/UMS/Series/"

	err = os.MkdirAll(basedir+serie.Title, os.ModePerm)
	if err != nil {
		golog.Error("Error creating directory: {}", err)
		http.Error(w, fmt.Sprintf("Error creating directory: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	serie.BaseDir = basedir + serie.Title

	serie.CurrentIndex = 0

	err = repo.SeriesRepository.Save(&serie)
	if err != nil {
		golog.Error("Error creating serie record: {}", err)
		http.Error(w, fmt.Sprintf("Error creating serie record: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(serie)
}

func ListSeries(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request /series handler, method: {}", r.Method)
	series, err := repo.SeriesRepository.FindAll()
	if err != nil {
		golog.Error("Error retrieving series: {}", err)
		http.Error(w, fmt.Sprintf("Error retrieving series: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	seriesList := series.ToSlice()
	if len(seriesList) == 0 {
		golog.Error("No series found")
		http.Error(w, "No series found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(seriesList)
}

func GetSerie(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request /series/:id handler, method: {}", r.Method)
	if r.Method != http.MethodGet {
		golog.Error("Invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	idParam := GetParam(r.Context(), "id")
	if idParam == "" {
		http.Error(w, "Serie ID is required", http.StatusBadRequest)
		return
	}

	id, err := stringToUint(idParam)
	if err != nil {
		golog.Error("Invalid serie ID: {}", err)
		http.Error(w, fmt.Sprintf("Invalid serie ID: %s", err.Error()), http.StatusBadRequest)
		return
	}

	serie, err := repo.SeriesRepository.FindByID(uint(id))
	if err != nil {
		golog.Error("Serie not found: {}", err)
		http.Error(w, fmt.Sprintf("Serie not found: %s", err.Error()), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(serie)
}

func DeleteSerie(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request /series/:id/delete handler, method: {}", r.Method)
	if r.Method != http.MethodDelete {
		golog.Error("Invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	idParam := GetParam(r.Context(), "id")
	if idParam == "" {
		http.Error(w, "Serie ID is required", http.StatusBadRequest)
		return
	}

	id, err := stringToUint(idParam)
	if err != nil {
		golog.Error("Invalid serie ID: {}", err)
		http.Error(w, fmt.Sprintf("Invalid serie ID: %s", err.Error()), http.StatusBadRequest)
		return
	}

	serie, err := repo.SeriesRepository.FindByID(uint(id))
	if err != nil {
		golog.Error("Serie not found: {}", err)
		http.Error(w, fmt.Sprintf("Serie not found: %s", err.Error()), http.StatusNotFound)
		return
	}

	err = os.RemoveAll(serie.BaseDir)
	if err != nil {
		golog.Error("Error deleting directory: {}", err)
		http.Error(w, fmt.Sprintf("Error deleting directory: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	err = repo.SeriesRepository.DeleteByID(uint(id))
	if err != nil {
		golog.Error("Error deleting serie record: {}", err)
		http.Error(w, fmt.Sprintf("Error deleting serie record: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode("Serie deleted successfully")
}

func EditSerie(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request /series/:id/update handler, method: {}", r.Method)
	if r.Method != http.MethodPut {
		golog.Error("Invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	idParam := GetParam(r.Context(), "id")
	if idParam == "" {
		http.Error(w, "Serie ID is required", http.StatusBadRequest)
		return
	}

	id, err := stringToUint(idParam)
	if err != nil {
		golog.Error("Invalid serie ID: {}", err)
		http.Error(w, fmt.Sprintf("Invalid serie ID: %s", err.Error()), http.StatusBadRequest)
		return
	}

	serie, err := repo.SeriesRepository.FindByID(uint(id))
	if err != nil {
		golog.Error("Serie not found: {}", err)
		http.Error(w, fmt.Sprintf("Serie not found: %s", err.Error()), http.StatusNotFound)
		return
	}

	var serieReq entity.SeriesRequest

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&serieReq)
	if err != nil {
		golog.Error("Invalid JSON format: {}", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
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
		golog.Error("Error updating serie record: {}", err)
		http.Error(w, fmt.Sprintf("Error updating serie record: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(serie)
}

func AppendEpisodeToSeries(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request /series/:id/append handler, method: {}", r.Method)
	if r.Method != http.MethodPost {
		golog.Error("Invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	pattern := "/series/:id/append"
	path := r.URL.Path
	params, matched := matchPattern(pattern, path)
	if !matched {
		golog.Error("Not Found")
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	id, err := stringToUint(params["id"])
	if err != nil {
		golog.Error("Invalid serie ID: {}", err)
		http.Error(w, fmt.Sprintf("Invalid serie ID: %s", err.Error()), http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 20<<30) // 5GB
	file, header, err := r.FormFile("File")

	serie, err := repo.SeriesRepository.FindByID(uint(id))
	if err != nil {
		golog.Error("Serie not found: {}", err)
		http.Error(w, fmt.Sprintf("Serie not found: %s", err.Error()), http.StatusNotFound)
		return
	}

	// get all episodes by querying the database
	query := func(db *gorm.DB) *gorm.DB {
		return db.Where("series_id = ?", id)
	}

	episodes, err := repo.EpisodeRepository.FindByQuery(query)
	if err != nil {
		golog.Error("Error retrieving episodes: {}", err)
		http.Error(w, fmt.Sprintf("Error retrieving episodes: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	destinationFile, err := os.Create(serie.BaseDir + "/" + header.Filename)
	if err != nil {
		golog.Error("Error creating file: {}", err)
		http.Error(w, fmt.Sprintf("Error creating file: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	defer destinationFile.Close()

	if _, err := io.Copy(destinationFile, file); err != nil {
		golog.Error("Error saving file: {}", err)
		http.Error(w, fmt.Sprintf("Error saving file: %s", err.Error()), http.StatusInternalServerError)
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
		golog.Error("Error creating episode record: {}", err)
		http.Error(w, fmt.Sprintf("Error creating episode record: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(episode)
}

func AppendEpisodeToSeriesSpecial(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request /series/:id/special handler, method: {}", r.Method)
	if r.Method != http.MethodPost {
		golog.Error("Invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	pattern := "/series/:id/special"
	path := r.URL.Path
	params, matched := matchPattern(pattern, path)
	if !matched {
		golog.Error("Not Found")
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	id, err := stringToUint(params["id"])
	if err != nil {
		golog.Error("Invalid serie ID: {}", err)
		http.Error(w, fmt.Sprintf("Invalid serie ID: %s", err.Error()), http.StatusBadRequest)
		return
	}

	filename := r.URL.Query().Get("file")

	serie, err := repo.SeriesRepository.FindByID(uint(id))
	if err != nil {
		golog.Error("Serie not found: {}", err)
		http.Error(w, fmt.Sprintf("Serie not found: %s", err.Error()), http.StatusNotFound)
		return
	}

	// get all episodes by querying the database
	query := func(db *gorm.DB) *gorm.DB {
		return db.Where("series_id = ?", id)
	}

	episodes, err := repo.EpisodeRepository.FindByQuery(query)
	if err != nil {
		golog.Error("Error retrieving episodes: {}", err)
		http.Error(w, fmt.Sprintf("Error retrieving episodes: %s", err.Error()), http.StatusInternalServerError)
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
		golog.Error("Error creating episode record: {}", err)
		http.Error(w, fmt.Sprintf("Error creating episode record: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(episode)
}

func GetSerieEpisodes(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request /series/:id/episodes handler, method: {}", r.Method)
	if r.Method != http.MethodGet {
		golog.Error("Invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	pattern := "/series/:id/episodes"
	path := r.URL.Path
	params, matched := matchPattern(pattern, path)
	golog.Info("Matched: {}", matched)
	golog.Info("Params: {}", params)
	if !matched {
		golog.Error("Not Found")
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	id, err := stringToUint(params["id"])
	if err != nil {
		golog.Error("Invalid serie ID: {}", err)
		http.Error(w, fmt.Sprintf("Invalid serie ID: %s", err.Error()), http.StatusBadRequest)
		return
	}

	// get all episodes by querying the database
	query := func(db *gorm.DB) *gorm.DB {
		return db.Where("series_id = ?", id)
	}

	episodes, err := repo.EpisodeRepository.FindByQuery(query)
	if err != nil {
		golog.Error("Error retrieving episodes: {}", err)
		http.Error(w, fmt.Sprintf("Error retrieving episodes: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	episodesList := episodes.ToSlice()
	if len(episodesList) == 0 {
		golog.Error("No episodes found")
		http.Error(w, "No episodes found", http.StatusNotFound)
		return
	}

	MergeSortEpisodesByIndex(&episodesList)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(episodesList)
}

func HandleLastAccessForEpisode(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request /episodes/:id/last-access handler, method: {}", r.Method)
	if r.Method != http.MethodPost {
		golog.Error("Invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	pattern := "/episodes/:id/last-access"
	path := r.URL.Path
	params, matched := matchPattern(pattern, path)
	if !matched {
		golog.Error("Not Found")
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	id, err := stringToUint(params["id"])
	if err != nil {
		golog.Error("Invalid episode ID: {}", err)
		http.Error(w, fmt.Sprintf("Invalid episode ID: %s", err.Error()), http.StatusBadRequest)
		return
	}

	episode, err := repo.EpisodeRepository.FindByID(uint(id))
	if err != nil {
		golog.Error("Episode not found: {}", err)
		http.Error(w, fmt.Sprintf("Episode not found: %s", err.Error()), http.StatusNotFound)
		return
	}

	episode.ResumeAt = r.URL.Query().Get("time")
	err = repo.EpisodeRepository.Save(episode)
	if err != nil {
		golog.Error("Error updating episode record: {}", err)
		http.Error(w, fmt.Sprintf("Error updating episode record: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	serie, err := repo.SeriesRepository.FindByID(episode.SeriesID)
	if err != nil {
		golog.Error("Serie not found: {}", err)
		http.Error(w, fmt.Sprintf("Serie not found: %s", err.Error()), http.StatusNotFound)
		return
	}

	serie.Episodes = []entity.Episode{*episode}

	jsonSerie, _ := json.Marshal(serie)
	// append a type to the json object to differentiate between movie and serie
	jsonSerie = jsonSerie[:len(jsonSerie)-1]
	jsonSerie = append(jsonSerie, []byte(`,"Type":"Serie"}`)...)
	UpdateUsageData(jsonSerie)
}

func HandleSetSeriesIndex(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request /series/:id/current/set handler, method: {}", r.Method)
	if r.Method != http.MethodPost {
		golog.Error("Invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	pattern := "/series/:id/current/set"
	path := r.URL.Path
	params, matched := matchPattern(pattern, path)
	if !matched {
		golog.Error("Not Found")
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	id, err := stringToUint(params["id"])
	if err != nil {
		golog.Error("Invalid serie ID: {}", err)
		http.Error(w, fmt.Sprintf("Invalid serie ID: %s", err.Error()), http.StatusBadRequest)
		return
	}

	indexStr := r.URL.Query().Get("index")
	index, err := strconv.ParseUint(indexStr, 10, 64)
	if err != nil {
		golog.Error("Invalid index: {}", err)
		http.Error(w, fmt.Sprintf("Invalid index: %s", err.Error()), http.StatusBadRequest)
		return
	}

	serie, err := repo.SeriesRepository.FindByID(uint(id))
	if err != nil {
		golog.Error("Serie not found: {}", err)
		http.Error(w, fmt.Sprintf("Serie not found: %s", err.Error()), http.StatusNotFound)
		return
	}

	serie.CurrentIndex = uint(index)
	err = repo.SeriesRepository.Save(serie)
	if err != nil {
		golog.Error("Error updating serie record: {}", err)
		http.Error(w, fmt.Sprintf("Error updating serie record: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(serie)
}

func HandleGetLastEpisodeIndex(w http.ResponseWriter, r *http.Request) {
	golog.Info("Request /series/:id/current/get handler, method: {}", r.Method)
	if r.Method != http.MethodGet {
		golog.Error("Invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	pattern := "/series/:id/current/get"
	path := r.URL.Path
	params, matched := matchPattern(pattern, path)
	if !matched {
		golog.Error("Not Found")
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	id, err := stringToUint(params["id"])
	if err != nil {
		golog.Error("Invalid serie ID: {}", err)
		http.Error(w, fmt.Sprintf("Invalid serie ID: %s", err.Error()), http.StatusBadRequest)
		return
	}

	serie, err := repo.SeriesRepository.FindByID(uint(id))
	if err != nil {
		golog.Error("Serie not found: {}", err)
		http.Error(w, fmt.Sprintf("Serie not found: %s", err.Error()), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(serie.CurrentIndex)
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
