package user

import (
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// ----------------------------------------------------------------------------
//	Constants
// ----------------------------------------------------------------------------
// authStatus ensures consistency between authorization check returned code
var authStatus = struct {
	isAuthorized           int // All good
	isNotAuthorized        int // All good but nope
	isTokenExpired         int // Timing is everything
	isTokenMalformed       int // What the hell was sent?
	isTokenInvalid         int // Unknown Token parsing error
	isTokenInvalidated     int // Token has been manually invalidated
	isAuthorizationMissing int // Hey, you need to send something!
	isAuthorizationInvalid int // You need to send something correct
	isUnknownError         int // I just don't know
}{
	isAuthorized:           0,
	isNotAuthorized:        -1,
	isTokenExpired:         -2,
	isTokenMalformed:       -3,
	isTokenInvalid:         -16,
	isTokenInvalidated:     -17,
	isAuthorizationMissing: -4,
	isAuthorizationInvalid: -5,
	isUnknownError:         -9,
}

var authErrorMsg = struct {
	isNotAuthorized        string
	isTokenExpired         string
	isTokenMalformed       string
	isTokenInvalid         string
	isTokenInvalidated     string
	isAuthorizationMissing string
	isAuthorizationInvalid string
	isUnknownError         string
}{
	isNotAuthorized:        "Not authorized",
	isTokenExpired:         "Token is expired",
	isTokenMalformed:       "Token is malformed",
	isTokenInvalid:         "Token is just invalid",
	isTokenInvalidated:     "Token has been invalidated",
	isAuthorizationMissing: "Authorization header is missing",
	isAuthorizationInvalid: "Authorization header must start with \"Bearer \"",
	isUnknownError:         "Unknown error during Authorization check",
}

// AuthCheckMode is a convenient int enum for AuthCheckConfig.Mode
var AuthCheckMode = struct {
	isLogged      int
	isAdmin       int
	isUser        int
	isAdminOrUser int
}{
	isLogged:      1,
	isAdmin:       2,
	isUser:        3,
	isAdminOrUser: 4,
}

// AuthCheckIsLogged is a shortcut to select "authenticated users only"
var AuthCheckIsLogged = AuthCheckConfig{
	Mode: AuthCheckMode.isLogged,
}

// AuthCheckIsAdmin is a shortcut for "admin only"
var AuthCheckIsAdmin = AuthCheckConfig{
	Mode: AuthCheckMode.isAdmin,
}

// AuthCheckIsUser is a convenience method for getting a "this user only check"
func AuthCheckIsUser(userID string) AuthCheckConfig {
	return AuthCheckConfig{
		Mode:   AuthCheckMode.isUser,
		UserID: userID,
	}
}

// AuthCheckIsAdminOrUser is a convenience method for getting a "this user or admin check"
func AuthCheckIsAdminOrUser(userID string) AuthCheckConfig {
	return AuthCheckConfig{
		Mode:   AuthCheckMode.isAdminOrUser,
		UserID: userID,
	}
}

// ----------------------------------------------------------------------------
//	Types
// ----------------------------------------------------------------------------

// User represents a loggable entity
// db.al_users.insertOne({username:"pouet", password:"plop"})
//
// curl http://localhost:8000/users/register --data '{"username": "plop", "password": "plop"}'
type User struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Username string             `json:"username" bson:"username"`
	Password string             `json:"-" bson:"password"` // not present in JSON: https://golang.org/pkg/encoding/json/
	IsAdmin  bool               `json:"isAdmin" bson:"isAdmin"`
}

// JwtToken extends standard JWT
type JwtToken struct {
	jwt.Token
}

// JwtClaims extends standard claims for our User model
type JwtClaims struct {
	IsAdmin bool   `json:"isAdmin"`
	UserID  string `json:"userId"`
	jwt.StandardClaims
}

// Token saves the generated JWT in the database for re-usability
// or other features such as token invalidation
type Token struct {
	Jwt       string    `json:"jwt" bson:"jwt,omitempty"`                       // Stringified JWT
	ExpiresOn time.Time `json:"expiresOn,omitempty" bson:"expiresOn,omitempty"` // Convenience for checking token expiration
	IsInvalid bool      `json:"isInvalid" bson:"isInvalid"`                     // Token is expired or invalidated by user
}

// Login tracks user login and associated generated token. If a token was re-generated,
// it should create another Login as there is no automatic regeneration upon token expiration
type Login struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"` // Uniquely identify to invalidate it
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`        // Login timestamp
	UserID    primitive.ObjectID `json:"userId" bson:"userId"`
	Token     Token              `json:"token,omitempty" bson:"token,omitempty"` // Token generated during login
}

// AuthCheckConfig encapsulates the information required for authentication
type AuthCheckConfig struct {
	Mode   int    // check AuthCheckMode for convenience
	UserID string // for user-based check
}
