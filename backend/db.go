package backend

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             100 * time.Millisecond, // Slow SQL threshold
			LogLevel:                  logger.Warn,            // Log level (overridden by SetInitialLogLevel)
			IgnoreRecordNotFoundError: false,                  // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,                   // Enable color
		},
	)

	db, err = gorm.Open(sqlite.Open(dbFilePath), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		log.Println("Connected to DB at", dbFilePath)
	}
	return
}

func InitAndMigrateDB() (err error) {
	log.Println("InitAndMigrateDB")
	db.AutoMigrate(&Record{})
	db.AutoMigrate(&Tag{})
	db.AutoMigrate(&Artifact{})
	db.AutoMigrate(&Embedding{})
	db.AutoMigrate(&GlobalConfig{})
	db.AutoMigrate(&UserConfig{})
	return nil
}
