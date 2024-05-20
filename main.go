package main

import (
	"flag"
	"go-cinema/model"
	"go-cinema/theatre"
	"log"

	"github.com/gin-gonic/gin"
)

type functionsMap map[string]func()

func main() {
	gin.DisableConsoleColor()

	db, err := theatre.InitDB()
	if err != nil {
		log.Fatal("Cannot connect to the database...")
		return
	}

	functions := functionsMap{
		"migrate": func() {
			log.Println("Migrate is executing...")
			if err != nil {
				log.Printf("Cannot connect to the database due to: %v", err)
			}

			err = db.AutoMigrate(&model.User{}, &theatre.Movie{}, &theatre.Series{}, &theatre.Episode{})
			if err != nil {
				log.Printf("Cannot map models to the database due to: %v", err)
				return
			}
		},
	}

	funcName := flag.String("f", "", "The name of the function to execute.")
	flag.Parse()

	function := functions[*funcName]

	if function != nil {
		log.Printf("Executing function: %v", *funcName)
		// here the function gets executed if everything went well
		function()
		log.Println("Done.")
		return
	}

	router := theatre.SetupRoutes(db)

	err = router.Run(":8080")
	if err != nil {
		log.Fatal("Cannot start gin web server....")
		return
	}

}
