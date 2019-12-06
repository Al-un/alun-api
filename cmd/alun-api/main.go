package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Al-un/alun-api/pkg/user"
	"github.com/gorilla/mux"
)

var serverPort = 8000

func setupRouter() *mux.Router {
	router := mux.NewRouter()

	user.SetupRoutes(router)

	return router
}

func main() {
	r := setupRouter()

	log.Printf("[Server] Starting server on port %d...\n", serverPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", serverPort), r))
}
