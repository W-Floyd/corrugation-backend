package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// TODO: Complete and test

func init() {
	DbHandlers["mysql"] = DbHandler(func(dbName, dbUser, dbPassword, dbAddress, dbTimeZone string) (db *gorm.DB, err error) {
		dsn := dbUser + ":" + dbPassword + "@tcp(" + dbAddress + ")/" + dbName + "?charset=utf8mb4&parseTime=True&loc=Local"
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		return
	})
}
