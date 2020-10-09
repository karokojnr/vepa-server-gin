package controllers

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/karokojnr/vepa-server-gin/app/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"

	model "github.com/karokojnr/vepa-server-gin/app/models"
)

// Static Collection
const UserCollection = "users"

// Get DB from Mongo Config
func MongoConfig() *mgo.Database {
	db, err := config.GetMongoDB()
	if err != nil {
		fmt.Println(err)
	}
	return db
}

func RegisterHandler(c *gin.Context) {
	////db := *MongoConfig()
	db := c.MustGet("db").(*mgo.Database)
	fmt.Println("MONGO RUNNING...", db)
	user := model.User{}
	err := c.Bind(&user)
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Get Body",
		})
		return
	}

	result := model.User{}

	err = db.C(UserCollection).Find(bson.M{"email": user.Email}).One(&result)
	if err != nil {

		hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 5)

		if err != nil {

			c.JSON(200, gin.H{
				"message": "Error While Hashing Password, Try Again Later",
			})
			return
		}
		user.Password = string(hash)
		user.ID = primitive.NewObjectID()

		err = db.C(UserCollection).Insert(user)

		if err != nil {
			c.JSON(403, gin.H{
				"message": "Error Insert User",
			})
			c.Abort()
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"id":    user.ID,
			"email": user.Email,
		})

		tokenString, err := token.SignedString([]byte("secret"))
		exp := 60 * 60
		fmt.Println("Expires in ... seconds: ")
		fmt.Println(exp)
		if err != nil {
			c.JSON(403, gin.H{"message": "rror while generating token,Try again"})
			c.Abort()
			return
		}

		user.Token = tokenString
		user.Password = ""
		user.Exp = exp

		c.JSON(200, gin.H{
			"message": "Success Insert User",
			"user":    &user,
		})
		return
	}
	c.JSON(403, gin.H{"message": "Email is already in use"})
	c.Abort()
	return
}
func LoginHandler(c *gin.Context) {
	//db := *MongoConfig()
	db := c.MustGet("db").(*mgo.Database)

	fmt.Println("MONGO RUNNING...", db)
	user := model.User{}
	err := c.Bind(&user)
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Get Body",
		})
		return
	}
	var result model.User
	err = db.C(UserCollection).Find(bson.M{"email": user.Email}).One(&result)
	if err != nil {
		c.JSON(404, gin.H{"message": "User account was not found"})
		c.Abort()
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(user.Password))
	if err != nil {
		c.JSON(404, gin.H{"message": "Invalid password"})
		c.Abort()
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    result.ID,
		"email": result.Email,
	})

	tokenString, err := token.SignedString([]byte("secret"))
	exp := 60 * 60
	fmt.Println("Expires in ... seconds: ")
	fmt.Println(exp)
	if err != nil {
		c.JSON(404, gin.H{"message": "Error while generating token"})
		c.Abort()
		return
	}
	result.Token = tokenString
	result.Password = ""
	result.Exp = exp
	c.JSON(200, gin.H{
		"message": "Successful login",
		"user":    &result,
	})
	return

}
func ProfileHandler(c *gin.Context) {
	//db := *MongoConfig()
	db := c.MustGet("db").(*mgo.Database)

	fmt.Println("MONGO RUNNING...", db)

	user := model.User{}
	userID := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(userID)

	err := db.C(UserCollection).Find(bson.M{"_id": id}).One(&user)

	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Getting User",
		})
		return
	}

	c.JSON(200, gin.H{
		"user": &user,
	})
	return
}
func EditProfile(c *gin.Context) {
	//db := *MongoConfig()
	db := c.MustGet("db").(*mgo.Database)

	fmt.Println("MONGO RUNNING...", db)

	user := model.User{}
	userID := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(userID)
	err := c.Bind(&user)

	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Getting Body",
		})
		return
	}
	user.ID = id
	update := bson.M{"$set": bson.M{
		"firstName":   user.Firstname,
		"lastName":    user.Lastname,
		"email":       user.Email,
		"idNumber":    user.IDNumber,
		"phoneNumber": user.PhoneNumber,
	}}
	err = db.C(UserCollection).Update(bson.M{"_id": &id}, update)
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Updating User",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "Success Updating User",
		"user":    &user,
	})
	return
}
func FCMTokenHandler(c *gin.Context) {
	//db := *MongoConfig()
	db := c.MustGet("db").(*mgo.Database)

	fmt.Println("MONGO RUNNING...", db)

	user := model.User{}
	userID := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(userID)
	err := c.Bind(&user)

	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Getting Body",
		})
		return
	}
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"fcmtoken": user.FCMToken}}
	err = db.C(UserCollection).Update(filter, update)
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Updating FCMToken",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "Success Updating FCMToken",
		//"user":    &user,
	})

}
