package user

import (
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a loggable entity
// db.al_users.insertOne({username:"pouet", password:"plop"})
//
// curl http://localhost:8000/users/register --data '{"username": "plop", "password": "plop"}'
type User struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Username string             `json:"username" bson:"username"`
	Password string             `json:"-" bson:"password"` // not present in JSON: https://golang.org/pkg/encoding/json/
	IsAdmin  bool               `json:"isAdmin" bson:"isAdmin"`
}

// JwtClaims extends standard claims for our User model
type JwtClaims struct {
	IsAdmin bool `json:"isAdmin"`
	jwt.StandardClaims
}
