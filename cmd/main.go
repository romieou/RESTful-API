package main

import (
	"log"
	"rest/controllers"
	"rest/models/mysql"
	"rest/models/redis"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

func main() {
	db, err := mysql.NewMySQL()
	if err != nil {
		log.Println(err)
		return
	}
	redis, err := redis.NewRedisCache(0)
	if err != nil {
		log.Println(err)
		return
	}
	server := controllers.NewMyServer(db, redis)
	go server.DispatchWorkers()
	// r := routes.NewRouter(server)
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
	fasthttp.ListenAndServe(":8080", r.Handler)

}
