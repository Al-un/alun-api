package core

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CheckStatus encapsulates a status code as well as an Error message
// to send as a default response
type CheckStatus struct {
	Code    int
	Message string
}

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
