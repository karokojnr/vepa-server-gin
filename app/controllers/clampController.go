package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/AndroidStudyOpenSource/africastalking-go/sms"
	"github.com/AndroidStudyOpenSource/mpesa-api-go"
	"github.com/gin-gonic/gin"
	model "github.com/karokojnr/vepa-server-gin/app/models"
	"github.com/karokojnr/vepa-server-gin/app/util"
	"github.com/kyokomi/emoji"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2"
	"log"
	"time"
)

const ClampFeeCollection = "clamps"

func ClampVehicleHandler(c *gin.Context) {
	//db := *MongoConfig()
	db := c.MustGet("db").(*mgo.Database)

	fmt.Println("MONGO RUNNING: ", db)
	vehicleReg := c.Param("vehicleReg")
	vehicle := model.Vehicle{}
	filter := bson.M{"registrationNumber": vehicleReg}
	err := db.C(VehicleCollection).Find(filter).One(&vehicle)
	if err != nil {
		log.Println(err)
	}
	//Find userID to get the phone number
	uID := vehicle.UserID
	userID, _ := primitive.ObjectIDFromHex(uID)
	vID := vehicle.VeicleID

	user := model.User{}
	err = db.C(UserCollection).Find(bson.M{"_id": userID}).One(&user)
	if err != nil {
		log.Println(err)
	}
	userPhoneNumber := user.PhoneNumber
	//Test if phone number is available
	log.Println("---Phone Number---")
	log.Println(userPhoneNumber)
	util.Log("User phone number", userPhoneNumber)

	var (
		username = "karokojnr"                                        //Your Africa's Talking Username
		apiKey   = util.GoDotEnvVariable("AFRICA_IS_TALKING_API_KEY") //Production or Sandbox API Key
		env      = "production"                                       // Choose either Sandbox or Production
	)
	//Call the Gateway, and pass the constants here!
	smsService := sms.NewService(username, apiKey, env)
	plus := "+"
	vehicleModel := model.Vehicle{}
	if vehicleModel.IsWaitingClamp == true || vehicleModel.IsClamped == true {
		util.Log("vehicle is already  clamped")
		c.JSON(200, gin.H{
			"message": "vehicle is already clamped",
		})
		return
	}
	//Send SMS - REPLACE Recipient and Message with REAL Values
	smsResponse, err := smsService.Send("", plus+userPhoneNumber, "Hello, Your have not paid for your vehicle("+vehicleReg+"). It will be clamped in 30 minutes incase you don't pay. Kindly make a payment now. ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(smsResponse)
	//------------isWaitingClamp==true----------//
	if vehicleModel.IsWaitingClamp == false || vehicleModel.IsClamped == false {
		vehicleFilter := bson.M{"_id": vID}
		vehicleUpdate := bson.M{"$set": bson.M{
			"isWaitingClamp": true,
		}}
		err = db.C(VehicleCollection).Update(vehicleFilter, vehicleUpdate)
		if err != nil {
			util.Log("Error updating payment:", err.Error())
			fmt.Printf("error...")
			c.JSON(200, gin.H{
				"message": "Error Updating Vehicle",
			})
			return
		}
		util.Log("isWaitingClamp == true")
		c.JSON(200, gin.H{
			"message": "isWaitingClamp updated Successfully --> true",
		})
	}
	//Set Timer
	timerMessage := emoji.Sprint(":alarm_clock:")
	util.Log("Clamp timer started" + timerMessage)
	time.Sleep(60 * time.Second)
	util.Log("Clamp timer ended" + timerMessage)
	//Timer complete

	paymentModel := model.Payment{}
	err = db.C(VehicleCollection).Find(bson.M{"registrationNumber": vehicleReg}).One(&vehicle)
	currentTime := time.Now().Local()
	formatCurrentTime := currentTime.Format("2006-01-02")
	paymentFilter := bson.M{"vehicleReg": vehicleModel.RegistrationNumber, "days": formatCurrentTime, "isSuccessful": true}
	err = db.C(PaymentCollection).Find(paymentFilter).One(&paymentModel)
	if err != nil {
		log.Println(err)
		if vehicleModel.IsClamped == false {
			vehicleClampFilter := bson.M{"_id": vID}
			vehicleClampUpdate := bson.M{"$set": bson.M{
				"isClamped":      true,
				"isWaitingClamp": false,
			}}
			err = db.C(VehicleCollection).Update(vehicleClampFilter, vehicleClampUpdate)
			if err != nil {
				util.Log("Error fetching payment:", err.Error())
				fmt.Printf("error...")
				c.JSON(200, gin.H{
					"message": "Error Updating Vehicle",
				})
				return
			}
			util.Log("isClamped == true")
			c.JSON(200, gin.H{
				"message": "isClamped updated Successfully --> true",
			})

		}
	}
	vFilter := bson.M{"_id": vID}
	vUpdate := bson.M{"$set": bson.M{
		"isClamped":      false,
		"isWaitingClamp": false,
	}}
	err = db.C(VehicleCollection).Update(vFilter, vUpdate)
	if err != nil {
		util.Log("Error updating payment:", err.Error())
		fmt.Printf("error...")
		return
	}
	util.Log("Vehicle Parking Fee Paid, don't proceed to clamp")
	c.JSON(200, gin.H{
		"message": "Paid, don't clamp",
	})
	util.SendNotifications("fi3ytpKGhRo:APA91bFqPPPFnpeQo2BRxB0NKTMfGxmaZNwX0XNu4NnJsz7inArbgrkDihHJF_om46NW2Bd-1pwHHZmOiV03s2hSZ_XLm2EkbxxOmwH9KukPaaZeq_0dSXe5giGCeD3s924XZDkMDfLv", "The vehicle has not yet been paid , Please clamp!")
	return
}

func ClearClampFee(c *gin.Context) {
	//db := *MongoConfig()
	db := c.MustGet("db").(*mgo.Database)

	fmt.Println("MONGO RUNNING: ", db)
	clampFee := model.ClampFee{}
	err := c.Bind(&clampFee)
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Getting Body",
		})
		c.Abort()
		return
	}
	userID := c.Param("id")
	clampFee.UserID = userID
	clampFee.IsSuccessful = false
	clampFee.ClampFeeID = primitive.NewObjectID()
	err = db.C(ClampFeeCollection).Insert(clampFee)
	if err != nil {
		c.JSON(403, gin.H{
			"message": "Error Inserting Clamp Fee Payment",
		})
		c.Abort()
		return
	}
	util.Log("Payment (amount), added successfully...")
	c.JSON(200, gin.H{
		"message":   "Success Insert Vehicle",
		"clamp fee": &clampFee,
	})
	cID := clampFee.ClampFeeID.Hex()
	log.Println(cID)

	//--------------------------------------------STK Push--------------------------------------------------------------
	util.Log("ClampPushHandler Initialized...")
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
		CallBackURL:       "http://34.121.65.106:3500/clamprcb?id=" + userID + "&paymentID=" + cID, //ClampCallBackHandler
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
