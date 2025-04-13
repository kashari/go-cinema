package theatre

import (
	"go-cinema/extras"
	"go-cinema/handler"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() (db *gorm.DB, err error) {
	dsn := "host=192.168.3.200 user=misen password=root dbname=theatre port=5432 sslmode=disable TimeZone=Europe/Tirane"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, err
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token not provided"})
			c.Abort()
			return
		}

		if strings.Contains(token, "Bearer ") {
			token = strings.Split(token, "Bearer ")[1]
		}

		claims, err := extras.VerifyToken(token)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims["user_id"])
		c.Next()
	}
}

func SetupRoutes(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.Use(CORSMiddleware())

	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		log.Printf("Database connection injected for %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	})

	r.POST("/movies", CreateMovie)
	r.POST("/movie_special", CreateMovieSpecial)
	r.PUT("/movies/:id", EditMovie)
	r.GET("/movies", GetMovies)
	r.GET("/movies/:id", GetMovie)
	r.DELETE("/movies/:id", DeleteMovie)

	r.GET("/video", VideoServerHandler)
	r.GET("/video/download", HandleDownloadFile)
	r.POST("/last-access/:id", HandleLastAccessForMovie)
	r.GET("left-at", GetUsageData)

	r.GET("/series", ListSeries)
	r.POST("/series", CreateSerie)
	r.GET("/series/:id", GetSerie)
	r.PUT("/series/:id", EditSerie)
	r.DELETE("/series/:id", DeleteSerie)
	r.POST("series/:id/append", AppendEpisodeToSeries)
	r.POST("series/:id/special", AppendEpisodeToSeriesSpecial)
	r.GET("series/:id/episodes", GetSerieEpisodes)

	r.POST(("/episodes/:id/last-access"), HandleLastAccessForEpisode)
	r.POST("/series/:id/current", HandleSetSeriesIndex)
	r.GET("/series/:id/current", HandleGetLastEpisodeIndex)

	return r
}
