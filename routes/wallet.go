package routes

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"hirensavani.com/models"
)

func getWalletById(context *gin.Context) {
	userId, err := strconv.ParseInt(context.Param("userId"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid userId"})
		return
	}

	walleteData, err := models.GetWallet(userId)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "issue returning wallet", "err": err})
		return
	}

	context.JSON(http.StatusOK, gin.H{"wallet": walleteData})
}
