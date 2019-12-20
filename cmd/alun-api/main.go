package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Al-un/alun-api/pkg/core"
	"github.com/Al-un/alun-api/pkg/logger"
)

var serverPort = 8000

func main() {
	r := core.SetupRouter(core.AuthAPI)

	rootLogger := logger.NewConsoleLogger(logger.LogLevelInfo)

	rootLogger.Info("[Server] Starting server on port %d...", serverPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", serverPort), r))
}
