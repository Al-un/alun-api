package memo

import (
	"encoding/json"
	"net/http"

	"github.com/Al-un/alun-api/alun/core"
)

func handleListBoards(w http.ResponseWriter, r *http.Request, claims core.JwtClaims) {
	boards, err := findBoardsByUserID(claims.UserID)
	if err != nil {
		err.Write(w, r)
		return
	}

	if len(boards) > 0 {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
	json.NewEncoder(w).Encode(boards)
}

func handleGetBoard(w http.ResponseWriter, r *http.Request, claims core.JwtClaims) {
	boardID := core.GetVar(r, "boardId")
	board, err := findBoardByID(boardID)
	if err != nil {
		err.Write(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(board)
}

func handleCreateBoard(w http.ResponseWriter, r *http.Request, claims core.JwtClaims) {
	var toCreateBoard Board
	json.NewDecoder(r.Body).Decode(&toCreateBoard)
	toCreateBoard.PrepareForCreate(claims)

	newBoard, err := createBoard(toCreateBoard)
	if err != nil {
		err.Write(w, r)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(newBoard)
}

func handleUpdateBoard(w http.ResponseWriter, r *http.Request, claims core.JwtClaims) {
	boardID := core.GetVar(r, "boardId")
	var toUpdateBoard Board
	json.NewDecoder(r.Body).Decode(&toUpdateBoard)
	toUpdateBoard.PrepareForUpdate(claims)

	updatedBoard, err := updateBoard(boardID, toUpdateBoard)
	if err != nil {
		err.Write(w, r)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedBoard)
}

func handleDeleteBoard(w http.ResponseWriter, r *http.Request, claims core.JwtClaims) {
	boardID := core.GetVar(r, "boardId")
	deletedBoardCount, _, err := deleteBoard(boardID)

	if err != nil {
		err.Write(w, r)
		return
	}

	if deletedBoardCount > 0 {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func handleCreateMemo(w http.ResponseWriter, r *http.Request, claims core.JwtClaims) {
	boardID := core.GetVar(r, "boardId")

	var toCreateMemo Memo
	json.NewDecoder(r.Body).Decode(&toCreateMemo)

	// memoLogger.Verbose("Parsed new memo: %v", toCreateMemo)
	toCreateMemo.PrepareForCreate(claims)
	// memoLogger.Verbose("Prepared memo %v for creation with %v", toCreateMemo, claims)

	newMemo, err := createMemo(boardID, toCreateMemo)
	if err != nil {
		err.Write(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(newMemo)
}

func handleUpdateMemo(w http.ResponseWriter, r *http.Request, claims core.JwtClaims) {
	boardID := core.GetVar(r, "boardId")
	memoID := core.GetVar(r, "memoId")
	var toUpdateMemo Memo
	json.NewDecoder(r.Body).Decode(&toUpdateMemo)

	toUpdateMemo.PrepareForUpdate(claims)

	newMemo, err := updateMemo(boardID, memoID, toUpdateMemo)
	if err != nil {
		err.Write(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(newMemo)
}

func handleDeleteMemo(w http.ResponseWriter, r *http.Request, claims core.JwtClaims) {
	boardID := core.GetVar(r, "boardId")
	memoID := core.GetVar(r, "memoId")
	deleteCount, err := deleteMemo(boardID, memoID)

	if err != nil {
		err.Write(w, r)
		return
	}

	if deleteCount > 0 {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
