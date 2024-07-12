package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
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
type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Token    string `json:"token"`
}

var (
	mongoClient *mongo.Client
	collection  *mongo.Collection
)

var sampleSecretKey = []byte("josemycoach")

func mongoInit() {
	uri := "mongodb://localhost:27017"

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	mongoClient = client
	collection = client.Database("cloud").Collection("users")
}

func hashPassword(password string) (string, error) {
	passwordBytes := []byte(password)

	hashedPasswordBytes, err := bcrypt.
		GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)

	return string(hashedPasswordBytes), err
}

func generateJWT(user User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["id"] = user.ID
	claims["username"] = user.Username
	claims["role"] = user.Role
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString(sampleSecretKey)
	if err != nil {
		fmt.Errorf("Something Went Wrong: %s", err.Error())
		return "", err
	}
	return tokenString, nil
}

func validateToken(tokenString string) (UserResponse, error) {
	var user UserResponse
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("there was an error in parsing")
		}
		return sampleSecretKey, nil
	})
	if err != nil {
		return user, fmt.Errorf("invalid token: %v", err)
	}

	if token == nil {
		return user, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return user, errors.New("couldn't parse claims")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return user, errors.New("token does not contain expiration time")
	}

	if int64(exp) < time.Now().Unix() {
		return user, errors.New("token expired")
	}
	user.ID = claims["id"].(string)
	user.Username = claims["username"].(string)
	user.Role = claims["role"].(string)

	return user, nil
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
	// JSON Web token
	token, err := generateJWT(user)
	if err != nil {
		fmt.Println("failed to generate token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	UserResponse := UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		Token:    token,
	}
	c.JSON(http.StatusOK, gin.H{"user": UserResponse})
}

//func logoutHandler(c *gin.Context) {
//	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
//}
