package routes

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"hirensavani.com/models"
)

func signUp(context *gin.Context) {
	var user models.User

	err := context.ShouldBindJSON(&user)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data."})

	}

	err = user.Save()

	fmt.Println(err)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not save user.", "err": err})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func createGroup(context *gin.Context) {
	var group models.Groups

	err := context.ShouldBindJSON(&group)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data."})

	}

	err = group.Save()

	fmt.Println(err)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not save group.", "err": err})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "group created successfully"})

}
