package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"log"
	"os"
)

var db *gorm.DB

func connect() *gorm.DB {
	var err error

	db, err = gorm.Open("mysql", "db_user:db_password@tcp(db:3306)/db_name?charset=utf8&parseTime=True&charset=utf8&loc=Local")
	if err != nil {
		log.Fatalf("Got error when connect database, the error is '%v'", err)
	}
	return db
}

// Using this function to get a connection, you can create your connection pool here.
func GetDB() *gorm.DB {
	if db == nil {
		newDb := connect()
		newDb.DB().SetMaxIdleConns(4)
		newDb.DB().SetMaxOpenConns(20)
		newDb.LogMode(os.Getenv("GIN_MODE") != "release")
		db = newDb
	}
	return db
}
