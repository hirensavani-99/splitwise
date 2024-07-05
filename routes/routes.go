package routes

import "github.com/gin-gonic/gin"

func RegisterRoutes(server *gin.Engine) {
	server.POST("/signup", signUp)
	server.POST("/createGroup", createGroup)
	server.POST("/groups/:group_Id/member", AddMemberToGroup)
	server.POST("/groups/expense", createExpense)
	server.GET("/get")
}
