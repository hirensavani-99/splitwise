package routes

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"hirensavani.com/models"
)

type requestBody struct {
	UserID int64 `json:"userId" binding:"required"`
}

func createGroup(context *gin.Context) {
	var group models.Groups

	err := context.ShouldBindJSON(&group)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data."})
		return
	}

	err = group.Save()

	fmt.Println(err)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not save group.", "err": err})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "group created successfully"})

}

func AddMemberToGroup(context *gin.Context) {
	groupId, err := strconv.ParseInt(context.Param("group_Id"), 10, 64)

	var requestBody requestBody

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid groupId"})
		return
	}

	if err := context.ShouldBindJSON(&requestBody); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data."})
		return
	}

	group := models.Groups{ID: groupId}
	err = group.AddMember(requestBody.UserID)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not add member to group.", "err": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Member added successfully"})
}

// func AddMemberToGroup1(groupId int64, userId int64) error {

// 	group := models.Groups{ID: groupId}
// 	err := group.AddMember(userId)

// 	if err != nil {
// 		return errors.New("unable to add member in to group")
// 	}

// 	return nil
// }

// // [1,2,3,4,5]
// func AddgroupMembers(members []int) {
// 	for index, value := range members {

// 	}
// }
