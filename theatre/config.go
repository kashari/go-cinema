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
	dsn := "host=192.168.3.150 user=misen password=root dbname=theatre port=5432 sslmode=disable TimeZone=Europe/Tirane"
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

	r.POST("/login", handler.Login)
	r.POST("/refresh-token", handler.RefreshToken)
	r.POST("/register", handler.CreateUser)
	r.GET("/video", VideoServerHandler)

	auth := r.Group("/")
	auth.Use(AuthMiddleware())
	{
		auth.POST("/movies", CreateMovie)
		auth.PUT("/movies/:id", EditMovie)
		auth.GET("/movies", GetMovies)
		auth.GET("/movies/:id", GetMovie)
		auth.DELETE("/movies/:id", DeleteMovie)

		auth.GET("/video/download", HandleDownloadFile)
		auth.POST("/last-access/:id", HandleLastAccessForMovie)
		auth.GET("/left-at", GetUsageData)

		auth.GET("/series", ListSeries)
		auth.POST("/series", CreateSerie)
		auth.GET("/series/:id", GetSerie)
		auth.PUT("/series/:id", EditSerie)
		auth.DELETE("/series/:id", DeleteSerie)
		auth.POST("/series/:id/append", AppendEpisodeToSeries)
		auth.GET("/series/:id/episodes", GetSerieEpisodes)

		auth.POST("/episodes/:id/last-access", HandleLastAccessForEpisode)
		auth.POST("/series/:id/current", HandleSetSeriesIndex)
		auth.GET("/series/:id/current", HandleGetLastEpisodeIndex)
	}

	return r
}
