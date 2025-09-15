package controllers

import (
	"net/http"
	"os"
	"strings"
	"users/models"
	"users/service"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	service service.UserService
}

func NewUserController(s service.UserService) *UserController {
	return &UserController{service: s}
}

func (uc *UserController) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminSecret := c.Query("admin_secret")
	value, _ := os.LookupEnv("ADMIN_SECRET")

	if strings.ToLower(user.Role) == "admin" && adminSecret != value {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid admin secret: you are not authorized to register as an admin"})
		return
	}

	if err := uc.service.Register(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user registered successfully"})
}

func (uc *UserController) Login(c *gin.Context) {
	var creds struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	token, err := uc.service.Login(creds.Email, creds.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
