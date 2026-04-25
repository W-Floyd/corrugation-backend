package backend

import (
	"context"
	"errors"

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
	Log.Infow("connecting to DB", "path", dbFilePath)
	if db != nil {
		return errors.New("db is already defined, will not override")
	}

	db, err = gorm.Open(sqlite.Open(dbFilePath), &gorm.Config{
		Logger: newGORMLogger(),
	})
	return
}

func InitAndMigrateDB() (err error) {
	Log.Info("running DB migrations")
	db.AutoMigrate(&Record{})
	db.AutoMigrate(&Tag{})
	db.AutoMigrate(&Artifact{})
	db.AutoMigrate(&Embedding{})
	db.AutoMigrate(&GlobalConfig{})
	db.AutoMigrate(&User{})
	return nil
}
