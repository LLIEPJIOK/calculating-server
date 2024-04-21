package main

import (
	"log"

	"github.com/LLIEPJIOK/calculating-server/internal/controllers"
	"github.com/LLIEPJIOK/calculating-server/internal/database"
	"github.com/LLIEPJIOK/calculating-server/internal/workers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}
	database.Configure()
	workers.Initialize()
}

func main() {
	defer database.Close()
	defer workers.CloseExpressionsChan()

	router := mux.NewRouter()
	controllers.ConfigureControllers(router)
}
