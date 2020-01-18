package core // import github.com/Al-un/alun-api/alun/core

import (
	"context"
	"log"
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ----------------------------------------------------------------------------
//	>> MongoDB utilities
// ----------------------------------------------------------------------------

// ----- MongoDB exported variables -------------------------------------------

// MongoClient is an instance representing the connection to the database
var MongoClient *mongo.Client

// MongoDatabase is an instance representing the single DB for this project
var MongoDatabase *mongo.Database

// ----- MongoDB local variables -----------------------------------------------

const dbDefaulMongoURI = "mongodb://localhost:27017/alun"

// dbConnectionString to locate the DB
var dbConnectionString string

// dbName : database name for this project. All collections will be stored in the
// same database for convenience
var dbName string

// ----------------------------------------------------------------------------
// Initialisation
// ----------------------------------------------------------------------------

// Init the connection with MongoDB upon app initialisation
func init() {
	if MongoDatabase == nil {
		initDao()
	}
}

func initDao() {
	// Variable initialisation
	// mongodb://heroku_rl0mksb2:ekaett1c181uem6kbph1tg53fo@ds241408.mlab.com:41408/heroku_rl0mksb2

	if mongoDbURI := os.Getenv("MONGODB_URI"); mongoDbURI != "" {
		// Loading for Heroku
		dbConnectionString, dbName = parseMongoDbURI(mongoDbURI)
		coreLogger.Debug("[MongoDB] Loading values from MONGODB_URI")
	} else {
		// Loading for local development
		dbConnectionString, dbName = parseMongoDbURI(dbDefaulMongoURI)
		coreLogger.Debug("[MongoDB] Loading default values [%v][%v]", dbConnectionString, dbName)
	}

	// Client options
	clientOptions := options.Client().ApplyURI(dbConnectionString)

	// Try to connect to DB and init the MongoClient
	MongoClient, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal("[MongoDB][ERROR] connection error: ", err)
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
