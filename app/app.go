package app

import (
	"github.com/joho/godotenv"
	"github.com/karokojnr/vepa-server-gin/app/routes"
	"github.com/karokojnr/vepa-server-gin/app/util"
)

func init() {
}

func Run() {

	godotenv.Load()
	util.InitLogger()
	routes.Routes()

}
