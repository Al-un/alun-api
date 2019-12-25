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
	CreatedOn  time.Time          `json:"createdOn,omitempty" bson:"createdOn,omitempty"`
	ModifiedBy primitive.ObjectID `json:"modifiedBy,omitempty" bson:"modifiedBy,omitempty"`
	ModifiedOn time.Time          `json:"modifiedOn,omitempty" bson:"mofidiedOn,omitempty"`
}

// ServiceMessage is a token to forward the status of an action to the next function /
// whatever handler processing it.
//
// While it is meant to standardize errors handling, it can also help to identify internal
// success status thanks to its code
type ServiceMessage struct {
	Code       int    `json:"code"`    // Internal code: 0 is fine, any code different from 0 is an error
	Message    string `json:"message"` // Explicit description
	HTTPStatus int    `json:"-"`       // HTTP Status code, skipped during serialisation
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
