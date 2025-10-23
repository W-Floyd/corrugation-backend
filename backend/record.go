package backend

import (
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

const maxSearchDepth = 100

type RecordInput struct {
	Quantity    *uint     `required:"false"`
	Label       *string   `required:"false"`
	Title       *string   `required:"false"`
	Description *string   `required:"false"`
	Tags        []*string `required:"false"`
	ParentID    *uint     `required:"false"`
}

type Record struct {
	gorm.Model

	Quantity    *uint
	Label       *string
	Title       *string
	Description *string
	Tags        []*string `gorm:"type:text"`

	ParentID *uint
	Parent   *Record `gorm:"foreignKey:ParentID" json:"-"`
}

func (i *RecordInput) Convert() (o Record, err error) {
	o.Quantity = i.Quantity
	o.Label = i.Label
	o.Title = i.Title
	o.Description = i.Description

	if i.ParentID != nil {
		var found []Record
		found, err = gorm.G[Record](db).Where("id = ?", *i.ParentID).Find(dbCtx)
		if err != nil {
			return
		} else if len(found) > 1 {
			err = huma.Error500InternalServerError(errorMoreRecordsThanExpected)
			return
		} else if len(found) == 0 {
			err = huma.Error404NotFound(errorRecordNotFound)
			return
		}
		o.ParentID = i.ParentID
	}

	o.Tags = i.Tags

	return

}

func (record *Record) PrettyString() (output string) {
	output = strconv.FormatUint(uint64(record.ID), 10)
	if record.Label != nil && *record.Label != "" {
		output += " (" + *record.Label + ")"
	}
	return
}
