package core

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ----------------------------------------------------------------------------
//	Types: Authentication and Authorization
// ----------------------------------------------------------------------------

// JwtClaims extends standard claims for our User model.
//
// By including the IsAdmin and UserID fields, authorization check can be
// based on those values
type JwtClaims struct {
	IsAdmin bool   `json:"isAdmin"`
	UserID  string `json:"userId"`
	jwt.StandardClaims
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

// ----------------------------------------------------------------------------
//	Types: Basic data model
// ----------------------------------------------------------------------------

// TrackedEntity is the basic structure for all entities which require tracking:
// user tracking and time tracking
//
// User reference are `primitive.ObjectID` to match "primary keys" of the users
// collection
type TrackedEntity struct {
	CreatedBy  primitive.ObjectID `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
	CreatedAt  time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	ModifiedBy primitive.ObjectID `json:"modifiedBy,omitempty" bson:"modifiedBy,omitempty"`
	ModifiedAt time.Time          `json:"modifiedAt,omitempty" bson:"modifiedAt,omitempty"`
}

// PrepareForCreate set creation related fields
func (t *TrackedEntity) PrepareForCreate(claims JwtClaims) {
	t.CreatedAt = time.Now()
	userID, _ := primitive.ObjectIDFromHex(claims.UserID)
	t.CreatedBy = userID
}

// PrepareForUpdate updates modification related field before any update based on
// the UserID provided by the claims
func (t *TrackedEntity) PrepareForUpdate(claims JwtClaims) {
	t.ModifiedAt = time.Now()
	userID, _ := primitive.ObjectIDFromHex(claims.UserID)
	t.ModifiedBy = userID
}

// Equals check the equality of each field and time fields are compared with a
// precision of one second
func (t *TrackedEntity) Equals(t2 TrackedEntity) bool {
	return t.CreatedBy == t2.CreatedBy &&
		t.CreatedAt.Round(1*time.Second).Equal(t2.CreatedAt.Round(1*time.Second)) &&
		t.ModifiedBy == t2.ModifiedBy &&
		t.ModifiedAt.Round(1*time.Second).Equal(t2.ModifiedAt.Round(1*time.Second))
}

const (
	// TrackedCreatedBy is the createdBy key. TrackedEntity is assumed to in "bson:,inline"
	TrackedCreatedBy = "createdBy"
	// TrackedCreatedAt is the createdAt key. TrackedEntity is assumed to in "bson:,inline"
	TrackedCreatedAt = "createdAt"
	// TrackedModifiedBy is the modifiedBy key. TrackedEntity is assumed to in "bson:,inline"
	TrackedModifiedBy = "modifiedBy"
	// TrackedModifiedAt is the modifiedAt key. TrackedEntity is assumed to in "bson:,inline"
	TrackedModifiedAt = "modifiedAt"
)

// ----------------------------------------------------------------------------
//	Types: Internal object
// ----------------------------------------------------------------------------

// ServiceMessage is a token to forward the status of an action to the next function /
// whatever handler processing it.
//
// While it is meant to standardize errors handling, it can also help to identify internal
// success status thanks to its code
type ServiceMessage struct {
	Code       int    `json:"code"`            // Internal code: 0 is fine, any code different from 0 is an error
	Message    string `json:"message"`         // Explicit description
	HTTPStatus int    `json:"-"`               // HTTP Status code, skipped during serialisation
	Error      error  `json:"error,omitempty"` // Error if any
}

// Write writes the ServiceMessage into a Http.ReponseWriter and uses the incoming request
// for logging purpose only
func (msg *ServiceMessage) Write(rw http.ResponseWriter, req *http.Request) {
	// Write response
	rw.WriteHeader(msg.HTTPStatus)
	json.NewEncoder(rw).Encode(msg)

	// Log
	coreLogger.Info("[%d] %s => %d %s", msg.Code, req.URL.Path, msg.HTTPStatus, msg.Message)
}

// NewServiceErrorMessage generate a ServiceMessage from an error. By default, status
// error is 500
func NewServiceErrorMessage(err error) *ServiceMessage {
	return &ServiceMessage{
		HTTPStatus: 500,
		Error:      err,
	}
}
