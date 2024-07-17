package routes

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"hirensavani.com/db"
	"hirensavani.com/models"
)

func AddComments(context *gin.Context) {

	expenseId, err := strconv.ParseInt(context.Param("expense_Id"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid expense id"})
		return
	}

	var comment models.Comment
	if err := context.ShouldBindJSON(&comment); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse comment data."})
		return
	}

	comment.ExpenseID = expenseId

	err = comment.Save(db.DB)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not add comments for expense.", "err": err.Error()})
		return
	}

}
