package core

import (
	"context"
	"time"

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
	// init() function are called in order of files so DAO needs to be init
	if MongoDatabase == nil {
		initDao()
	}

	// Initialisation: collections name
	dbUserCollectionName = "al_users"
	dbUserLoginCollectionName = "al_users_login"

	// Initialisation: collections instances
	dbUserCollection = MongoDatabase.Collection(dbUserCollectionName)
	dbUserLoginCollection = MongoDatabase.Collection(dbUserLoginCollectionName)

	coreLogger.Info("[MongoDB] User initialisation!")
}

// ---------- CRUD ------------------------------------------------------------

// findUser fetches an user for a given username and CLEAR password
func findUserByUsernamePassword(username string, clearPassword string) (User, error) {
	var authUser authenticatedUser
	var hashedPassword = hashPassword(clearPassword)

	filter := bson.M{"username": username, "password": hashedPassword}
	if err := dbUserCollection.FindOne(context.TODO(), filter).Decode(&authUser); err != nil {
		coreLogger.Verbose("Credentials %s/%s (hashed: %s) are NOT valid T_T due to error: %v",
			username, clearPassword, hashedPassword, err)
		return User{}, err
	}

	coreLogger.Verbose("Credentials %s/%s are valid \\o/", username, clearPassword)

	return authUser.User, nil
}

func isEmailAlreadyRegistered(email string) (bool, *ServiceMessage) {
	filter := bson.M{"email": email}
	userCount, err := dbUserCollection.CountDocuments(context.TODO(), filter)

	if err != nil {
		coreLogger.Info("Error when counting user with email %s %v", email, err)
		return false, NewServiceErrorMessage(err)
	}

	return userCount > 0, nil
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
	filter := bson.M{"userId": user.ID, "token.status": tokenStatusActive}

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

// createUser creates the user with only an email. The email is checked is
//
// returns newly created user
func createUser(user User) (User, *ServiceMessage) {

	coreLogger.Verbose("Checking %+v", user)

	// Check if email is already registered
	isEmailAlreadyTaken, errMsg := isEmailAlreadyRegistered(user.Email)
	if errMsg != nil {
		return User{}, errMsg
	}

	// Nut
	if isEmailAlreadyTaken {
		return User{}, hasEmailNotAvailable
	}

	coreLogger.Verbose("Creating %+v", user)
	createdUser, err := dbUserCollection.InsertOne(context.TODO(), user)
	if err != nil {
		coreLogger.Warn("Error when creating user of email %s: %v", user.Email, err)
		return User{}, NewServiceErrorMessage(err)
	}

	var newUser User
	filter := bson.M{"_id": createdUser.InsertedID}
	if err := dbUserCollection.FindOne(context.TODO(), filter).Decode(&newUser); err != nil {
		return User{}, NewServiceErrorMessage(err)
	}
	coreLogger.Verbose("[User] Created newUser <%+v>", newUser)

	return newUser, nil
}

func changePassword(pwdChgRequest pwdChangeRequest) (authenticatedUser, *ServiceMessage) {

	// Is password reset token found
	var user User
	filter := bson.M{"pwdResetToken.token": pwdChgRequest.Token}
	if err := dbUserCollection.FindOne(context.TODO(), filter).Decode(&user); err != nil {
		coreLogger.Debug(">>> %v", err)
		return authenticatedUser{}, pwdResetTokenNotFound
	}

	// Is password reset token expired?
	if user.PwdResetToken.ExpiresAt.Before(time.Now()) {
		return authenticatedUser{}, pwdResetTokenExpired
	}

	// Hash password
	hashedPassword := hashPassword(pwdChgRequest.Password)

	update := bson.M{
		// https://docs.mongodb.com/manual/reference/operator/update/set/
		"$set": bson.M{
			"password": hashedPassword,
		},
		// https://docs.mongodb.com/manual/reference/operator/update/unset/
		"$unset": bson.M{
			// https://stackoverflow.com/a/6852039/4906586
			"pwdResetToken": 1,
		},
	}

	var updatedUser authenticatedUser

	if err := dbUserCollection.FindOneAndUpdate(context.TODO(), filter, update).Decode(&updatedUser); err != nil {
		return authenticatedUser{}, NewServiceErrorMessage(err)
	}

	return updatedUser, nil
}

// createLogin just saves the login in the DB
func createLogin(login Login) (Login, error) {
	created, err := dbUserLoginCollection.InsertOne(context.TODO(), login)
	if err != nil {
		return Login{}, err
	}
	coreLogger.Debug("[User] Creating login <%v> with insertResult <%v>", login, created)

	var newLogin Login
	filter := bson.M{"_id": created.InsertedID}
	if err := dbUserLoginCollection.FindOne(context.TODO(), filter).Decode(&newLogin); err != nil {
		return Login{}, err
	}

	return newLogin, nil
}

// https://www.mongodb.com/blog/post/quick-start-golang--mongodb--how-to-update-documents
// https://kb.objectrocket.com/mongo-db/how-to-update-a-mongodb-document-using-the-golang-driver-458
func updateUser(userID string, user User) (User, error) {
	id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": id}

	update := bson.M{
		"$set": bson.M{
			"username": user.Username,
			"isAdmin":  user.IsAdmin,
		},
	}

	var updatedUser User

	if err := dbUserCollection.FindOneAndUpdate(context.TODO(), filter, update).Decode(&updatedUser); err != nil {
		return User{}, err
	}
	coreLogger.Debug("[User] Creating user <%v> with result <%v>", user, updatedUser)
	return updatedUser, nil
}

// invalidateToken invalidates the login for the given token by setting up a non-active
// status on the token
func invalidateToken(jwt string, invalidStatusCode int) (Login, error) {
	filter := bson.M{"token.jwt": jwt}
	update := bson.M{
		"$set": bson.M{
			"token.status": invalidStatusCode,
		},
	}

	var invalidatedLogin Login
	if err := dbUserLoginCollection.FindOneAndUpdate(context.TODO(), filter, update).Decode(&invalidatedLogin); err != nil {
		return Login{}, err
	}
	coreLogger.Debug("[User] Invalidate <%v> with result <%v>", jwt, invalidatedLogin)
	return invalidatedLogin, nil

}

func deleteUser(userID string) int64 {
	id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": id}
	d, err := dbUserCollection.DeleteMany(context.TODO(), filter, nil)
	if err != nil {
		coreLogger.Info("[User] error in user deletion: ", err)
	}

	coreLogger.Debug("[User] Deleting ID <%v>: %d count(s)", userID, d.DeletedCount)
	return d.DeletedCount
}
