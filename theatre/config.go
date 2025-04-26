package theatre

import (
	"go-cinema/cronos"
	"net/http"
	"time"

	gjallarhorn "github.com/kashari/gjallarhorn/engine"
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

func SetupRoutes(db *gorm.DB) *gjallarhorn.Router {

	r := gjallarhorn.Heimdallr().WithFileLogging("/tmp/theatre-http.log").WithRateLimiter(100, 1*time.Second).WithWorkerPool(10)
	r.Use(CORSMiddleware)

	r.POST("/movies/create", CreateMovie)
	r.POST("/movie_special", CreateMovieSpecial)
	r.PUT("/movies/:id", EditMovie)
	r.GET("/movies", GetMovies)
	r.GET("/movies/:id", GetMovie)
	r.DELETE("/movies/:id", DeleteMovie)

	r.GET("/video", VideoStreamer)
	r.GET("/video/legacy", VideoServerHandler)
	r.POST("/last-access/:id", HandleLastAccessForMovie)
	r.GET("/left-at", GetUsageData)

	r.GET("/series", ListSeries)
	r.POST("/series/create", CreateSerie)
	r.GET("/series/:id", GetSerie)
	r.PUT("/series/:id", EditSerie)
	r.DELETE("/series/:id", DeleteSerie)
	r.POST("/series/:id/append", AppendEpisodeToSeries)
	r.POST("/series/:id/special", AppendEpisodeToSeriesSpecial)
	r.GET("/series/:id/episodes", GetSerieEpisodes)

	r.POST(("/episodes/:id/last-access"), HandleLastAccessForEpisode)
	r.POST("/series/:id/current", HandleSetSeriesIndex)
	r.GET("/series/:id/current", HandleGetLastEpisodeIndex)

	r.POST("/start-cronos", cronos.StartCronos)
	r.POST("/stop-cronos", cronos.StopCronos)

	return r
}
