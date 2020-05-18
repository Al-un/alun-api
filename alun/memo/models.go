package memo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Al-un/alun-api/alun/core"
)

const (
	accessPrivate = 10
	accessPublic  = 0
)

// BasicInfo provide simple information about Memo entities
type BasicInfo struct {
	Title       string `json:"title,omitempty" bson:"title,omitempty"`
	Description string `json:"description,omitempty" bson:"description,omitempty"`
}

func (bi *BasicInfo) equals(bi2 BasicInfo) bool {
	return bi.Title == bi2.Title && bi.Description == bi2.Description
}

// Board is a memo container
type Board struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id"`
	BasicInfo          `bson:",inline"`
	Access             int    `json:"access" bson:"access"`
	Memos              []Memo `json:"memos,omitempty" bson:"memos,omitempty"`
	core.TrackedEntity `bson:",inline"`
}

// Memo is a group of items to be remembered. Comparing to a manual TODO list
// or checklist, a memo would be a single page
type Memo struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id"`
	BasicInfo          `bson:",inline"`
	Items              []Item `json:"items,omitempty" bson:"items"`
	core.TrackedEntity `bson:",inline"`
	// BoardID            primitive.ObjectID `json:"boardId" bson:"boardId"`
}

func (m *Memo) equals(m2 Memo) bool {
	areFieldsEquals := m.ID == m2.ID &&
		m.BasicInfo.equals(m2.BasicInfo) &&
		m.TrackedEntity.Equals(m2.TrackedEntity)

	if !areFieldsEquals {
		return false
	}

	return areItemsArrayEquals(m.Items, m2.Items)
}

// Item is a single action or thing to remember
type Item struct {
	Text       string    `json:"text" bson:"text"`
	IsFinished bool      `json:"isFinished,omitempty" bson:"isFinished"`
	DueDate    time.Time `json:"dueDate,omitempty" bson:"dueDate,omitempty"`
}

// Equals checks the equality all fields and time values are checked with a precision
// of one minute
func (i *Item) equals(i2 Item) bool {
	return i.Text == i2.Text &&
		i.IsFinished == i2.IsFinished &&
		i.DueDate.Round(1*time.Minute).Equal(i2.DueDate.Round(1*time.Minute))
}

func areItemsArrayEquals(items1 []Item, items2 []Item) bool {
	if len(items1) != len(items2) {
		return false
	}
	for idx, item := range items1 {
		item2 := items2[idx]
		if !item.equals(item2) {
			return false
		}
	}

	return true
}
