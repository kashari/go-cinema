package theatre

import (
	"go-cinema/cronos"
	"net/http"

	dbs "github.com/misenkashari/goutils/db"
	"gorm.io/gorm"
)

func InitDB() (db *gorm.DB, err error) {
	return dbs.Postgres().WithUser("misen").WithHost("192.168.3.200").WithDatabase("theatre").WithPassword("root").Open()
}

func CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if req.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, req)
	}
}

func SetupRoutes() http.Handler {
	router := NewRouter()

	router.POST("/movies/create", CreateMovieLegacy)
	router.POST("/movie_special", CreateMovieSpecial)
	router.PUT("/movies/:id/update", EditMovie)
	router.GET("/movies", GetMovies)
	router.GET("/movies/:id", GetMovie)
	router.DELETE("/movies/:id/delete", DeleteMovie)

	router.GET("/video", VideoServerHandler)
	router.POST("/last-access/:id", HandleLastAccessForMovie)
	router.GET("/left-at", GetUsageData)

	router.GET("/series", ListSeries)
	router.POST("/series/create", CreateSerie)
	router.GET("/series/:id", GetSerie)
	router.PUT("/series/:id/update", EditSerie)
	router.DELETE("/series/:id/delete", DeleteSerie)
	router.POST("/series/:id/append", AppendEpisodeToSeries)
	router.POST("/series/:id/special", AppendEpisodeToSeriesSpecial)
	router.GET("/series/:id/episodes", GetSerieEpisodes)

	router.POST("/episodes/:id/last-access", HandleLastAccessForEpisode)
	router.POST("/series/:id/current/set", HandleSetSeriesIndex)
	router.GET("/series/:id/current/get", HandleGetLastEpisodeIndex)

	router.POST("/start-cronos", cronos.StartCronos)
	router.POST("/stop-cronos", cronos.StopCronos)

	return router
}
