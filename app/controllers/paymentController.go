package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/AndroidStudyOpenSource/mpesa-api-go"
	"github.com/gin-gonic/gin"
	model "github.com/karokojnr/vepa-server-gin/app/models"
	"github.com/karokojnr/vepa-server-gin/app/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func PaymentHandler(c *gin.Context) {
	ctx := context.TODO()
	paymentCollection, err := util.GetCollection("payments")
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Cannot get payment collection",
		})
		return
	}
	var payment model.Payment
	userID := c.Param("id")
	//id, _ := primitive.ObjectIDFromHex(userID)
	err = c.Bind(&payment)
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
	_, err = paymentCollection.InsertOne(ctx, payment)
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
	userCollection, err := util.GetCollection("users")
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Cannot get user collection",
		})
		return
	}
	var rUser model.User
	err = userCollection.FindOne(ctx, filter).Decode(&rUser)
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
		PartyA:            "254799338805",
		PartyB:            "174379",
		PhoneNumber:       "254799338805",
		CallBackURL:       "https://gin-vepa.herokuapp.com/rcb?id=" + userID + "&paymentid=" + pID, //CallBackHandler
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

func CallBackHandler(c *gin.Context, r *http.Request) {
	log.Println("Callback called by M-pesa...")
	util.Log("Callback called by M-pesa...")
	var bd interface{}
	rbody := r.Body
	body, err := ioutil.ReadAll(rbody)
	err = json.Unmarshal(body, &bd)
	resultCode := bd.(map[string]interface{})["Body"].(map[string]interface{})["stkCallback"].(map[string]interface{})["ResultCode"]
	log.Println("Result code")
	log.Println(resultCode)
	rBody := bd.(map[string]interface{})["Body"].(map[string]interface{})["stkCallback"].(map[string]interface{})["ResultDesc"]
	util.Log("Result code:", resultCode, " Result Body:", rBody)
	log.Println("Result code:", resultCode, " Result Body:", rBody)

	util.Log("Reading request body...")
	if err != nil {
		log.Println("Error")
		util.Log("Error parsing request:", err.Error())
		//res.Result = "Unable to read request"
		//json.NewEncoder(w).Encode(res)
		return
	}
	//var bd interface{}
	log.Println("--C--")
	log.Println(c)

}
func UserPaymentsHandler(c *gin.Context) {
	var results []*model.Payment
	ctx := context.TODO()
	paymentCollection, err := util.GetCollection("payments")
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Cannot get payment collection",
		})
		return
	}
	userID := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"userId": id, "isSuccessful": true}
	cur, err := paymentCollection.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	for cur.Next(context.TODO()) {
		var elem model.Payment
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, &elem)

	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	_ = cur.Close(context.TODO())
	c.JSON(200, gin.H{
		"payments": &results,
	})
	return
}
func GetPaidDays(c *gin.Context) {
	ctx := context.TODO()
	paymentCollection, err := util.GetCollection("payments")
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Cannot get payment collection",
		})
		return
	}
	vehicleReg := c.Param("vehicleReg")
	var results []*model.Payment
	filter := bson.M{"vehicleReg": vehicleReg, "isSuccessful": true}
	cur, err := paymentCollection.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	for cur.Next(context.TODO()) {
		var elem model.Payment
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, &elem)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	_ = cur.Close(context.TODO())
	c.JSON(200, gin.H{
		"payments": &results,
	})
	return
}
func VerificationHandler(c *gin.Context) {
	ctx := context.TODO()
	paymentCollection, err := util.GetCollection("payments")
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Cannot get payment collection",
		})
		return
	}
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Cannot get vehicle collection",
		})
		return
	}
	vehicleReg := c.Param("vehicleReg")
	var payment model.Payment
	var vehicle model.Vehicle
	err = vehicleCollection.FindOne(ctx, bson.M{"registrationNumber": vehicleReg}).Decode(&vehicle)
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
	err = paymentCollection.FindOne(context.TODO(), filter).Decode(&payment)
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
	ctx := context.TODO()
	paymentCollection, err := util.GetCollection("payments")
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Cannot get payment collection",
		})
		return
	}
	vehicleReg := c.Param("vehicleReg")

	var results []*model.Payment
	filter := bson.M{"vehicleReg": vehicleReg}
	cur, err := paymentCollection.Find(ctx, filter)
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Getting All Clamped Vehicles",
		})
		return

	}
	for cur.Next(context.TODO()) {
		var elem model.Payment
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, &elem)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	_ = cur.Close(context.TODO())
	c.JSON(200, gin.H{
		"payments": &results,
	})
	return
}
