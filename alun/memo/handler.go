package memo

import (
	"encoding/json"
	"net/http"

	"github.com/Al-un/alun-api/alun/core"
)

func handleCreateMemo(w http.ResponseWriter, r *http.Request, claims core.JwtClaims) {
	var toCreateMemo Memo
	json.NewDecoder(r.Body).Decode(&toCreateMemo)

	memoLogger.Verbose("Parsed new memo: %v", toCreateMemo)

	toCreateMemo.PrepareForCreate(claims)

	newMemo, err := createMemo(toCreateMemo)
	if err != nil {
		core.HandleServerError(w, r, err)
		return
	}

	json.NewEncoder(w).Encode(newMemo)
}

func handleListMemo(w http.ResponseWriter, r *http.Request, claims core.JwtClaims) {
	memoLogger.Verbose("Loading memos for %v", claims.UserID)
	memos, err := findMemosByUserID(claims.UserID)
	if err != nil {
		core.HandleServerError(w, r, err)
		return
	}

	json.NewEncoder(w).Encode(memos)
}

func handleGetMemo(w http.ResponseWriter, r *http.Request, claims core.JwtClaims) {
	memoID := core.GetVar(r, "memoId")
	memo, err := findMemoByID(memoID)
	if err != nil {
		core.HandleServerError(w, r, err)
		return
	}

	json.NewEncoder(w).Encode(memo)
}

func handleUpdateMemo(w http.ResponseWriter, r *http.Request, claims core.JwtClaims) {
	memoID := core.GetVar(r, "memoId")
	var toUpdateMemo Memo
	json.NewDecoder(r.Body).Decode(&toUpdateMemo)

	toUpdateMemo.PrepareForUpdate(claims)

	newMemo, err := updateMemo(memoID, toUpdateMemo)
	if err != nil {
		core.HandleServerError(w, r, err)
		return
	}

	json.NewEncoder(w).Encode(newMemo)
}

func handleDeleteMemo(w http.ResponseWriter, r *http.Request, claims core.JwtClaims) {
	memoID := core.GetVar(r, "memoId")
	deleteCount := deleteMemo(memoID)

	if deleteCount > 0 {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
