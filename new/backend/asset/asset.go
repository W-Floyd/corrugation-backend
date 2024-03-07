package asset

import (
	"github.com/gabriel-vasile/mimetype"
	"gorm.io/gorm"
)

type Asset struct {
	gorm.Model
	Title    string
	Filepath string
	MIME     mimetype.MIME
}
