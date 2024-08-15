package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"hirensavani.com/db"
	"hirensavani.com/models"
)

// CreateAccount handles the user sign up process. It creates a user and a wallet.
func CreateAccount(c *gin.Context) {
	var user models.User

	// Bind the request body to the user struct.
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request data"})
		return
	}

	// Save the user to the database.
	userID, err := user.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user"})
		return
	}

	// Create a new wallet for the user.
	wallet := models.NewWallet(userID, 0.0, "CAD")

	// Save the wallet to the database.
	if err := wallet.Save(db.DB); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create wallet"})
		return
	}

	// Return a success message.
	c.JSON(http.StatusCreated, gin.H{"message": "User and wallet created successfully"})
}
