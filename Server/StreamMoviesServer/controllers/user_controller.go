package controllers

import (
	"context"
	"net/http"
	"time"

	database "github.com/faisal-shakeel/StreamMovies/Server/StreamMoviesServer/database"
	models "github.com/faisal-shakeel/StreamMovies/Server/StreamMoviesServer/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenConnection("users")

func HashPassword(password string) (string, error) {
	HashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}
	return string(HashPassword), nil
}

func Register() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User

		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Data"})
			return
		}
		validate := validator.New()

		if err := validate.Struct(user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
			return
		}

		hashedPassword, err := HashPassword(user.Password)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to hash password"})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing users"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}
		user.UserID = bson.NewObjectID().Hex()
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		user.Password = hashedPassword

		result, err := userCollection.InsertOne(ctx, user)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}

func LoginUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var UserLogin models.UserLogin

		if err := c.ShouldBindJSON(&UserLogin); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var FoundUser models.User

		err := userCollection.FindOne(ctx, bson.M{"email": UserLogin.Email}).Decode(&FoundUser)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Email or Password"})
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(FoundUser.Password), []byte(UserLogin.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Email or Password"})
			return
		}

	}
}
