package controllers

import (
	"net/http"

	"github.com/aoncodev/qrbackend/initializers"
	"github.com/aoncodev/qrbackend/models"
	"github.com/aoncodev/qrbackend/utils"
	"github.com/gin-gonic/gin"
)

func AdminLogin(c *gin.Context) {
	var body struct {
		OTP string `json:"otp" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var admin models.Employee
	// Ensure OTP is compared as a string in the database
	if err := initializers.DB.
		Where("CAST(otp AS TEXT) = ? AND role = ?", body.OTP, "admin").
		First(&admin).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP or not an admin"})
		return
	}

	accessToken, err := utils.GenerateJWT(admin.ID, admin.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	// Optional: Clear OTP after use
	// initializers.DB.Model(&admin).Update("otp", "")

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"user": gin.H{
			"id":   admin.ID,
			"name": admin.Name,
			"role": admin.Role,
		},
	})
}

func GetAllEmployees(c *gin.Context) {
	var employees []models.Employee
	if err := initializers.DB.Find(&employees).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve employees"})
		return
	}
	c.JSON(http.StatusOK, employees)
}