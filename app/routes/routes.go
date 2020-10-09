package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/karokojnr/vepa-server-gin/app/controllers"
	"github.com/karokojnr/vepa-server-gin/app/middlewares"
	"github.com/karokojnr/vepa-server-gin/app/util"
	"log"
	"net/http"
)

func Routes() {
	r := gin.Default()
	// Middlewares
	r.Use(middlewares.ErrorHandler)
	//Routes
	//Users
	r.POST("/register", controllers.RegisterHandler)
	r.POST("/login", controllers.LoginHandler)
	r.GET("/profile/:id", controllers.ProfileHandler)
	r.PUT("/editProfile", controllers.EditProfile)
	r.PUT("/token/:id", controllers.FCMTokenHandler)
	//Vehicles
	r.POST("/addVehicle", controllers.AddVehicleHandler)
	r.GET("/vehicle/:vehicleReg", controllers.GetVehicleHandler)
	r.PUT("/editVehicle/:id", controllers.EditVehicleHandler)
	r.GET("/userVehicles/:id", controllers.UserVehiclesHandler)
	r.DELETE("/deleteVehicle/:id", controllers.DeleteVehicleHandler)
	r.GET("/isWaitingClamp", controllers.VehiclesWaitingClamp)
	r.GET("/isClamped", controllers.ClampedVehicle)
	r.GET("/isVehicleClamped/{vehicleReg}", controllers.CheckIfVehicleIsClampedHandler)
	//Payment
	r.POST("/makePayment/:id", controllers.PaymentHandler)
	r.POST("/rcb", controllers.CallBackHandler)
	r.GET("/userPayments/:id", controllers.UserPaymentsHandler)
	r.GET("/fetchPaidDays/:vehicleReg", controllers.GetPaidDays)
	r.GET("/verifyPayment/:vehicleReg", controllers.VerificationHandler)
	r.GET("/unpaidVehicleHistory/vehicleReg", controllers.UnpaidVehicleHistoryHandler)
	r.GET("/clampVehicle/:vehicleReg", controllers.ClampVehicleHandler)
	r.POST("/clearclampfee/:id", controllers.ClearClampFeeHandler)
	r.POST("/clamprcb", controllers.ClampCallBackHandler)

	port := util.GetPort()
	util.Log("Starting app on port 👍 ✓ ⌛ :", port)
	log.Fatal(http.ListenAndServe(port, r))
}
