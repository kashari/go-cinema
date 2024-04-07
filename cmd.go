package main

// import (
// 	"flag"
// 	"fmt"
// 	"go-cinema/theatre"
// )

// type functionsMap map[string]func()

// func main() {
// 	functions := functionsMap{
// 		"migrate": func() {
// 			fmt.Println("Migrate is executing...")

// 			db, err := theatre.InitDB()
// 			if err != nil {
// 				fmt.Printf("Cannot connect to the database due to: %v", err)
// 			}

// 			// fmt.Println("Creating enum types...")
// 			// db.Exec("CREATE TYPE user_role AS ENUM ('admin', 'user');")

// 			err = db.AutoMigrate(&theatre.Movie{}, &theatre.Series{}, &theatre.Episode{})
// 			if err != nil {
// 				fmt.Printf("Cannot map models to the database due to: %v", err)
// 				return
// 			}
// 		},
// 	}

// 	funcName := flag.String("f", "", "The name of the function to execute.")
// 	flag.Parse()

// 	function := functions[*funcName]

// 	if function == nil {
// 		fmt.Printf("Function %v does not exist.", *funcName)
// 		return
// 	}

// 	// here the function gets executed if everything went well
// 	function()
// 	fmt.Println("Done.")
// }
