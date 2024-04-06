package main

import (
	"github.com/LLIEPJIOK/calculating-server/internal/controllers"
	"github.com/LLIEPJIOK/calculating-server/internal/database"
	"github.com/LLIEPJIOK/calculating-server/internal/workers"
	"github.com/gorilla/mux"
)

func init() {
	database.Configure()
	workers.Initialize()
}

func main() {
	defer database.Close()
	defer workers.CloseExpressionsChan()

	router := mux.NewRouter()
	controllers.ConfigureControllers(router)
}
