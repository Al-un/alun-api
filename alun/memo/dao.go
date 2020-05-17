package memo

import (
	"context"

	"github.com/Al-un/alun-api/alun/core"
	"github.com/Al-un/alun-api/alun/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ---------- Variable and init -----------------------------------------------
var (
	dbMemoCollectionName  string
	dbMemoCollection      *mongo.Collection
	dbBoardCollectionName string
	dbBoardCollection     *mongo.Collection
	returnOpt             options.ReturnDocument = 1
)

// Init the connection with MongoDB upon app initialisation
func initDao() {
	_, memoMongoDb, err := core.MongoConnectFromEnvVar(utils.EnvVarMemoDbURL, memoLogger)
	if err != nil {
		memoLogger.Fatal(1, "%v", err)
	}

	// Initialisation: collections name
	dbBoardCollectionName = "al_memos_boards"
	dbMemoCollectionName = "al_memos_memos"

	// Initialisation: collections instances
	dbBoardCollection = memoMongoDb.Collection(dbBoardCollectionName)
	dbMemoCollection = memoMongoDb.Collection(dbMemoCollectionName)

	memoLogger.Debug("[MongoDB] Memo initialisation!")
}

// ---------- CRUD ------------------------------------------------------------
func findBoardsByUserID(userID string) ([]Board, *core.ServiceMessage) {
	id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{
		core.TrackedCreatedBy: id,
	}

	var boards []Board

	cur, err := dbBoardCollection.Find(context.TODO(), filter)
	if err != nil {
		return boards, core.NewServiceErrorMessage(err)
	}

	var next Board
	for cur.Next(context.TODO()) {
		cur.Decode(&next)
		boards = append(boards, next)
	}

	return boards, nil
}

func findMemoByID(memoID string) (*Memo, *core.ServiceMessage) {
	id, _ := primitive.ObjectIDFromHex(memoID)
	filter := bson.M{"_id": id}

	var memo Memo

	if err := dbMemoCollection.FindOne(context.TODO(), filter).Decode(&memo); err != nil {
		memoLogger.Debug("Memo fetching ID<%s/%s> error: ", memoID, id, err)
		return nil, core.NewServiceErrorMessage(err)
	}

	return &memo, nil
}

func findBoardByID(boardID string) (*Board, *core.ServiceMessage) {
	id, _ := primitive.ObjectIDFromHex(boardID)
	filter := bson.M{"_id": id}

	var board Board

	if err := dbBoardCollection.FindOne(context.TODO(), filter).Decode(&board); err != nil {
		return nil, core.NewServiceErrorMessage(err)
	}

	return &board, nil
}

func createBoard(toCreateBoard Board) (*Board, *core.ServiceMessage) {
	insertResult, err := dbBoardCollection.InsertOne(context.TODO(), toCreateBoard)
	if err != nil {
		return nil, core.NewServiceErrorMessage(err)
	}

	var newBoard Board
	filter := bson.M{"_id": insertResult.InsertedID}
	if err := dbBoardCollection.FindOne(context.TODO(), filter).Decode(&newBoard); err != nil {
		return nil, core.NewServiceErrorMessage(err)
	}

	return &newBoard, nil
}

func createMemo(toCreateMemo Memo) (*Memo, *core.ServiceMessage) {
	memoLogger.Verbose("Creating %s with items %v", toCreateMemo.Title, toCreateMemo.Items)
	toCreateMemo.ID = primitive.NewObjectID()

	insertResult, err := dbMemoCollection.InsertOne(context.TODO(), toCreateMemo)
	if err != nil {
		memoLogger.Debug("Memo creation error: ", err)
		return nil, core.NewServiceErrorMessage(err)
	}

	var newMemo Memo
	filter := bson.M{"_id": insertResult.InsertedID}
	if err := dbMemoCollection.FindOne(context.TODO(), filter).Decode(&newMemo); err != nil {
		memoLogger.Debug("Memo fetching error: ", err)
		return nil, core.NewServiceErrorMessage(err)
	}

	return &newMemo, nil
}

func updateBoard(boardID string, toUpdateBoard Board) (*Board, *core.ServiceMessage) {
	id, _ := primitive.ObjectIDFromHex(boardID)
	filter := bson.M{
		"_id": id,
	}
	options := &options.FindOneAndUpdateOptions{
		ReturnDocument: &returnOpt,
	}
	update := bson.M{
		"title":       toUpdateBoard.Title,
		"description": toUpdateBoard.Description,
		"access":      toUpdateBoard.Access,
		"modifiedBy":  toUpdateBoard.ModifiedBy,
		"modifiedAt":  toUpdateBoard.ModifiedAt,
	}

	var updatedBoard Board
	if err := dbBoardCollection.FindOneAndUpdate(context.TODO(), filter, update, options).Decode(&updatedBoard); err != nil {
		return nil, core.NewServiceErrorMessage(err)
	}

	return &updatedBoard, nil
}

func updateMemo(memoID string, toUpdateMemo Memo) (*Memo, *core.ServiceMessage) {
	id, _ := primitive.ObjectIDFromHex(memoID)
	filter := bson.M{
		"_id": id,
	}
	options := &options.FindOneAndUpdateOptions{
		ReturnDocument: &returnOpt,
	}
	update := bson.M{
		"title":       toUpdateMemo.Title,
		"description": toUpdateMemo.Description,
		"items":       toUpdateMemo.Items,
		"modifiedBy":  toUpdateMemo.ModifiedBy,
		"modifiedAt":  toUpdateMemo.ModifiedAt,
	}

	var updatedMemo Memo
	if err := dbMemoCollection.FindOneAndUpdate(context.TODO(), filter, update, options).Decode(&updatedMemo); err != nil {
		return nil, core.NewServiceErrorMessage(err)
	}

	return &updatedMemo, nil
}

func deleteBoard(boardID string) (int64, int64, *core.ServiceMessage) {
	id, _ := primitive.ObjectIDFromHex(boardID)

	// Delete board
	filter := bson.M{"_id": id}
	deletedBoard, err := dbBoardCollection.DeleteMany(context.TODO(), filter, nil)
	if err != nil {
		return -1, -1, core.NewServiceErrorMessage(err)
	}

	// Delete memos
	filter = bson.M{"boardId": id}
	deletedMemos, err := dbMemoCollection.DeleteMany(context.TODO(), filter, nil)
	if err != nil {
		return -1, -1, core.NewServiceErrorMessage(err)
	}

	return deletedBoard.DeletedCount, deletedMemos.DeletedCount, nil
}

func deleteMemo(memoID string) (int64, *core.ServiceMessage) {
	id, _ := primitive.ObjectIDFromHex(memoID)
	filter := bson.M{"_id": id}
	d, err := dbMemoCollection.DeleteMany(context.TODO(), filter, nil)
	if err != nil {
		return -1, core.NewServiceErrorMessage(err)
	}

	return d.DeletedCount, nil
}
