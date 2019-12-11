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
var dbUserLoginCollectionName string

// dbUserCollection : user collection instance
var dbUserCollection *mongo.Collection
var dbUserLoginCollection *mongo.Collection

// Init the connection with MongoDB upon app initialisation
func init() {
	// Initialisation: collections name
	dbUserCollectionName = "al_users"
	dbUserLoginCollectionName = "al_users_login"

	// Initialisation: collections instances
	dbUserCollection = core.MongoDatabase.Collection(dbUserCollectionName)
	dbUserLoginCollection = core.MongoDatabase.Collection(dbUserLoginCollectionName)

	log.Println("[MongoDB] User initialisation!")
}

// ---------- CRUD ------------------------------------------------------------

// findUser fetches an user for a given username and CLEAR password
func findUserByUsernamePassword(username string, clearPassword string) (User, error) {
	var user User
	var hashedPassword = hashPassword(clearPassword)

	filter := bson.M{"username": username, "password": hashedPassword}
	if err := dbUserCollection.FindOne(context.TODO(), filter).Decode(&user); err != nil {
		return User{}, err
	}

	return user, nil
}

// findUserById fetches an user for a given ID
func findUserByID(userID string) (User, error) {
	var user User
	id, _ := primitive.ObjectIDFromHex(userID)

	filter := bson.M{"_id": id}
	if err := dbUserCollection.FindOne(context.TODO(), filter).Decode(&user); err != nil {
		return User{}, err
	}

	return user, nil
}

// findLoginWithValidToken find the first valid token for an user
// TODO: returns an array?
func findLoginWithValidToken(user User) (Login, error) {
	var login Login
	filter := bson.M{"userId": user.ID, "token.isInvalid": false}

	if err := dbUserLoginCollection.FindOne(context.TODO(), filter).Decode(&login); err != nil {
		return Login{}, err
	}

	return login, nil
}

func findLoginByToken(jwt string) (Login, error) {
	var login Login
	filter := bson.M{"token.jwt": jwt}

	if err := dbUserLoginCollection.FindOne(context.TODO(), filter).Decode(&login); err != nil {
		return Login{}, err
	}

	return login, nil
}

// createUser creates the user assuming that passowrd is not hashed yet
func createUser(user User) *mongo.InsertOneResult {
	var hashedPassword = hashPassword(user.Password)
	user.Password = hashedPassword

	createdUser, _ := dbUserCollection.InsertOne(context.TODO(), user)
	log.Printf("[User] Creating user <%v> with insertResult <%v>", user, createdUser)
	return createdUser
}

// createLogin just saves the login in the DB
func createLogin(login Login) *mongo.InsertOneResult {
	created, _ := dbUserLoginCollection.InsertOne(context.TODO(), login)
	log.Printf("[User] Creating login <%v> with insertResult <%v>", login, created)
	return created
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

	result, _ := dbUserCollection.UpdateOne(context.TODO(), filter, update)
	log.Printf("[User] Creating user <%v> with result <%v>", user, result)
	return result
}

// invalidateLogin set the isInvalid flag to true based on a JWT
func invalidateLogin(jwt string) *mongo.UpdateResult {
	filter := bson.M{"token.jwt": jwt}
	update := bson.M{
		"$set": bson.M{
			"token.isInvalid": true,
		},
	}

	result, _ := dbUserLoginCollection.UpdateOne(context.TODO(), filter, update)
	log.Printf("[User] Invalidate <%v> with result <%v>", jwt, result)
	return result

}

func deleteUser(userID string) int64 {
	id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": id}
	d, err := dbUserCollection.DeleteMany(context.TODO(), filter, nil)
	if err != nil {
		log.Println("[User] error in user deletion: ", err)
	}

	log.Printf("[User] Deleting ID <%v>: %d count(s)", userID, d.DeletedCount)
	return d.DeletedCount
}
