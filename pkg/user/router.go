package user

import "github.com/gorilla/mux"

// SetupRoutes for user module
func SetupRoutes(router *mux.Router) {
	router.HandleFunc("/v1/users/auth", authUser).Methods("POST")
	router.HandleFunc("/v1/users/register", registerUser).Methods("POST")
	router.Handle("/v1/users/detail/{userId}", IsAuthorized(handleGetUser, AuthCheckIsAdminOrUser(""))).Methods("GET")
	router.Handle("/v1/users/detail/{userId}", IsAuthorized(handleUpdateUser, AuthCheckIsAdminOrUser(""))).Methods("PUT")
	router.Handle("/v1/users/detail/{userId}", IsAuthorized(handleDeleteUser, AuthCheckIsAdminOrUser(""))).Methods("DELETE")
}
