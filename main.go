package main

import (
	"github.com/gin-gonic/gin"
	"hirensavani.com/db"
	"hirensavani.com/routes"
)

func main() {

	db.InitDB()

	defer db.DB.Close()

	server := gin.Default()

	routes.RegisterRoutes(server)

	server.Run(":8080")
}
