package core

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
	IsAdmin  bool               `json:"isAdmin" bson:"isAdmin"`
}

// authenticatedUser has the password field so that when the server sends
// an User back to the client, the Password is not sent.

// However, the  JSON field is required otherwise the decoding will ignore
// the password field
type authenticatedUser struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Username string             `json:"username" bson:"username"`
	IsAdmin  bool               `json:"isAdmin" bson:"isAdmin"`
	Password string             `json:"password" bson:"password"` // not present in JSON: https://golang.org/pkg/encoding/json/
}

func (au *authenticatedUser) extractUser() User {
	return User{
		ID:       au.ID,
		Username: au.Username,
		IsAdmin:  au.IsAdmin,
	}
}

// JwtClaims extends standard claims for our User model.
//
// By including the IsAdmin and UserID fields, authorization check can be
// based on those values
type JwtClaims struct {
	IsAdmin bool   `json:"isAdmin"`
	UserID  string `json:"userId"`
	jwt.StandardClaims
}

// Token saves the generated JWT in the database for re-usability or other
// features such as token invalidation
type token struct {
	Jwt       string    `json:"jwt" bson:"jwt,omitempty"`                       // Stringified JWT
	ExpiresOn time.Time `json:"expiresOn,omitempty" bson:"expiresOn,omitempty"` // Convenience for checking token expiration
	Status    int       `json:"status" bson:"status"`                           // Token status
}

const tokenStatusActive = 0       // Token is still active
const tokenStatusLogout = 1       // Token has been disabled by an user logout
const tokenStatusExpired = 2      // Token has expired
const tokenStatusInvalidated = 10 // Token has been manually disabled by user

// Login tracks user login and associated generated token.
//
// If a token was re-generated, it should create another Login as there is
// no automatic regeneration upon token expiration.
// UserID field is required to avoid decoding the JWT from `Token`
type Login struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"` // Uniquely identify to invalidate it
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`        // Login timestamp
	UserID    primitive.ObjectID `json:"userId" bson:"userId"`              // Logged-in user, should match the token of the token :)
	Token     token              `json:"token" bson:"token"`                // Token generated during login
}

// AuthenticatedHandler is meant to be the core logic of the handler with the user
// informations already extracted from the request.
//
// claims are assumed to be always non-nil and always validated beforehand
type AuthenticatedHandler func(w http.ResponseWriter, r *http.Request, claims JwtClaims)

// AccessChecker ensures that the provided request is allowed to proceed or not.
//
// Most of the checks are based on the token header. An AccessChecker always
// assumes that the JWT is properly formed, hence the jwtClaims argument.
// An AccessChecker's check success often leads to some AuthenticatedHandler to
// proceed.
type AccessChecker func(r *http.Request, jwtClaims JwtClaims) bool

type successfulLogin struct {
	Token string `json:"token"`
}
