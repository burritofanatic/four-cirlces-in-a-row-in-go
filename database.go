package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"os"
	"sync"
	"time"
)

type DB struct {
	Database *mgo.Database
}

const (
	MongoDBHosts = "xxxxxxxxxx"
	AuthDatabase = "xxxxxxxxxx"
	AuthUserName = "xxxxxxxxxx"
	AuthPassword = "xxxxxxxxxx"
)

var _init_ctx sync.Once
var _instance *DB

func NewDB() *mgo.Database {
	_init_ctx.Do(func() {
		_instance = new(DB)

		mongoDBDialDict := &mgo.DialInfo{
			Addrs: []string{MongoDBHosts},
			Timeout: 400 * time.Second,
			Database: AuthDatabase,
			Username: AuthUserName,
			Password: AuthPassword,
		}

		session, err := mgo.DialWithInfo(mongoDBDialDict)

		if err != nil {
			fmt.Printf("Mongo connection error: %+v\n", err)
			os.Exit(1)
		}

		_instance.Database = session.DB(AuthDatabase)
	})


	return _instance.Database
}