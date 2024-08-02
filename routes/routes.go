package routes

import "github.com/gin-gonic/gin"

func RegisterRoutes(server *gin.Engine) {
	server.POST("/signup", signUp)
	server.POST("/createGroup", createGroup)
	server.POST("/groups/:group_Id/member", AddMemberToGroup)
	server.POST("/groups/expense/:expense_Id/comment", AddComments)
	server.POST("/groups/expense", createExpense)
	server.GET("/getWallet/:userId", getWalletById)
	server.GET("/getExpenses/:userId", getAllExpenses)
	server.PATCH("/Expense/:expenseId", updateExpense)
	server.DELETE("/Expense/:expenseId", deleteExpense)
}
