package controllers

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	model "github.com/karokojnr/vepa-server-gin/app/models"
	"github.com/karokojnr/vepa-server-gin/app/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(c *gin.Context) {
	ctx := context.TODO()
	userCollection, err := util.GetCollection("users")

	if err != nil {
		c.JSON(200, gin.H{
			"error": "Cannot get user collection",
		})
		return
	}
	user := model.User{}
	err = c.Bind(&user)
	if err != nil {
		c.JSON(200, gin.H{
			"error": "Error Getting Body",
		})
		return
	}

	var result model.User
	err = userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&result)
	if err != nil {

		hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 5)

		if err != nil {

			c.JSON(200, gin.H{
				"error": "Error While Hashing Password, Try Again Later",
			})
			return
		}
		user.Password = string(hash)
		user.ID = primitive.NewObjectID()
		_, err = userCollection.InsertOne(ctx, &user)

		if err != nil {
			c.JSON(403, gin.H{
				"error": "Error Inserting User",
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
			c.JSON(403, gin.H{"error": "Error while generating token,Try again"})
			c.Abort()
			return
		}

		user.Token = tokenString
		user.Password = ""
		user.Exp = exp

		c.JSON(200, gin.H{
			"message": "Success Insertting User",
			"user":    &user,
		})
		return
	}
	c.JSON(403, gin.H{"error": "Email already exists!!!"})
	c.Abort()
	return
}
func LoginHandler(c *gin.Context) {
	ctx := context.TODO()
	userCollection, err := util.GetCollection("users")

	if err != nil {
		c.JSON(200, gin.H{
			"error": "Cannot get user collection",
		})
		return
	}
	user := model.User{}
	err = c.Bind(&user)
	if err != nil {
		c.JSON(200, gin.H{
			"error": "Error Get Body",
		})
		return
	}
	var result model.User
	err = userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&result)
	if err != nil {
		c.JSON(404, gin.H{"error": "User account was not found"})
		c.Abort()
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(user.Password))
	if err != nil {
		c.JSON(404, gin.H{"error": "Invalid password"})
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
		c.JSON(404, gin.H{"error": "Error while generating token"})
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
	ctx := context.TODO()
	userCollection, err := util.GetCollection("users")

	if err != nil {
		c.JSON(200, gin.H{
			"message": "Cannot get user collection",
		})
		return
	}

	user := model.User{}
	userID := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(userID)
	tokenString := c.GetHeader("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte("secret"), nil
	})
	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		err = userCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
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
	c.JSON(403, gin.H{
		"error": "You are not Authorized!",
	})
	return
}
func EditProfile(c *gin.Context) {
	ctx := context.TODO()
	userCollection, err := util.GetCollection("users")

	if err != nil {
		c.JSON(200, gin.H{
			"message": "Cannot get user collection",
		})
		return
	}
	tokenString := c.GetHeader("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte("secret"), nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id := claims["id"].(string)
		userID, _ := primitive.ObjectIDFromHex(id)
		user := model.User{}
		err = c.Bind(&user)
		if err != nil {
			c.JSON(200, gin.H{
				"message": "Error Getting Body",
			})
			return
		}
		update := bson.M{"$set": bson.M{
			"firstName":   user.Firstname,
			"lastName":    user.Lastname,
			"email":       user.Email,
			"idNumber":    user.IDNumber,
			"phoneNumber": user.PhoneNumber,
		}}
		var result model.User
		err = userCollection.FindOneAndUpdate(ctx, bson.M{"_id": userID}, update).Decode(&result)
		if err != nil {
			c.JSON(200, gin.H{
				"error": "Error Updating User",
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "Success Updating User",
			"user":    &user,
		})
		return
	}
	c.JSON(403, gin.H{
		"error": "You are not Authorized!",
	})
	return
}
func FCMTokenHandler(c *gin.Context) {
	ctx := context.TODO()
	userCollection, err := util.GetCollection("users")

	if err != nil {
		c.JSON(200, gin.H{
			"message": "Cannot get user collection",
		})
		return
	}

	user := model.User{}
	userID := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(userID)
	err = c.Bind(&user)

	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Getting Body",
		})
		return
	}
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"fcmtoken": user.FCMToken}}
	_, err = userCollection.UpdateOne(ctx, filter, update)
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
