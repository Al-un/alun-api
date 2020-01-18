package core

import (
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TrackedEntity is the basic structure for all entities which require tracking:
// user tracking and time tracking
//
// User reference are `primitive.ObjectID` to match "primary keys" of the users
// collection
type TrackedEntity struct {
	CreatedBy  primitive.ObjectID `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
	CreatedAt  time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	ModifiedBy primitive.ObjectID `json:"modifiedBy,omitempty" bson:"modifiedBy,omitempty"`
	ModifiedAt time.Time          `json:"modifiedAt,omitempty" bson:"mofidiedAt,omitempty"`
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
