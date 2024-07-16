package routes

import (
	"fmt"
	"net/http"
	"sort"
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
	}

	err = expense.Save()

	fmt.Println(err)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Error saving Expense.", "err": err})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Expense added succesfully."})
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

	sort.SliceStable(expenses, func(i, j int) bool {
		return !expenses[i].AddedAt.Before(expenses[j].AddedAt)
	})

	context.JSON(http.StatusOK, gin.H{"Expenses": expenses})
}
