package main

import (
	"flag"
	entity "go-cinema/entities"
	"go-cinema/model"
	repo "go-cinema/repository"
	"go-cinema/theatre"

	"github.com/kashari/golog"
)

type functionsMap map[string]func()

func main() {
	golog.Init("/tmp/theatre.log")
	db, err := theatre.InitDB()
	if err != nil {
		golog.Error("Cannot connect to the database...")
		return
	}

	functions := functionsMap{
		"migrate": func() {
			golog.Info("Migrate is executing...")
			if err != nil {
				golog.Error("Cannot connect to the database due to: " + err.Error())
			}

			err = db.AutoMigrate(&model.User{}, &entity.Movie{}, &entity.Series{}, &entity.Episode{})
			if err != nil {
				golog.Error("Cannot map models to the database due to: " + err.Error())
				return
			}
		},
	}

	funcName := flag.String("f", "", "The name of the function to execute.")
	flag.Parse()

	function := functions[*funcName]

	if function != nil {
		golog.Info("Executing function: " + *funcName)
		function()
		golog.Info("Function executed successfully")
		return
	}

	repo.InitRepositories(db)

	router := theatre.SetupRoutes(db)

	port := "9090"
	golog.Info("Starting server on port: " + port)
	if err := router.Start(port); err != nil {
		golog.Error("Failed to start server: " + err.Error())
		return
	}

}
