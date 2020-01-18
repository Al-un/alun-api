package memo

import (
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Al-un/alun-api/alun/core"
)

// Memo is a group of items to be remembered. Comparing to a manual TODO list
// or checklist, a memo would be a single page
type Memo struct {
	ID    primitive.ObjectID `json:"id" bson:"_id"`
	Title string             `json:"title,omitempty" bson:"title,omitempty"`
	Items []Item             `json:"items,omitempty" bson:"items"`
	// 0=Public, 1=Private
	Visibility         int `json:"visibility" bson:"visibility"`
	core.TrackedEntity `bson:",inline"`
}

type memoList struct {
	ID         primitive.ObjectID `json:"id" bson:"_id"`
	Title      string             `json:"title,omitempty" bson:"title,omitempty"`
	Visibility int                `json:"visibility" bson:"visibility"`
	core.TrackedEntity
}

// Item is a single action or thing to remember
type Item struct {
	Text       string `json:"text" bson:"text"`
	IsFinished bool   `json:"isFinished,omitempty" bson:"isFinished"`
}
