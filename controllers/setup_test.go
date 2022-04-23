package controllers

import (
	"fmt"
	"rest/models"
	"rest/myerrors"
	"strconv"

	"github.com/buaazp/fasthttprouter"
)

type testDB struct{}

func (db *testDB) CreateUser(u *models.User) (int64, error) {
	fmt.Println("ass")
	return 0, nil
}

func (db *testDB) GetUser(ID string) (*models.User, error) {
	fmt.Println("asas")
	return nil, nil
}

func (db *testDB) UpdateUser(ID string, u models.User) error {
	return nil
}

func (db *testDB) DeleteUser(ID string) error {
	return nil
}

type testRedis struct{}

func (r *testRedis) GetCounter() (string, error) {
	return "0", nil
}

func (r *testRedis) SetCounter(n int) (string, error) {
	// testing add
	if n == 0 {
		return "0", nil
	}
	if n == 1 {
		return "1", nil
	}
	if n == -1 {
		return "0", nil
	}
	if n == 2 {
		return "0", fmt.Errorf("some error")
	}
	// testing sub
	if n == -3 {
		return "2", nil
	}

	if n == -1234567 {
		return "0", myerrors.ErrNegativeCounter
	}
	if n < 0 {
		return "0", fmt.Errorf("some error")
	}
	return strconv.Itoa(n), nil
}

func (r *testRedis) Set(key string, val interface{}) error {
	return nil
}

func (r *testRedis) Get(key string) (string, error) {
	return "", nil
}

// NewRouter returns fasthttprouter.Router for supported routes
func NewRouter(server *MyServer) *fasthttprouter.Router {
	r := fasthttprouter.New()
	r.GET("/rest/substr", server.SubstringHandler)
	r.POST("/rest/substr/find", server.GetSubstring)
	r.GET("/rest/email", server.EmailHandler)
	r.POST("/rest/email/check", server.GetEmail)
	r.POST("/rest/iin/check", server.GetIIN)
	r.POST("/rest/counter/add/:add", server.AddCounter)
	r.POST("/rest/counter/sub/:sub", server.SubCounter)
	r.GET("/rest/counter/val", server.GetCounter)
	r.POST("/rest/user", server.CreateUser)
	r.GET("/rest/user/:id", server.GetUser)
	r.PUT("/rest/user/:id", server.UpdateUser)
	r.DELETE("/rest/user/:id", server.DeleteUser)
	r.POST("/rest/hash/calc", server.GenerateHash)
	r.GET("/rest/hash/result/:id", server.GetHash)
	r.GET("/rest/hash", server.HashHandler)
	r.GET("/rest/self/find/:str", server.GetIdentifiers)
	return r
}
