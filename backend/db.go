package backend

import (
	"context"
	"errors"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	db    *gorm.DB
	dbCtx context.Context
)

func ConnectDB(dbFilePath string) (err error) {
	if dbCtx == nil {
		dbCtx = context.Background()
	}
	log.Println("ConnectDB")
	if db != nil {
		return errors.New("db is already defined, will not override")
	}

	db, err = gorm.Open(sqlite.Open(dbFilePath), &gorm.Config{})

	if err != nil {
		log.Println("Connected to DB at", dbFilePath)
	}
	return
}

func InitAndMigrateDB() (err error) {
	log.Println("InitAndMigrateDB")
	db.AutoMigrate(&Record{})
	db.AutoMigrate(&Tag{})
	return nil
}
