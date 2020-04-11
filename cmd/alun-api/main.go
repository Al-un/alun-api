package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Al-un/alun-api/alun/core"
	"github.com/Al-un/alun-api/alun/memo"
	"github.com/Al-un/alun-api/alun/user"
	"github.com/Al-un/alun-api/alun/utils"
	"github.com/Al-un/alun-api/pkg/logger"
	"github.com/joho/godotenv"
)

var serverPort int

func main() {
	rootLogger := logger.NewConsoleLogger(logger.LogLevelInfo)

	// Env var loading
	err := godotenv.Load()
	if err != nil {
		// rootLogger.Fatal(1, "Error when load .env:\n%v", err)
	}

	// Server config
	serverPort, err = strconv.Atoi(os.Getenv(utils.EnvVarServerPort))
	if err != nil {
		rootLogger.Fatal(1, "Port %s is not defined", utils.EnvVarServerPort)
	}

	r := core.SetupRouter(
		core.APIMonolithic,
		user.UserAPI,
		memo.MemoAPI,
	)

	// Go!
	rootLogger.Info("[Server] Starting server on port %d...", serverPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", serverPort), r))
}
