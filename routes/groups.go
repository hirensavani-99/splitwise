package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"hirensavani.com/models"
)

type requestBody struct {
	UserIDs []int64 `json:"userIds" binding:"required"`
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
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid groupId"})
		return
	}

	var requestBody requestBody
	if err := context.ShouldBindJSON(&requestBody); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse request data."})
		return
	}

	err = addMembersToGroup(groupId, requestBody.UserIDs)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not add members to group.", "err": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Members added successfully"})
}


func addMembersToGroup(groupId int64, userIds []int64) error {
	group := models.Groups{ID: groupId}
	for _, userId := range userIds {
		err := group.AddMember(userId)
		if err != nil {
			return fmt.Errorf("unable to add member with ID %d to group: %w", userId, err)
		}
	}
	return nil
}
