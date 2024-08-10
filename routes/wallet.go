package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"hirensavani.com/db"
	"hirensavani.com/models"
)

func getWalletById(context *gin.Context) {
	userId, err := strconv.ParseInt(context.Param("userId"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid userId"})
		return
	}
	wallet := &models.Wallet{}
	err = wallet.Get(db.DB, userId)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "issue returning wallet", "err": err})
		return
	}

	context.JSON(http.StatusOK, gin.H{"wallet": wallet})
}

func SettledUpWallet(context *gin.Context) {

	fmt.Println("12")
	var settlement models.SettlementType

	err := context.ShouldBindJSON(&settlement)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data."})
		return
	}

	err = settlement.SettleUpWallet(db.DB)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "could not settled up wallet", "err": err})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Wallet Settled up !"})

}
