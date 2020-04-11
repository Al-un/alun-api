package core // import github.com/Al-un/alun-api/alun/core

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ----------------------------------------------------------------------------
//	>> MongoDB utilities
// ----------------------------------------------------------------------------

// MongoConnectFromEnvVar connects to a Mongo database with the provided environment
// variable
func MongoConnectFromEnvVar(envVarName string) (*mongo.Client, *mongo.Database, error) {
	// Ensure that dotenv file is/are loaded
	err := godotenv.Load()
	if err != nil {
		coreLogger.Warn("Error when loading .env file:\n%v", err)
		return nil, nil, err
	}

	// Load URI
	mongoDbURI := os.Getenv(envVarName)
	if mongoDbURI == "" {
		return nil, nil, errors.New(fmt.Sprintf("MongoDB URL not found for variable: %s", envVarName))
	}

	return MongoConnectToDb(mongoDbURI)
}

// MongoConnectToDb creates a Mongo client instance from an URI as well as the
// Mongo database instance depending on the database name in the URI
//
//
// When initializing database instances, beware of package loading order and dotenv loading, such
// as the authentication information (user and stuff).
func MongoConnectToDb(mongoDbURI string) (*mongo.Client, *mongo.Database, error) {
	dbConnectionString, dbName := MongoParseDbURI(mongoDbURI)
	coreLogger.Debug("[MongoDB] Loading connection info [%v][%v]", dbConnectionString, dbName)

	// Client options
	clientOptions := options.Client().ApplyURI(dbConnectionString)

	// Try to connect to DB and init the MongoClient
	mongoClient, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		coreLogger.Fatal(2, "[MongoDB][ERROR] connection to %s failed:\n%v", dbConnectionString, err)
		return nil, nil, err
	}

	// Check the connection
	err = mongoClient.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("[MongoDB][ERROR] ping error: ", err)
		return nil, nil, err
	}
	coreLogger.Verbose("[MongoDB] Connected to database \\o/")

	// Init the database instance
	mongoDatabase := mongoClient.Database(dbName)

	return mongoClient, mongoDatabase, nil
}

// MongoParseDbURI parse the DB URL. It is assumed that argument has the proper format such as
// mongodb://{user}:{passowrd}@{host}:{port}/{database}
//
// return (connectionString, databaseName)
func MongoParseDbURI(mongoDbURI string) (string, string) {
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
