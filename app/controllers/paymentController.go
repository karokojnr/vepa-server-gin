package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/AndroidStudyOpenSource/mpesa-api-go"
	"github.com/gin-gonic/gin"
	model "github.com/karokojnr/vepa-server-gin/app/models"
	"github.com/karokojnr/vepa-server-gin/app/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2"
	"log"
	"time"
)

// Static Collection
const PaymentCollection = "payments"

func PaymentHandler(c *gin.Context) {
	//db := *MongoConfig()

	db := c.MustGet("db").(*mgo.Database)
	fmt.Println("MONGO RUNNING...", db)
	payment := model.Payment{}
	userID := c.Param("id")
	//id, _ := primitive.ObjectIDFromHex(userID)
	err := c.Bind(&payment)
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Getting Body",
		})
		c.Abort()
		return
	}
	payment.UserID = userID
	payment.IsSuccessful = false
	payment.PaymentID = primitive.NewObjectID()
	err = db.C(PaymentCollection).Insert(payment)
	if err != nil {
		c.JSON(403, gin.H{
			"message": "Error Inserting Payment",
		})
		c.Abort()
		return
	}
	util.Log("Payment added successfully..")
	c.JSON(200, gin.H{
		"message": "Payment added successfully..",
		"payment": &payment,
	})
	pID := payment.PaymentID.Hex()

	//--------------------------------------------INITIALIZE STK PUSH-------------------------------------------------//
	util.Log("GetPushHandler Initialized...")
	id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": id}
	// Get user to know the USER PHONE NUMBER
	rUser := model.User{}
	err = db.C(UserCollection).Find(filter).One(&rUser)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			log.Println("User not Found!")
			c.JSON(200, gin.H{
				"message": "User not Found!",
			})
			return
		}
	}
	//Initialize STK Push
	var (
		appKey    = util.GoDotEnvVariable("MPESA_APP_KEY")
		appSecret = util.GoDotEnvVariable("MPESA_APP_SECRET")
	)
	svc, err := mpesa.New(appKey, appSecret, mpesa.SANDBOX)
	if err != nil {
		log.Println(err)
	}
	mres, err := svc.Simulation(mpesa.Express{
		BusinessShortCode: "174379",
		Password:          util.GoDotEnvVariable("MPESA_PASSWORD"),
		Timestamp:         "20200421175555",
		TransactionType:   "CustomerPayBillOnline",
		Amount:            1,
		PartyA:            rUser.PhoneNumber,
		PartyB:            "174379",
		PhoneNumber:       rUser.PhoneNumber,
		CallBackURL:       "http://34.121.65.106:3500/rcb?id=" + userID + "&paymentid=" + pID, //CallBackHandler
		AccountReference:  "Vepa",
		TransactionDesc:   "Vepa Payment",
	})
	if err != nil {
		log.Println("STK Push not sent")
	}

	var mresMap map[string]interface{}
	errm := json.Unmarshal([]byte(mres), &mresMap)
	if errm != nil {
		log.Println("Error decoding response body")
	}
	rCode := mresMap["ResponseCode"]
	rCodeString := fmt.Sprintf("%v", rCode)
	rMessage := mresMap["ResponseDescription"]
	cMessage := mresMap["CustomerMessage"]
	//log.Println(cMessage)
	util.Log(cMessage)
	// Send error message(notification) if rCode != 0
	if rCodeString == string('0') {
		//// Proceed to STK Push
		return

	}
	rMessageConv := fmt.Sprintf("%v", rMessage)
	//Send message...
	util.SendNotifications(rUser.FCMToken, rMessageConv)
	return
}

func CallBackHandler(c *gin.Context) {
	log.Println("Callback called by M-pesa...")
	util.Log("Callback called by M-pesa...")
	//var bd interface{}
	log.Println(c)

}
func UserPaymentsHandler(c *gin.Context) {
	//db := *MongoConfig()

	db := c.MustGet("db").(*mgo.Database)
	fmt.Println("MONGO RUNNING: ", db)
	userID := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(userID)

	payments := model.Payments{}
	err := db.C(PaymentCollection).Find(bson.M{"userId": id}).All(&payments)

	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Getting All User Payments",
		})
		return
	}

	c.JSON(200, gin.H{
		"payments": &payments,
	})
	return
}
func GetPaidDays(c *gin.Context) {
	//db := *MongoConfig()

	db := c.MustGet("db").(*mgo.Database)
	fmt.Println("MONGO RUNNING: ", db)
	vehicleReg := c.Param("vehicleReg")
	payments := model.Payments{}
	err := db.C(PaymentCollection).Find(bson.M{"vehicleReg": vehicleReg, "isSuccessful": true}).All(&payments)

	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Getting All User Payments",
		})
		return
	}

	c.JSON(200, gin.H{
		"payments": &payments,
	})
	return
}
func VerificationHandler(c *gin.Context) {
	//db := *MongoConfig()

	db := c.MustGet("db").(*mgo.Database)
	fmt.Println("MONGO RUNNING: ", db)
	vehicleReg := c.Param("vehicleReg")
	vehicle := model.Vehicle{}
	payment := model.Payment{}
	err := db.C(VehicleCollection).Find(bson.M{"registrationNumber": vehicleReg}).One(&vehicle)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			c.JSON(200, gin.H{
				"message": "notfound",
			})
			return
		}
	}
	currentTime := time.Now().Local()
	formatCurrentTime := currentTime.Format("2006-01-02")

	for i := range payment.Days {
		if payment.Days[i] == formatCurrentTime {
			fmt.Println("Found")
			// Found!
		}
	}
	log.Println("---Payment Days---")
	log.Println(payment.Days)
	// We create filter. If it is unnecessary to sort data for you, you can use bson.M{}
	filter := bson.M{"vehicleReg": vehicleReg, "days": formatCurrentTime, "isSuccessful": true}
	err = db.C(PaymentCollection).Find(filter).One(&payment)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			c.JSON(200, gin.H{
				"message": "unpaid",
			})
			return
		}
	}
	c.JSON(200, gin.H{
		"payment": &payment,
	})
	return
}
func UnpaidVehicleHistoryHandler(c *gin.Context) {
	//db := *MongoConfig()

	db := c.MustGet("db").(*mgo.Database)
	fmt.Println("MONGO RUNNING: ", db)
	vehicleReg := c.Param("vehicleReg")

	payments := model.Payments{}
	err := db.C(PaymentCollection).Find(bson.M{"vehicleReg": vehicleReg}).All(&payments)
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Getting All Clamped Vehicles",
		})
		return
	}

	c.JSON(200, gin.H{
		"payments": &payments,
	})
	return

}
