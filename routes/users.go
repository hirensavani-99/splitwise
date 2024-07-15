package routes

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"hirensavani.com/db"
	"hirensavani.com/models"
)

func signUp(context *gin.Context) {
	var user models.User

	err := context.ShouldBindJSON(&user)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data."})
		return
	}

	userId, err := user.Save()

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not save user.", "err": err})
		return
	}
	res := models.NewWallet(userId, 0.0, "CAD")
	err = res.Save(db.DB)
	fmt.Println(err)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not create wallet.", "err": err})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "User & wallete are  created successfully"})
}
