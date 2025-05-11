package main

import (
	"flag"
	entity "go-cinema/entities"
	"go-cinema/model"
	repo "go-cinema/repository"
	"go-cinema/theatre"
	"net/http"
	"time"

	"github.com/kashari/golog"
)

type functionsMap map[string]func()

func main() {
	golog.Init("/tmp/theatre.log")
	db, err := theatre.InitDB()
	if err != nil {
		golog.Error("Cannot connect to the database due to: {}", err.Error())
		return
	}

	functions := functionsMap{
		"migrate": func() {
			golog.Info("Running migration")

			err = db.AutoMigrate(&model.User{}, &entity.Movie{}, &entity.Series{}, &entity.Episode{})
			if err != nil {
				golog.Error("Failed to run migration: {}", err.Error())
				return
			}
		},
	}

	funcName := flag.String("f", "", "The name of the function to execute.")
	flag.Parse()

	function := functions[*funcName]

	if function != nil {
		golog.Info("Executing function: {}", *funcName)
		// here the function gets executed if everything went well
		function()
		golog.Info("Function executed successfully")
		return
	}

	repo.InitRepositories(db)

	router := theatre.SetupRoutes()

	logo := `
            __       .__  .__               .__                          
   ____    |__|____  |  | |  | _____ _______|  |__   ___________  ____   
  / ___\   |  \__  \ |  | |  | \__  \\_  __ \  |  \ /  _ \_  __ \/    \  
 / /_/  >  |  |/ __ \|  |_|  |__/ __ \|  | \/   Y  (  <_> )  | \/   |  \ 
 \___  /\__|  (____  /____/____(____  /__|  |___|  /\____/|__|  |___|  / 
/_____/\______|    \/               \/           \/                  \/  

			gjållårhðrñ - A simple HTTP router for Go
	`

	port := "9090"
	golog.Info(logo)
	golog.Info("Started server on port {}", port)

	server := &http.Server{
		Addr:         ":9090",
		Handler:      router, // Use the ServeMux
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		golog.Error("Failed to start server: {}", err.Error())
	}
}
