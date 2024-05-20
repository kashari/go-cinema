package handler

import (
	"go-cinema/extras"
	"go-cinema/model"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Login(c *gin.Context) {
	var user model.User
	var loginRequest model.LoginRequest
	DB := c.MustGet("db").(*gorm.DB)

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if err := DB.Where("username = ?", loginRequest.Username).First(&user).Error; err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect password"})
		return
	}

	accessToken, err := extras.CreateToken(&user, time.Minute*5)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	refreshToken, err := extras.CreateToken(&user, time.Hour*24)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	loginResponse := model.LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken}

	c.JSON(http.StatusOK, loginResponse)
}

func RefreshToken(c *gin.Context) {
	var refreshTokenRequest extras.RefreshTokenRequest
	DB := c.MustGet("db").(*gorm.DB)
	if err := c.ShouldBindJSON(&refreshTokenRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	claims, err := extras.VerifyToken(refreshTokenRequest.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	var user model.User
	if err := DB.Where("username = ?", claims["username"].(string)).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	accessToken, err := extras.CreateToken(&user, time.Minute*5)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}

func CreateUser(c *gin.Context) {
	var userRequest model.UserRequest
	DB := c.MustGet("db").(*gorm.DB)
	if err := c.ShouldBindJSON(&userRequest); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	log.Println(userRequest)

	hash, err := bcrypt.GenerateFromPassword([]byte(userRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to hash password"})
		return
	}

	userRequest.Password = string(hash)

	var existingUser model.User
	if err := DB.Where("username = ? OR email = ?", userRequest.Username, userRequest.Email).First(&existingUser).Error; err == nil {
		c.JSON(409, gin.H{"error": "Username or email already exists"})
		return
	}

	user := model.User{Username: userRequest.Username, Email: userRequest.Email, Password: userRequest.Password}

	if err := DB.Create(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(201, gin.H{"message": "User created successfully"})
}

func GetUser(c *gin.Context) {
	userID := c.Param("id")
	DB := c.MustGet("db").(*gorm.DB)

	var user model.User

	if err := DB.First(&user, userID).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	c.JSON(200, user)
}

func GetAllUsers(c *gin.Context) {
	var users []model.User
	DB := c.MustGet("db").(*gorm.DB)
	if err := DB.Find(&users).Limit(50).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve users"})
		return
	}
	c.JSON(200, users)
}

func UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	DB := c.MustGet("db").(*gorm.DB)

	var user model.User

	if err := DB.First(&user, userID).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	var userRequest model.UserRequest
	if err := c.ShouldBindJSON(&userRequest); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Update only the provided fields
	if userRequest.Username != "" {
		user.Username = userRequest.Username
	}
	if userRequest.Email != "" {
		user.Email = userRequest.Email
	}

	if userRequest.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(userRequest.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to hash password"})
			return
		}
		user.Password = string(hash)
	}

	if userRequest.IsEmpty() {
		c.JSON(301, "No fields to update!")
	}
	userRequest.ID = userID
	user.Username = userRequest.Username
	user.Email = userRequest.Email
	user.Password = userRequest.Password

	if err := DB.Save(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(200, "User updated successfully!")
}

func DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	DB := c.MustGet("db").(*gorm.DB)

	var user model.User

	if err := DB.First(&user, userID).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	if err := DB.Delete(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(200, gin.H{"message": "User deleted successfully"})
}
