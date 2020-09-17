package memo

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Al-un/alun-api/alun/core"
	"github.com/Al-un/alun-api/alun/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ---------- Variable and init -----------------------------------------------
var (
	dbMemoCollectionName string
	dbMemoCollection     *mongo.Collection
	returnOpt            options.ReturnDocument = 1
)

// Init the connection with MongoDB upon app initialisation
func initDao() {
	_, memoMongoDb, err := core.MongoConnectFromEnvVar(utils.EnvVarMemoDbURL, memoLogger)
	if err != nil {
		memoLogger.Fatal(1, "%v", err)
	}

	// Initialisation: collections name
	dbMemoCollectionName = "al_memos"

	// Initialisation: collections instances
	dbMemoCollection = memoMongoDb.Collection(dbMemoCollectionName)

	memoLogger.Debug("[MongoDB] Memo initialisation!")
}

// ---------- CRUD ------------------------------------------------------------
func findBoardsByUserID(userID string) ([]Board, *core.ServiceMessage) {
	id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{
		core.TrackedCreatedBy: id,
	}
	options := &options.FindOptions{
		Projection: bson.M{
			"memos": 0,
		},
	}

	boards := make([]Board, 0)

	cur, err := dbMemoCollection.Find(context.TODO(), filter, options)
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

func findBoardByID(boardID string) (*Board, *core.ServiceMessage) {
	id, _ := primitive.ObjectIDFromHex(boardID)
	filter := bson.M{"_id": id}

	var board Board

	if err := dbMemoCollection.FindOne(context.TODO(), filter).Decode(&board); err != nil {
		return nil, core.NewServiceErrorMessage(err)
	}

	return &board, nil
}
func findMemoByID(boardID string, memoID string) (*Memo, *core.ServiceMessage) {
	bID, _ := primitive.ObjectIDFromHex(boardID)
	mID, _ := primitive.ObjectIDFromHex(memoID)
	filter := bson.M{
		"_id":       bID,
		"memos._id": mID,
	}
	options := &options.FindOneOptions{
		Projection: bson.M{
			"memos.$": 1,
		},
	}

	var board Board

	if err := dbMemoCollection.FindOne(context.TODO(), filter, options).Decode(&board); err != nil {
		return nil, core.NewServiceErrorMessage(err)
	}

	var memo Memo
	memo = board.Memos[0]

	return &memo, nil
}

// func findBoardWithMemosByID(boardID string) (*BoardWithMemo, *core.ServiceMessage) {
// 	id, _ := primitive.ObjectIDFromHex(boardID)

// 	matchStage := bson.D{primitive.E{Key: "$match", Value: bson.D{
// 		primitive.E{Key: "_id", Value: id},
// 	}}}
// 	lookupStage := bson.D{primitive.E{Key: "$lookup", Value: bson.D{
// 		primitive.E{Key: "from", Value: dbMemoCollectionName},
// 		primitive.E{Key: "localField", Value: "_id"},
// 		primitive.E{Key: "foreignField", Value: "boardId"},
// 		primitive.E{Key: "as", Value: "memos"},
// 	}}}
// 	// matchStage := bson.D{primitive.E{Key: "$match", Value: bson.D{
// 	// 	primitive.E{Key: "boardId", Value: id},
// 	// }}}
// 	// lookupStage := bson.D{primitive.E{Key: "$lookup", Value: bson.D{
// 	// 	primitive.E{Key: "from", Value: dbBoardCollectionName},
// 	// 	primitive.E{Key: "localField", Value: "boardId"},
// 	// 	primitive.E{Key: "foreignField", Value: "_id"},
// 	// 	primitive.E{Key: "as", Value: "memos"},
// 	// }}}
// 	// unwindStage := bson.D{primitive.E{Key: "$unwind", Value: bson.D{
// 	// 	primitive.E{Key: "path", Value: "$memos"},
// 	// 	primitive.E{Key: "preserveNullAndEmptyArrays", Value: false},
// 	// }}}

// 	cursor, err := dbBoardCollection.Aggregate(context.TODO(), mongo.Pipeline{matchStage, lookupStage})
// 	if err != nil {
// 		return nil, core.NewServiceErrorMessage(err)
// 	}

// 	// var memos []bson.M
// 	var memos []BoardWithMemo
// 	if err = cursor.All(context.TODO(), &memos); err != nil {
// 		return nil, core.NewServiceErrorMessage(err)
// 	}

// 	fmt.Printf("Loaded: %+v\n", memos)

// 	var board BoardWithMemo

// 	return &board, nil
// }

func isBoardIDExist(boardID string) (bool, *core.ServiceMessage) {
	id, _ := primitive.ObjectIDFromHex(boardID)
	filter := bson.M{"_id": id}

	count, err := dbMemoCollection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return false, core.NewServiceErrorMessage(err)
	}

	return count > 0, nil
}

// func isMemoIDExist(memoID string) (bool, *core.ServiceMessage) {
// 	id, _ := primitive.ObjectIDFromHex(memoID)
// 	filter := bson.M{"_id": id}

// 	count, err := dbMemoCollection.CountDocuments(context.TODO(), filter)
// 	if err != nil {
// 		return false, core.NewServiceErrorMessage(err)
// 	}

// 	return count > 0, nil
// }

func createBoard(toCreateBoard Board) (*Board, *core.ServiceMessage) {
	toCreateBoard.ID = primitive.NewObjectID()

	insertResult, err := dbMemoCollection.InsertOne(context.TODO(), toCreateBoard)
	if err != nil {
		return nil, core.NewServiceErrorMessage(err)
	}

	var newBoard Board
	filter := bson.M{"_id": insertResult.InsertedID}
	if err := dbMemoCollection.FindOne(context.TODO(), filter).Decode(&newBoard); err != nil {
		return nil, core.NewServiceErrorMessage(err)
	}

	return &newBoard, nil
}

func createMemo(boardID string, toCreateMemo Memo) (*Memo, *core.ServiceMessage) {
	memoLogger.Verbose("Creating %s with items %v", toCreateMemo.Title, toCreateMemo.Items)

	toCreateMemo.ID = primitive.NewObjectID()

	bID, _ := primitive.ObjectIDFromHex(boardID)
	filter := bson.M{
		"_id": bID,
	}
	update := bson.M{
		"$push": bson.M{
			"memos": toCreateMemo,
		},
	}
	if err := dbMemoCollection.FindOneAndUpdate(context.TODO(), filter, update).Err(); err != nil {
		return nil, core.NewServiceErrorMessage(err)
	}

	return &toCreateMemo, nil
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
		"$set": bson.M{
			"title":       toUpdateBoard.Title,
			"description": toUpdateBoard.Description,
			"access":      toUpdateBoard.Access,
			"updatedBy":   toUpdateBoard.UpdatedBy,
			"updatedAt":   toUpdateBoard.UpdatedAt,
		},
	}

	var updatedBoard Board
	if err := dbMemoCollection.FindOneAndUpdate(context.TODO(), filter, update, options).Decode(&updatedBoard); err != nil {
		return nil, core.NewServiceErrorMessage(err)
	}

	return &updatedBoard, nil
}

func updateMemo(boardID string, memoID string, toUpdateMemo Memo) (*Memo, *core.ServiceMessage) {
	bID, _ := primitive.ObjectIDFromHex(boardID)
	mID, _ := primitive.ObjectIDFromHex(memoID)
	filter := bson.M{
		"_id":       bID,
		"memos._id": mID,
	}
	options := &options.FindOneAndUpdateOptions{
		ReturnDocument: &returnOpt,
	}
	update := bson.M{
		"$set": bson.M{
			"memos.$.title":       toUpdateMemo.Title,
			"memos.$.description": toUpdateMemo.Description,
			"memos.$.items":       toUpdateMemo.Items,
			"memos.$.updatedBy":   toUpdateMemo.UpdatedBy,
			"memos.$.updatedAt":   toUpdateMemo.UpdatedAt,
		},
	}

	if err := dbMemoCollection.FindOneAndUpdate(context.TODO(), filter, update, options).Err(); err != nil {
		return nil, core.NewServiceErrorMessage(err)
	}

	updatedMemo, err := findMemoByID(boardID, memoID)
	if err != nil {
		return nil, err
	}

	return updatedMemo, nil
}

func deleteBoard(boardID string) (int64, int64, *core.ServiceMessage) {
	id, _ := primitive.ObjectIDFromHex(boardID)

	// Delete board
	filter := bson.M{"_id": id}
	deletedBoard, err := dbMemoCollection.DeleteMany(context.TODO(), filter, nil)
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

func deleteMemo(boardID string, memoID string) (int64, *core.ServiceMessage) {
	_, err := findMemoByID(boardID, memoID)
	if err != nil {
		if err.HTTPStatus == http.StatusNotFound {
			return -1, nil
		}

		return -1, err
	}

	bID, _ := primitive.ObjectIDFromHex(boardID)
	mID, _ := primitive.ObjectIDFromHex(memoID)
	filter := bson.M{
		"_id": bID,
	}
	update := bson.M{
		"$pull": bson.M{
			"memos": bson.M{"_id": mID},
		},
	}

	if err := dbMemoCollection.FindOneAndUpdate(context.TODO(), filter, update).Err(); err != nil {
		fmt.Println(err)
		return -1, core.NewServiceErrorMessage(err)
	}

	return 1, nil
}
