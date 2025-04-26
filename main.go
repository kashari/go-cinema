package main

import (
	"flag"
	entity "go-cinema/entities"
	"go-cinema/model"
	repo "go-cinema/repository"
	"go-cinema/theatre"
	"log"
)

type functionsMap map[string]func()

func main() {
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

			err = db.AutoMigrate(&model.User{}, &entity.Movie{}, &entity.Series{}, &entity.Episode{})
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

	repo.InitRepositories(db)

	router := theatre.SetupRoutes(db)

	// Start the server on port 9090.
	port := "9090"
	log.Printf("Server starting on port %s", port)
	if err := router.Start(port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

}
