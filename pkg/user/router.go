package user

import "github.com/gorilla/mux"

// SetupRoutes for user module
func SetupRoutes(router *mux.Router) {
	router.HandleFunc("/users/auth", authUser).Methods("POST")
	router.HandleFunc("/users/register", registerUser).Methods("POST")
	router.Handle("/users/detail/{id}", IsAuthorized(handleGetUser, false)).Methods("GET")
	router.Handle("/users/detail/{id}", IsAuthorized(handleUpdateUser, false)).Methods("PUT")
	router.Handle("/users/detail/{id}", IsAuthorized(handleDeleteUser, false)).Methods("DELETE")
}
