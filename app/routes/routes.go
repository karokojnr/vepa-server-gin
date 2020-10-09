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
	r.Use(middlewares.Connect)
	r.Use(middlewares.ErrorHandler)
	//Routes
	r.POST("/register", controllers.RegisterHandler)
	r.POST("/login", controllers.LoginHandler)
	r.GET("/user/:id", controllers.ProfileHandler)
	r.PUT("/updateProfile/:id", controllers.EditProfile)
	r.PUT("/token/:id", controllers.FCMTokenHandler)
	r.POST("/addVehicle/:id", controllers.AddVehicleHandler)
	r.GET("/vehicle/:vehicleReg", controllers.GetVehicleHandler)
	r.PUT("/editVehicle/:id", controllers.EditVehicleHandler)
	//r.GET("/userVehicles/:id", controllers.UserVehiclesHandler)
	//r.DELETE("/deleteVehicle/:id", controllers.DeleteVehicleHandler)
	//r.GET("/isWaitingClamp", controllers.VehiclesWaitingClamp)
	//r.GET("/isClamped", controllers.ClampedVehicle)
	//r.POST("/makePayment/:id", controllers.PaymentHandler)

	port := util.GetPort()
	util.Log("Starting app on port 👍 ✓ ⌛ :", port)
	log.Fatal(http.ListenAndServe(port, r))
}
