package memo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Al-un/alun-api/alun/core"
)

const (
	accessPrivate   = 20
	accessShareable = 10
	accessPublic    = 0
)

// BasicInfo provide simple information about Memo entities
type BasicInfo struct {
	Title       string `json:"title,omitempty" bson:"title,omitempty"`
	Description string `json:"description,omitempty" bson:"description,omitempty"`
}

// Board is a memo container
type Board struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id"`
	BasicInfo          `bson:",inline"`
	Access             int `json:"access" bson:"access"`
	core.TrackedEntity `bson:",inline"`
}

// BoardWithMemo is only for sending a JSON response of Board with memos
type BoardWithMemo struct {
	Board
	Memos []Memo `json:"memos"`
}

// Memo is a group of items to be remembered. Comparing to a manual TODO list
// or checklist, a memo would be a single page
type Memo struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id"`
	BasicInfo          `bson:",inline"`
	BoardID            primitive.ObjectID `json:"boardId" bson:"boardId"`
	Items              []Item             `json:"items,omitempty" bson:"items"`
	core.TrackedEntity `bson:",inline"`
}

// Item is a single action or thing to remember
type Item struct {
	Text       string    `json:"text" bson:"text"`
	IsFinished bool      `json:"isFinished,omitempty" bson:"isFinished"`
	DueDate    time.Time `json:"dueDate,omitempty" bson:"dueDate,omitempty"`
}
