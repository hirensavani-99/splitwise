package routes

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"hirensavani.com/models"
)

// func getExpense(context, *gin.Context) {

// }

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
