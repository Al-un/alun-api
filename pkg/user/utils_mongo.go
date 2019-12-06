package user

import (
	"context"
	"log"

	"github.com/Al-un/alun-api/pkg/core"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ----------------------------------------------------------------------------
//	>> Database utilities
// ----------------------------------------------------------------------------

// ---------- Variable and init -----------------------------------------------

// dbUserCollectionName : user collection name
var dbUserCollectionName string

// dbUserCollection : user collection instance
var dbUserCollection *mongo.Collection

// Init the connection with MongoDB upon app initialisation
func init() {
	// Initialisation: collections name
	dbUserCollectionName = "al_users"

	// Initialisation: collections instances
	dbUserCollection = core.MongoDatabase.Collection(dbUserCollectionName)

	log.Println("[MongoDB] User initialisation!")
}

// ---------- CRUD ------------------------------------------------------------

// findUser fetches an user for a given username and CLEAR password
func findUserByUsernamePassword(username string, clearPassword string) (User, error) {
	var user User
	var hashedPassword = hashPassword(clearPassword)

	filter := bson.M{"username": username, "password": hashedPassword}
	if err := dbUserCollection.FindOne(context.Background(), filter).Decode(&user); err != nil {
		return User{}, err
	}

	return user, nil
}

// findUserById fetches an user for a given ID
func findUserByID(userID string) (User, error) {
	var user User
	id, _ := primitive.ObjectIDFromHex(userID)

	filter := bson.M{"_id": id}
	if err := dbUserCollection.FindOne(context.Background(), filter).Decode(&user); err != nil {
		return User{}, err
	}

	return user, nil
}

// createUser creates the user assuming that passowrd is not hashed yet
func createUser(user User) *mongo.InsertOneResult {
	var hashedPassword = hashPassword(user.Password)
	user.Password = hashedPassword

	createdUser, _ := dbUserCollection.InsertOne(context.Background(), user)
	log.Printf("[User] Creating user <%v> with insertResult <%v>", user, createdUser)
	return createdUser
}

// https://www.mongodb.com/blog/post/quick-start-golang--mongodb--how-to-update-documents
// https://kb.objectrocket.com/mongo-db/how-to-update-a-mongodb-document-using-the-golang-driver-458
func updateUser(userID string, user User) *mongo.UpdateResult {
	id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": id}

	update := bson.M{
		"$set": bson.M{
			"username": user.Username,
			"isAdmin":  user.IsAdmin,
		},
	}

	// result := dbUserCollection.FindOneAndUpdate(context.Background(), filter, update)
	result, _ := dbUserCollection.UpdateOne(context.Background(), filter, update)
	log.Printf("[User] Creating user <%v> with result <%v>", user, result)
	return result
}

func deleteUser(userID string) int64 {
	id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": id}
	d, err := dbUserCollection.DeleteMany(context.Background(), filter, nil)
	if err != nil {
		log.Println("[User] error in user deletion: ", err)
	}

	log.Printf("[User] Deleting ID <%v>: %d count(s)", userID, d.DeletedCount)
	return d.DeletedCount
}
