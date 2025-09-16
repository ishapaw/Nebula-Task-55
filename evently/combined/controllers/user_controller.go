package controllers

import (
	"log"
	"net/http"
	"os"
	"strings"
	"combined/models"
	"combined/service"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	service service.UserService
}

func NewUserController(s service.UserService) *UserController {
	return &UserController{service: s}
}

func (uc *UserController) Register(c *gin.Context) {
	log.Println("Register endpoint called")

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Println("Failed to bind JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("Received registration request for email: %s, role: %s\n", user.Email, user.Role)

	adminSecret := c.Query("admin_secret")
	value, _ := os.LookupEnv("ADMIN_SECRET")

	if strings.ToLower(user.Role) == "admin" && adminSecret != value {
		log.Println("Invalid admin secret attempt for email:", user.Email)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid admin secret: you are not authorized to register as an admin"})
		return
	}

	if err := uc.service.Register(&user); err != nil {
		log.Println("Error registering user:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Println("User registered successfully:", user.Email)
	c.JSON(http.StatusCreated, gin.H{"message": "user registered successfully"})
}

func (uc *UserController) Login(c *gin.Context) {
	log.Println("Login endpoint called")

	var creds struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&creds); err != nil {
		log.Println("Invalid login input:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	log.Println("Login attempt for email:", creds.Email)

	token, err := uc.service.Login(creds.Email, creds.Password)
	if err != nil {
		log.Println("Login failed for email:", creds.Email, "error:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	log.Println("Login successful for email:", creds.Email)
	c.JSON(http.StatusOK, gin.H{"token": token})
}
