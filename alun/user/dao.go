package user

import (
	"context"
	"time"

	"github.com/Al-un/alun-api/alun/core"
	"github.com/Al-un/alun-api/alun/utils"

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
	_, mongoDb, err := core.MongoConnectFromEnvVar(utils.EnvVarUserDbURL)
	if err != nil {
		userLogger.Fatal(1, "%v", err)
	}

	// Initialisation: collections name
	dbUserCollectionName = "al_users"
	dbUserLoginCollectionName = "al_users_login"

	// Initialisation: collections instances
	dbUserCollection = mongoDb.Collection(dbUserCollectionName)
	dbUserLoginCollection = mongoDb.Collection(dbUserLoginCollectionName)

	userLogger.Info("[MongoDB] User initialisation!")
}

// ---------- CRUD ------------------------------------------------------------

// findUser fetches an user for a given email and CLEAR password
func findUserByEmailPassword(email string, clearPassword string) (User, error) {
	var authUser authenticatedUser
	var hashedPassword = hashPassword(clearPassword)

	filter := bson.M{"email": email, "password": hashedPassword}
	if err := dbUserCollection.FindOne(context.TODO(), filter).Decode(&authUser); err != nil {
		userLogger.Verbose("Credentials %s/%s (hashed: %s) are NOT valid T_T due to error: %v",
			email, clearPassword, hashedPassword, err)
		return User{}, err
	}

	userLogger.Verbose("Credentials %s/%s are valid \\o/", email, clearPassword)

	return authUser.User, nil
}

func isEmailAlreadyRegistered(email string) (bool, *core.ServiceMessage) {
	filter := bson.M{"email": email}
	userCount, err := dbUserCollection.CountDocuments(context.TODO(), filter)

	if err != nil {
		userLogger.Info("Error when counting user with email %s %v", email, err)
		return false, core.NewServiceErrorMessage(err)
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
func createUser(user User) (User, *core.ServiceMessage) {

	userLogger.Verbose("Checking %+v", user)

	// Check if email is already registered
	isEmailAlreadyTaken, errMsg := isEmailAlreadyRegistered(user.Email)
	if errMsg != nil {
		return User{}, errMsg
	}

	// Nut
	if isEmailAlreadyTaken {
		return User{}, hasEmailNotAvailable
	}

	userLogger.Verbose("Creating %+v", user)
	createdUser, err := dbUserCollection.InsertOne(context.TODO(), user)
	if err != nil {
		userLogger.Warn("Error when creating user of email %s: %v", user.Email, err)
		return User{}, core.NewServiceErrorMessage(err)
	}

	var newUser User
	filter := bson.M{"_id": createdUser.InsertedID}
	if err := dbUserCollection.FindOne(context.TODO(), filter).Decode(&newUser); err != nil {
		return User{}, core.NewServiceErrorMessage(err)
	}
	userLogger.Verbose("[User] Created newUser <%+v>", newUser)

	return newUser, nil
}

func changePassword(pwdChgRequest pwdChangeRequest) (authenticatedUser, *core.ServiceMessage) {

	// Is password reset token found
	var user User
	filter := bson.M{"pwdResetToken.token": pwdChgRequest.Token}
	if err := dbUserCollection.FindOne(context.TODO(), filter).Decode(&user); err != nil {
		userLogger.Debug(">>> %v", err)
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
		return authenticatedUser{}, core.NewServiceErrorMessage(err)
	}

	return updatedUser, nil
}

// createLogin just saves the login in the DB
func createLogin(login Login) (Login, error) {
	created, err := dbUserLoginCollection.InsertOne(context.TODO(), login)
	if err != nil {
		return Login{}, err
	}
	userLogger.Debug("[User] Creating login <%v> with insertResult <%v>", login, created)

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
			// "isAdmin":  user.IsAdmin,
		},
	}

	var updatedUser User

	if err := dbUserCollection.FindOneAndUpdate(context.TODO(), filter, update).Decode(&updatedUser); err != nil {
		return User{}, err
	}
	userLogger.Debug("[User] Creating user <%v> with result <%v>", user, updatedUser)
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
	userLogger.Debug("[User] Invalidate <%v> with result <%v>", jwt, invalidatedLogin)
	return invalidatedLogin, nil

}

func deleteUser(userID string) int64 {
	id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": id}
	d, err := dbUserCollection.DeleteMany(context.TODO(), filter, nil)
	if err != nil {
		userLogger.Info("[User] error in user deletion: ", err)
	}

	userLogger.Debug("[User] Deleting ID <%v>: %d count(s)", userID, d.DeletedCount)
	return d.DeletedCount
}
