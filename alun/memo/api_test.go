package memo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Al-un/alun-api/alun/testutils"
)

func TestE2EMemo(t *testing.T) {
	// Config
	t.Parallel()

	// Setup
	var testInfo testutils.APITestInfo
	_, token := setupUser(t)

	board1Title := "Board #1"
	board1TitleNew := "The best board"
	board1 := Board{
		BasicInfo: BasicInfo{Title: board1Title, Description: "Description of the first board"},
		Access:    accessPrivate,
	}

	memo1ItemsSet1 := []Item{
		{Text: "Item 1", IsFinished: false},
		{Text: "Item 2"},
	}
	// Item 2 has been deleted and item 4 and 5 are added
	memo1ItemsSet2 := []Item{
		{Text: "Item 1", IsFinished: true},
		{Text: "Item 4"},
		{Text: "Item 5", IsFinished: true},
	}
	memo2ItemsSet := []Item{
		{Text: "Plop 1", DueDate: time.Now().Add(42 * time.Minute)},
		{Text: "Plop 2", DueDate: time.Now().Add(72 * time.Hour), IsFinished: true},
	}

	memo1 := Memo{
		BasicInfo: BasicInfo{Title: "Memo #1", Description: "Description of the first memo"},
		Items:     memo1ItemsSet1,
	}
	memo2 := Memo{
		BasicInfo: BasicInfo{Title: "Memo #2", Description: "Description of the second memo"},
		Items:     memo2ItemsSet,
	}

	// Cleanup
	t.Cleanup(func() {
		tearDownUser(t)
		deleteBoard(board1.ID.Hex())
	})

	// Tests
	t.Run("LoadInitialEmptyBoardList", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:               "boards",
			Method:             http.MethodGet,
			Payload:            nil,
			ExpectedHTTPStatus: http.StatusNoContent,
			AuthToken:          token,
		}
		rr := apiTester.TestPath(t, testInfo)

		var boards []Board
		json.NewDecoder(rr.Body).Decode(&boards)
		testutils.Equals(t, testutils.CallFromTestFile, make([]Board, 0), boards)
	})

	t.Run("CreateFirstBoard", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:               "boards",
			Method:             http.MethodPost,
			Payload:            board1,
			ExpectedHTTPStatus: http.StatusOK,
			AuthToken:          token,
		}
		rr := apiTester.TestPath(t, testInfo)

		// Check response
		var newBoard Board
		json.NewDecoder(rr.Body).Decode(&newBoard)

		testutils.Equals(t, testutils.CallFromTestFile, board1.BasicInfo, newBoard.BasicInfo)
		testutils.Equals(t, testutils.CallFromTestFile, board1.Access, newBoard.Access)
		testutils.Assert(t, testutils.CallFromTestFile, !newBoard.CreatedAt.IsZero(), "CreatedAt is empty")
		testutils.Assert(t, testutils.CallFromTestFile, newBoard.ModifiedAt.IsZero(), "ModifiedAt is not empty %v", newBoard.ModifiedAt)

		// Check in DB
		boardFromDb, _ := findBoardByID(newBoard.ID.Hex())
		testutils.Equals(t, testutils.CallFromTestFile, board1.BasicInfo, boardFromDb.BasicInfo)
		testutils.Equals(t, testutils.CallFromTestFile, board1.Access, boardFromDb.Access)

		// Save ID for teardown
		board1.ID = boardFromDb.ID
		board1.TrackedEntity = boardFromDb.TrackedEntity
	})

	t.Run("CreateFirstMemo", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:               fmt.Sprintf("boards/%s/memos", board1.ID.Hex()),
			Method:             http.MethodPost,
			Payload:            memo1,
			ExpectedHTTPStatus: http.StatusOK,
			AuthToken:          token,
		}
		rr := apiTester.TestPath(t, testInfo)

		// Check response
		var newMemo Memo
		json.NewDecoder(rr.Body).Decode(&newMemo)

		testutils.Equals(t, testutils.CallFromTestFile, memo1.BasicInfo, newMemo.BasicInfo)
		testutils.Equals(t, testutils.CallFromTestFile, memo1.Items, newMemo.Items)
		testutils.Assert(t, testutils.CallFromTestFile, !newMemo.CreatedAt.IsZero(), "CreatedAt is empty")
		testutils.Assert(t, testutils.CallFromTestFile, newMemo.ModifiedAt.IsZero(), "ModifiedAt is not empty %v", memo1.ModifiedAt)

		// Check in DB
		memoFromDb, _ := findMemoByID(board1.ID.Hex(), newMemo.ID.Hex())
		testutils.Equals(t, testutils.CallFromTestFile, memo1.BasicInfo, memoFromDb.BasicInfo)
		testutils.Equals(t, testutils.CallFromTestFile, memo1.Items, memoFromDb.Items)

		// Save ID for teardown
		memo1.ID = memoFromDb.ID
		memo1.TrackedEntity = memoFromDb.TrackedEntity
	})

	t.Run("UpdateFirstBoard", func(t *testing.T) {
		board1.Title = board1TitleNew
		testInfo = testutils.APITestInfo{
			Path:               fmt.Sprintf("boards/%s", board1.ID.Hex()),
			Method:             http.MethodPut,
			Payload:            board1,
			ExpectedHTTPStatus: http.StatusOK,
			AuthToken:          token,
		}
		rr := apiTester.TestPath(t, testInfo)

		// Check response
		var newBoard Board
		json.NewDecoder(rr.Body).Decode(&newBoard)

		testutils.Equals(t, testutils.CallFromTestFile, board1TitleNew, newBoard.Title)
		testutils.Equals(t, testutils.CallFromTestFile, board1.CreatedAt, newBoard.CreatedAt)
		testutils.Assert(t, testutils.CallFromTestFile, !newBoard.ModifiedAt.IsZero(), "ModifiedAt is empty")

		// Check in DB
		boardFromDb, _ := findBoardByID(newBoard.ID.Hex())
		testutils.Equals(t, testutils.CallFromTestFile, board1TitleNew, boardFromDb.Title)
		testutils.Equals(t, testutils.CallFromTestFile, newBoard.ModifiedAt, boardFromDb.ModifiedAt)

		// Save Data
		board1.ModifiedAt = boardFromDb.ModifiedAt
		board1.ModifiedBy = boardFromDb.ModifiedBy
	})

	t.Run("LoadListWhichShouldNotHaveMemos", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:               "boards",
			Method:             http.MethodGet,
			Payload:            nil,
			ExpectedHTTPStatus: http.StatusOK,
			AuthToken:          token,
		}
		rr := apiTester.TestPath(t, testInfo)

		var boards []Board
		json.NewDecoder(rr.Body).Decode(&boards)

		testutils.Assert(t, testutils.CallFromTestFile, len(boards) == 1, "There is not one board")
		b := boards[0]

		testutils.Equals(t, testutils.CallFromTestFile, b.ID, board1.ID)
		testutils.Equals(t, testutils.CallFromTestFile, b.BasicInfo, board1.BasicInfo)
		testutils.Equals(t, testutils.CallFromTestFile, b.Access, board1.Access)
		testutils.Equals(t, testutils.CallFromTestFile, b.TrackedEntity, board1.TrackedEntity)
		testutils.Equals(t, testutils.CallFromTestFile, b.Memos, []Memo(nil))
	})

	t.Run("CreateSecondMemo", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:               fmt.Sprintf("boards/%s/memos", board1.ID.Hex()),
			Method:             http.MethodPost,
			Payload:            memo2,
			ExpectedHTTPStatus: http.StatusOK,
			AuthToken:          token,
		}
		rr := apiTester.TestPath(t, testInfo)

		// Check response
		var newMemo Memo
		json.NewDecoder(rr.Body).Decode(&newMemo)

		testutils.Equals(t, testutils.CallFromTestFile, memo2.BasicInfo, newMemo.BasicInfo)
		// time precision is not the same in Go and in the DB. need to check each field
		testutils.Equals(t, testutils.CallFromTestFile, len(memo2.Items), len(newMemo.Items))
		testutils.Assert(t, testutils.CallFromTestFile, areItemsArrayEquals(memo2.Items, newMemo.Items),
			"Items are not equals: \ngot:\n%+v\nexpected:\n%+v",
			newMemo.Items, memo2.Items)
		testutils.Assert(t, testutils.CallFromTestFile, !newMemo.CreatedAt.IsZero(), "CreatedAt is empty")
		testutils.Assert(t, testutils.CallFromTestFile, newMemo.ModifiedAt.IsZero(), "ModifiedAt is not empty %v", memo1.ModifiedAt)

		// Check in DB
		memoFromDb, _ := findMemoByID(board1.ID.Hex(), newMemo.ID.Hex())
		// Save ID for teardown
		memo2.ID = memoFromDb.ID
		memo2.TrackedEntity = memoFromDb.TrackedEntity
	})

	t.Run("UpdateFirstMemo", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:   fmt.Sprintf("boards/%s/memos/%s", board1.ID.Hex(), memo1.ID.Hex()),
			Method: http.MethodPut,
			Payload: Memo{
				BasicInfo: memo1.BasicInfo,
				Items:     memo1ItemsSet2,
			},
			ExpectedHTTPStatus: http.StatusOK,
			AuthToken:          token,
		}
		rr := apiTester.TestPath(t, testInfo)

		// Check response
		var newMemo Memo
		json.NewDecoder(rr.Body).Decode(&newMemo)

		testutils.Equals(t, testutils.CallFromTestFile, memo1.BasicInfo, newMemo.BasicInfo)
		testutils.Equals(t, testutils.CallFromTestFile, memo1.CreatedAt, newMemo.CreatedAt)
		testutils.Equals(t, testutils.CallFromTestFile, memo1ItemsSet2, newMemo.Items)
		testutils.Assert(t, testutils.CallFromTestFile, areItemsArrayEquals(memo1ItemsSet2, newMemo.Items),
			"Items are not equals: \ngot:\n%+v\nexpected:\n%+v",
			newMemo.Items, memo1ItemsSet2)
		testutils.Assert(t, testutils.CallFromTestFile, !newMemo.ModifiedAt.IsZero(), "ModifiedAt is empty")

		// // Check in DB
		memoFromDb, _ := findMemoByID(board1.ID.Hex(), memo1.ID.Hex())
		testutils.Equals(t, testutils.CallFromTestFile, memo1.BasicInfo, memoFromDb.BasicInfo)
		testutils.Equals(t, testutils.CallFromTestFile, newMemo.ModifiedAt, memoFromDb.ModifiedAt)
		testutils.Equals(t, testutils.CallFromTestFile, memo1ItemsSet2, memoFromDb.Items)

		memo1.Items = memo1ItemsSet2
		memo1.ModifiedAt = memoFromDb.ModifiedAt
		memo1.ModifiedBy = memoFromDb.ModifiedBy
	})

	t.Run("GetFirstBoardDetails", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:               fmt.Sprintf("boards/%s", board1.ID.Hex()),
			Method:             http.MethodGet,
			Payload:            nil,
			ExpectedHTTPStatus: http.StatusOK,
			AuthToken:          token,
		}
		rr := apiTester.TestPath(t, testInfo)

		var b Board
		json.NewDecoder(rr.Body).Decode(&b)

		// Compare properties
		testutils.Equals(t, testutils.CallFromTestFile, board1.ID.Hex(), b.ID.Hex())
		testutils.Equals(t, testutils.CallFromTestFile, board1.BasicInfo, b.BasicInfo)
		testutils.Equals(t, testutils.CallFromTestFile, board1.Access, b.Access)
		testutils.Assert(t, testutils.CallFromTestFile, board1.TrackedEntity.Equals(b.TrackedEntity),
			"TrackedEntities are not equals: \ngot:\n%+v\nexpected:\n%+v",
			board1.TrackedEntity, b.TrackedEntity)

		// Compare memos
		memos := []Memo{memo1, memo2}
		testutils.Assert(t, testutils.CallFromTestFile, len(b.Memos) == len(memos),
			"Memos counts are not equals: got %d, expected %d",
			len(b.Memos), len(memos))
		for idx, m := range b.Memos {
			testutils.Assert(t, testutils.CallFromTestFile, m.equals(memos[idx]),
				"Memos are not equals: \ngot:\n%+v\nexpected:\n%+v",
				m, memos[idx])
		}
	})

	t.Run("DeleteFirstMemo", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:               fmt.Sprintf("boards/%s/memos/%s", board1.ID.Hex(), memo1.ID.Hex()),
			Method:             http.MethodDelete,
			Payload:            nil,
			ExpectedHTTPStatus: http.StatusNoContent,
			AuthToken:          token,
		}
		apiTester.TestPath(t, testInfo)
	})

	t.Run("DeleteFirstMemoAgain", func(t *testing.T) {
		t.Skip("FindXXByID to be fixed to throw a 404 error")

		testInfo = testutils.APITestInfo{
			Path:               fmt.Sprintf("boards/%s/memos/%s", board1.ID.Hex(), memo1.ID.Hex()),
			Method:             http.MethodDelete,
			Payload:            nil,
			ExpectedHTTPStatus: http.StatusNotFound,
			AuthToken:          token,
		}
		apiTester.TestPath(t, testInfo)
	})

	t.Run("GetFirstBoardWithOnlyOneMemoLeft", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:               fmt.Sprintf("boards/%s", board1.ID.Hex()),
			Method:             http.MethodGet,
			Payload:            nil,
			ExpectedHTTPStatus: http.StatusOK,
			AuthToken:          token,
		}
		rr := apiTester.TestPath(t, testInfo)

		var b Board
		json.NewDecoder(rr.Body).Decode(&b)

		memos := []Memo{memo2}
		testutils.Assert(t, testutils.CallFromTestFile, len(b.Memos) == len(memos),
			"Memos counts are not equals: got %d, expected %d",
			len(b.Memos), len(memos))
		for idx, m := range b.Memos {
			testutils.Assert(t, testutils.CallFromTestFile, m.equals(memos[idx]),
				"Memos are not equals: \ngot:\n%+v\nexpected:\n%+v",
				m, memos[idx])
		}
	})

	t.Run("DeleteFirstBoard", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:               fmt.Sprintf("boards/%s", board1.ID.Hex()),
			Method:             http.MethodDelete,
			Payload:            nil,
			ExpectedHTTPStatus: http.StatusNoContent,
			AuthToken:          token,
		}
		apiTester.TestPath(t, testInfo)

		isBoardIDInDb, _ := isBoardIDExist(board1.ID.Hex())
		testutils.Assert(t, testutils.CallFromTestFile, !isBoardIDInDb, "Board ID %s still in DB", board1.ID)
	})

	t.Run("DeleteFirstBoardAgain", func(t *testing.T) {
		testInfo = testutils.APITestInfo{
			Path:               fmt.Sprintf("boards/%s", board1.ID.Hex()),
			Method:             http.MethodDelete,
			Payload:            nil,
			ExpectedHTTPStatus: http.StatusNotFound,
			AuthToken:          token,
		}
		apiTester.TestPath(t, testInfo)
	})
}
