package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"hirensavani.com/db"
	"hirensavani.com/models"
)

func createExpense(context *gin.Context) {
	var expense models.Expense

	err := context.ShouldBindJSON(&expense)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data."})
		return
	}

	fmt.Println(expense)

	err = expense.Save()

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Error saving Expense.", "err": err})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Expense added succesfully.", "data": expense})
}

func getAllExpenses(context *gin.Context) {
	userId, err := strconv.ParseInt(context.Param("userId"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid userId"})
		return
	}

	expenses, err := models.GetAllExpense(db.DB, userId)
	if err != nil {
		fmt.Println(err)
		context.JSON(http.StatusBadRequest, gin.H{"message": "issue returning Expenses", "err": err})
		return
	}

	context.JSON(http.StatusOK, gin.H{"Expenses": expenses})
}

func updateExpense(context *gin.Context) {
	ExpenseId, err := strconv.ParseInt(context.Param("expenseId"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ExpenseId"})
		return
	}

	var expense map[string]interface{}

	err = context.ShouldBindJSON(&expense)

	fmt.Println("--->", expense)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data."})
	}

	err = models.UpdateExpense(db.DB, expense, ExpenseId)

	if err != nil {
		fmt.Println(err)
		context.JSON(http.StatusBadRequest, gin.H{"message": "issue updating Expenses", "err": err})
		return
	}

	context.JSON(http.StatusOK, gin.H{"Expenses": expense})

}
