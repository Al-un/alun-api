package core // import github.com/Al-un/alun-api/alun/core

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/Al-un/alun-api/alun/utils"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ----------------------------------------------------------------------------
//	>> MongoDB utilities
// ----------------------------------------------------------------------------

// ----- MongoDB exported variables -------------------------------------------

var (
	// MongoClient is an instance representing the connection to the database
	MongoClient *mongo.Client
	// MongoDatabase is an instance representing the single DB for this project
	MongoDatabase *mongo.Database
)

// ----- MongoDB local variables -----------------------------------------------

const dbDefaulMongoURI = "mongodb://localhost:27017/alun"

var (
	// dbConnectionString to locate the DB
	dbConnectionString string
	// dbName : database name for this project. All collections will be stored in the
	// same database for convenience
	dbName string
)

// ----------------------------------------------------------------------------
// Initialisation
// ----------------------------------------------------------------------------

// Init the connection with MongoDB upon app initialisation
func init() {
	if MongoDatabase == nil {
		initDao()
	}
}

// initDao sets up the connection with the database.
//
// This method can be called when initializing other package, such as the authentication
// information (user and stuff). Consequently, the dotenv reading must occur in this
// method and not in the `init()` of this package
func initDao() {
	// Variable initialisation
	// mongodb://heroku_rl0mksb2:ekaett1c181uem6kbph1tg53fo@ds241408.mlab.com:41408/heroku_rl0mksb2

	// Load dotenv
	err := godotenv.Load()
	if err != nil {
		coreLogger.Fatal(2, "Error when loading .env file:\n%v", err)
	}

	// Load URI
	mongoDbURI := os.Getenv(utils.EnvVarMongoDbURI)
	if mongoDbURI == "" {
		coreLogger.Debug("[MongoDB] Loading default MongoDB URI %s", dbDefaulMongoURI)
	}
	dbConnectionString, dbName = parseMongoDbURI(mongoDbURI)
	coreLogger.Debug("[MongoDB] Loading connection info [%v][%v]", dbConnectionString, dbName)

	// Client options
	clientOptions := options.Client().ApplyURI(dbConnectionString)

	// Try to connect to DB and init the MongoClient
	MongoClient, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		coreLogger.Fatal(2, "[MongoDB][ERROR] connection to %s failed:\n%v", dbConnectionString, err)
	}

	// Check the connection
	err = MongoClient.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("[MongoDB][ERROR] ping error: ", err)
	}
	coreLogger.Info("[MongoDB] Connected to database \\o/")

	// Init the database instance
	MongoDatabase = MongoClient.Database(dbName)
}

// parseMongoDbURI separate the DB. It is assumed that argument has the proper format such as
// mongodb://{user}:{passowrd}@{host}:{port}/{database}
//
// return (connectionString, databaseName)
func parseMongoDbURI(mongoDbURI string) (string, string) {
	splits := strings.Split(mongoDbURI, "/")

	var connString string
	dbName := splits[len(splits)-1]

	for i := 0; i < len(splits)-1; i++ {
		if connString == "" {
			connString = splits[i]
		} else {
			connString = connString + "/" + splits[i]
		}
	}

	return connString, dbName
}
