package main

import (
	"flag"
	entity "go-cinema/entities"
	logger "go-cinema/file-logger"
	"go-cinema/model"
	repo "go-cinema/repository"
	"go-cinema/theatre"
)

type functionsMap map[string]func()

func main() {
	logger.Setup("/tmp/theatre.log")
	db, err := theatre.InitDB()
	if err != nil {
		logger.Error("Cannot connect to the database...")
		return
	}

	functions := functionsMap{
		"migrate": func() {
			logger.Info("Migrate is executing...")
			if err != nil {
				logger.Error("Cannot connect to the database due to: " + err.Error())
			}

			err = db.AutoMigrate(&model.User{}, &entity.Movie{}, &entity.Series{}, &entity.Episode{})
			if err != nil {
				logger.Error("Cannot map models to the database due to: " + err.Error())
				return
			}
		},
	}

	funcName := flag.String("f", "", "The name of the function to execute.")
	flag.Parse()

	function := functions[*funcName]

	if function != nil {
		logger.Info("Executing function: " + *funcName)
		// here the function gets executed if everything went well
		function()
		logger.Info("Function executed successfully")
		return
	}

	repo.InitRepositories(db)

	router := theatre.SetupRoutes(db)

	// Start the server on port 9090.
	port := "9090"
	logger.Info("Starting server on port: " + port)
	if err := router.Start(port); err != nil {
		logger.Error("Failed to start server: " + err.Error())
		return
	}

}
