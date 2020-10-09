package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	model "github.com/karokojnr/vepa-server-gin/app/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2"
)

// Static Collection
const VehicleCollection = "vehicles"

func AddVehicleHandler(c *gin.Context) {
	//db := *MongoConfig()
	db := c.MustGet("db").(*mgo.Database)

	fmt.Println("MONGO RUNNING...", db)
	userID := c.Param("id")

	vehicle := model.Vehicle{}
	err := c.Bind(&vehicle)
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error getting body",
		})
		return
	}
	err = db.C(VehicleCollection).Find(bson.M{"registrationNumber": vehicle.RegistrationNumber}).One(&vehicle)
	if err != nil {
		vehicle.VeicleID = primitive.NewObjectID()
		vehicle.UserID = userID
		vehicle.IsWaitingClamp = false
		vehicle.IsClamped = false
		err := db.C(VehicleCollection).Insert(vehicle)
		if err != nil {
			c.JSON(403, gin.H{
				"message": "Error Insert Vehicle",
			})
			c.Abort()
			return
		}
		c.JSON(200, gin.H{
			"message": "Success Insert Vehicle",
			"vehicle": &vehicle,
		})
		return
	}
	c.JSON(403, gin.H{
		"message": "Vehicle Already Exists",
	})
	c.Abort()
	return
}
func GetVehicleHandler(c *gin.Context) {
	//db := *MongoConfig()
	db := c.MustGet("db").(*mgo.Database)

	fmt.Println("MONGO RUNNING...", db)

	vehicle := model.Vehicle{}
	vehicleReg := c.Param("vehicleReg")

	err := db.C(VehicleCollection).Find(bson.M{"registrationNumber": vehicleReg}).One(&vehicle)

	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Getting Vehicle",
		})
		return
	}

	c.JSON(200, gin.H{
		"vehicle": &vehicle,
	})
	return
}
func EditVehicleHandler(c *gin.Context) {
	//db := *MongoConfig()
	db := c.MustGet("db").(*mgo.Database)

	fmt.Println("MONGO RUNNING...", db)

	vehicle := model.Vehicle{}
	vehicleID := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(vehicleID)
	err := c.Bind(&vehicle)

	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Getting Body",
		})
		return
	}
	vehicle.VeicleID = id
	update := bson.M{"$set": bson.M{
		"registrationNumber": vehicle.RegistrationNumber,
		"vehicleClass":       vehicle.VehicleClass,
	}}
	err = db.C(VehicleCollection).Update(bson.M{"_id": &id}, update)
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Updating Vehicle",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "Success Updating Vehicle",
		"vehicle": &vehicle,
	})
	return
}
func UserVehiclesHandler(c *gin.Context) {
	//db := *MongoConfig()
	db := c.MustGet("db").(*mgo.Database)

	fmt.Println("MONGO RUNNING...", db)

	userID := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(userID)

	vehicles := model.Vehicles{}
	err := db.C(VehicleCollection).Find(bson.M{"userId": id}).All(&vehicles)

	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Get All User Vehicle",
		})
		return
	}

	c.JSON(200, gin.H{
		"vehicles": &vehicles,
	})
	return
}
func DeleteVehicleHandler(c *gin.Context) {
	//db := *MongoConfig()
	db := c.MustGet("db").(*mgo.Database)

	fmt.Println("MONGO RUNNING: ", db)

	vehicleID := c.Param("id") // Get Param
	id, _ := primitive.ObjectIDFromHex(vehicleID)

	err := db.C(VehicleCollection).Remove(bson.M{"_id": id})
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Deleting Vehicle",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "Success Deleting Vehicle",
	})
	return
}
func VehiclesWaitingClamp(c *gin.Context) {
	//db := *MongoConfig()
	db := c.MustGet("db").(*mgo.Database)

	fmt.Println("MONGO RUNNING: ", db)

	vehicles := model.Vehicles{}
	err := db.C(VehicleCollection).Find(bson.M{"isWaitingClamp": true}).All(&vehicles)

	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Getting All Vehicles Waiting Clamp",
		})
		return
	}

	c.JSON(200, gin.H{
		"vehicles": &vehicles,
	})
	return

}
func ClampedVehicle(c *gin.Context) {
	//db := *MongoConfig()
	db := c.MustGet("db").(*mgo.Database)

	fmt.Println("MONGO RUNNING: ", db)

	vehicles := model.Vehicles{}
	err := db.C(VehicleCollection).Find(bson.M{"isClamped": true, "isWaitingClamp": false}).All(&vehicles)
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Getting All Clamped Vehicles",
		})
		return
	}

	c.JSON(200, gin.H{
		"vehicles": &vehicles,
	})
	return
}
func CheckIfVehicleIsClampedHandler(c *gin.Context) {
	//db := *MongoConfig()
	db := c.MustGet("db").(*mgo.Database)

	fmt.Println("MONGO RUNNING: ", db)
	vehicleReg := c.Param("vehicleReg")
	vehicle := model.Vehicle{}
	err := db.C(VehicleCollection).Find(bson.M{"registrationNumber": vehicleReg, "isClamped": true}).One(&vehicle)
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Getting Vehicle",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "clamped",
	})
	return
}
