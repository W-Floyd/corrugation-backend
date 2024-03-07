package entity

import (
	"github.com/W-Floyd/corrugation/backend/asset"
	"github.com/W-Floyd/corrugation/backend/tag"
	"gorm.io/gorm"
)

type Entity struct {
	gorm.Model
	Label       string
	Title       string
	Description string

	Location *Entity

	// Properties
	Quantity int
	Tags     []*tag.Tag
	Assets   []*asset.Asset
}
