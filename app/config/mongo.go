package config

import (
	"github.com/karokojnr/vepa-server-gin/app/util"
	mgo "gopkg.in/mgo.v2"
)

func GetMongoDB() (*mgo.Database, error) {
	host := util.GoDotEnvVariable("MONGO_HOST")
	dbName := util.GoDotEnvVariable("MONGO_DB_NAME")

	session, err := mgo.Dial(host)
	if err != nil {
		return nil, err
	}

	db := session.DB(dbName)
	return db, nil
}
