package user

import "github.com/gorilla/mux"

// SetupRoutes for user module
func SetupRoutes(router *mux.Router) {
	router.HandleFunc("/users/auth", authUser).Methods("POST")
	router.HandleFunc("/users/register", registerUser).Methods("POST")
	router.Handle("/users/detail/{userId}", IsAuthorized(handleGetUser, AuthCheckIsAdminOrUser(""))).Methods("GET")
	router.Handle("/users/detail/{userId}", IsAuthorized(handleUpdateUser, AuthCheckIsAdminOrUser(""))).Methods("PUT")
	router.Handle("/users/detail/{userId}", IsAuthorized(handleDeleteUser, AuthCheckIsAdminOrUser(""))).Methods("DELETE")
}
