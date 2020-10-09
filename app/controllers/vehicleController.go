package controllers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	model "github.com/karokojnr/vepa-server-gin/app/models"
	"github.com/karokojnr/vepa-server-gin/app/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
)


func AddVehicleHandler(c *gin.Context) {
	ctx := context.TODO()
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Cannot get vehicle collection",
		})
		return
	}
	userID := c.Param("id")

	vehicle := model.Vehicle{}
	err = c.Bind(&vehicle)
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error getting body",
		})
		return
	}
	var result model.Vehicle
	err = vehicleCollection.FindOne(ctx, bson.M{"registrationNumber": vehicle.RegistrationNumber}).Decode(&result)
	if err != nil {
		vehicle.VeicleID = primitive.NewObjectID()
		vehicle.UserID = userID
		vehicle.IsWaitingClamp = false
		vehicle.IsClamped = false
		_, err = vehicleCollection.InsertOne(ctx, vehicle)
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
	ctx := context.TODO()
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Cannot get vehicle collection",
		})
		return
	}

	vehicle := model.Vehicle{}
	vehicleReg := c.Param("vehicleReg")

	err = vehicleCollection.FindOne(ctx, bson.M{"registrationNumber": vehicleReg}).Decode(&vehicle)

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
	ctx := context.TODO()
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Cannot get vehicle collection",
		})
		return
	}

	vehicle := model.Vehicle{}
	vehicleID := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(vehicleID)
	err = c.Bind(&vehicle)

	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Getting Body",
		})
		return
	}
	vehicle.VeicleID = id
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"registrationNumber": vehicle.RegistrationNumber,
		"vehicleClass":       vehicle.VehicleClass,
	}}
	var result model.Vehicle
	err = vehicleCollection.FindOneAndUpdate(ctx, filter, update).Decode(&result)
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
	var results []*model.Vehicle
	ctx := context.TODO()
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Cannot get vehicle collection",
		})
		return
	}

	userID := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"userId": id}
	cur, err := vehicleCollection.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	for cur.Next(context.TODO()) {
		var elem model.Vehicle
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, &elem)
	}
	if err := cur.Err(); err != nil {
		c.JSON(200, gin.H{
			"message": err,
		})
		return
	}
	_ = cur.Close(context.TODO())
	c.JSON(200, gin.H{
		"vehicles": &results,
	})
	return
}
func DeleteVehicleHandler(c *gin.Context) {
	ctx := context.TODO()
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Cannot get vehicle collection",
		})
		return
	}
	vehicleID := c.Param("id") // Get Param
	id, _ := primitive.ObjectIDFromHex(vehicleID)
	filter := bson.M{"_id": id}
	var result model.Vehicle
	err = vehicleCollection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		log.Println(err)
	}
	util.Log("Vehicle to be deleted found - ", result.RegistrationNumber)
	if result.IsClamped == false {
		deleteResult, err := vehicleCollection.DeleteOne(ctx, filter)
		if err != nil {
			fmt.Println(err)
			return
		}
		util.Log("Deleted Vehicle")
		c.JSON(200, gin.H{
			"delete result": deleteResult,
		})
		return
	}
	util.Log("Vehicle Clamped! Deletion not allowed")
	c.JSON(200, gin.H{
		"message": "clamped",
	})
	return

}
func VehiclesWaitingClamp(c *gin.Context) {
	var vehicles []*model.Vehicle
	ctx := context.TODO()
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Cannot get vehicle collection",
		})
		return
	}
	vehicleFilter := bson.M{
		"isWaitingClamp": true,
		//"isClamped":      false,
	}
	cur, err := vehicleCollection.Find(ctx, vehicleFilter)
	if err != nil {
		log.Println(err)
	}
	for cur.Next(context.TODO()) {
		var elem model.Vehicle
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		vehicles = append(vehicles, &elem)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	_ = cur.Close(context.TODO())
	c.JSON(200, gin.H{
		"vehicles": vehicles,
	})
	return

}
func ClampedVehicle(c *gin.Context) {
	var vehicles []*model.Vehicle
	ctx := context.TODO()
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Cannot get vehicle collection",
		})
		return
	}
	vehicleFilter := bson.M{
		"isWaitingClamp": false,
		"isClamped":      true,
	}
	cur, err := vehicleCollection.Find(ctx, vehicleFilter)
	if err != nil {
		log.Println(err)
	}
	for cur.Next(context.TODO()) {
		var elem model.Vehicle
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		vehicles = append(vehicles, &elem)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	_ = cur.Close(context.TODO())
	c.JSON(200, gin.H{
		"vehicles": vehicles,
	})
	return
}
func CheckIfVehicleIsClampedHandler(c *gin.Context) {
	ctx := context.TODO()
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Cannot get vehicle collection",
		})
		return
	}
	vehicleReg := c.Param("vehicleReg")
	var vehicleModel model.Vehicle
	vehicleFilter := bson.M{"registrationNumber": vehicleReg, "isClamped": true}
	err = vehicleCollection.FindOne(ctx, vehicleFilter).Decode(&vehicleModel)
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
