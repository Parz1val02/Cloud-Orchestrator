package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        string    `bson:"_id"`
	Username  string    `bson:"username"`
	Password  string    `bson:"password"`
	Role      string    `bson:"role"`
	Token     string    `bson:"token"`
	LastLogin time.Time `bson:"last_login"`
}

var (
	mongoClient *mongo.Client
	collection  *mongo.Collection
)

func mongoInit() {
	uri := "mongodb://localhost:27017"

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	mongoClient = client
	collection = client.Database("cloud").Collection("users")
}

func updateUser(userID, token string, lastLogin time.Time) error {
	objID, err := primitive.ObjectIDFromHex(userID)
	fmt.Println(userID)

	if err != nil {
		fmt.Println("ObjectIDFromHex ERROR", err)
	} else {
		fmt.Println("ObjectIDFromHex:", objID)
	}
	filter := bson.M{"_id": bson.M{"$eq": objID}}

	update := bson.M{
		"$set": bson.M{
			"token":      token,
			"last_login": lastLogin,
		},
	}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func hashPassword(password string) (string, error) {
	passwordBytes := []byte(password)

	hashedPasswordBytes, err := bcrypt.
		GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)

	return string(hashedPasswordBytes), err
}

func loginHandler(c *gin.Context) {
	mongoInit()
	defer func() {
		if err := mongoClient.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	var loginReq struct {
		Username string `json:"username" binding:required`
		Password string `json:"password" binding:required`
	}

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user User
	fmt.Println(loginReq.Username)
	fmt.Println(loginReq.Password)
	//hashedPassword, err := hashPassword(loginReq.Password)
	//if err != nil {
	//	println(fmt.Println("Error hashing password"))
	//	return
	//}
	//fmt.Println(hashedPassword)
	filter := bson.M{"username": loginReq.Username}
	err := collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		fmt.Println("username mal")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password))
	if err != nil {
		fmt.Println("password mal")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	token := uuid.New().String()
	user.Token = token
	user.LastLogin = time.Now()
	err = updateUser(user.ID, user.Token, user.LastLogin)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error"})
		return
	}
	fmt.Println(user.Username)
	c.JSON(http.StatusOK, gin.H{"user": user})
}
