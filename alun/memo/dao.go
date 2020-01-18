package memo

import (
	"context"
	"fmt"

	"github.com/Al-un/alun-api/alun/core"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ---------- Variable and init -----------------------------------------------

var dbMemoCollectionName string
var dbMemoCollection *mongo.Collection

// Init the connection with MongoDB upon app initialisation
func init() {
	// Initialisation: collections name
	dbMemoCollectionName = "al_memos"

	// Initialisation: collections instances
	dbMemoCollection = core.MongoDatabase.Collection(dbMemoCollectionName)

	memoLogger.Debug("[MongoDB] Memo initialisation!")
}

// ---------- CRUD ------------------------------------------------------------
func findMemosByUserID(userID string) ([]memoList, error) {
	id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{
		core.TrackedCreatedBy: id,
	}

	var memos []memoList

	cur, err := dbMemoCollection.Find(context.TODO(), filter)
	if err != nil {
		memoLogger.Verbose("Memo listing error: ", err)
		return memos, err
	}

	var next memoList
	for cur.Next(context.TODO()) {
		cur.Decode(&next)
		memos = append(memos, next)
	}

	memoLogger.Verbose("Listing: %v", memos)

	return memos, nil
}

func findMemoByID(memoID string) (Memo, error) {
	id, _ := primitive.ObjectIDFromHex(memoID)
	filter := bson.M{"_id": id}

	var memo Memo

	if err := dbMemoCollection.FindOne(context.TODO(), filter).Decode(&memo); err != nil {
		memoLogger.Debug("Memo fetching ID<%s/%s> error: ", memoID, id, err)
		return Memo{}, err
	}

	return memo, nil
}

func createMemo(toCreateMemo Memo) (Memo, error) {
	memoLogger.Debug("Creating %s with items %v", toCreateMemo.Title, toCreateMemo.Items)
	toCreateMemo.ID = primitive.NewObjectID()

	insertResult, err := dbMemoCollection.InsertOne(context.TODO(), toCreateMemo)
	if err != nil {
		memoLogger.Debug("Memo creation error: ", err)
		return Memo{}, err
	}

	var newMemo Memo
	filter := bson.M{"_id": insertResult.InsertedID}
	if err := dbMemoCollection.FindOne(context.TODO(), filter).Decode(&newMemo); err != nil {
		memoLogger.Debug("Memo fetching error: ", err)
		return Memo{}, err
	}

	return newMemo, nil
}

func updateMemo(memoID string, toUpdateMemo Memo) (Memo, error) {
	id, _ := primitive.ObjectIDFromHex(memoID)
	filter := bson.M{
		"_id": id,
	}

	var returnOpt options.ReturnDocument = 1

	options := &options.FindOneAndReplaceOptions{
		ReturnDocument: &(returnOpt),
	}

	fmt.Printf("To update memo: %v / %v / %v\n", toUpdateMemo, toUpdateMemo.CreatedBy, toUpdateMemo.CreatedAt)

	var memo Memo
	if err := dbMemoCollection.FindOneAndReplace(context.TODO(), filter, toUpdateMemo, options).Decode(&memo); err != nil {
		memoLogger.Debug("Update of memo <%s> error: %v", memoID, err)
		return Memo{}, err

	}

	return memo, nil
}

func deleteMemo(memoID string) int64 {
	id, _ := primitive.ObjectIDFromHex(memoID)
	filter := bson.M{"_id": id}
	d, err := dbMemoCollection.DeleteMany(context.TODO(), filter, nil)
	if err != nil {
		memoLogger.Debug("Memo deletion error: ", err)
		return -1
	}

	memoLogger.Verbose("Deleting Memo#%s has count: %d", memoID, d.DeletedCount)

	return d.DeletedCount
}
