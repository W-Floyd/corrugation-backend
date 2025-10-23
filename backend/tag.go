package backend

import (
	"gorm.io/gorm"
)

type TagInput struct {
	Title string `required:"true" example:"Electrical" doc:"Title of tag being created"`
	Color string `required:"false" example:"#ff0000" doc:"Hex color of tag being created"`
}

type Tag struct {
	gorm.Model
	Title string `gorm:"uniqueIndex"`
	Color string
}

func (i *TagInput) Convert() (o Tag, err error) {
	o.Title = i.Title
	o.Color = i.Color
	return
}

func (record *Tag) PrettyString() (output string) {
	output = record.Title
	return
}
