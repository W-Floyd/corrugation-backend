package backend

import (
	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

func GetTags(Title *string, withRecords bool) (tags []Tag, err error) {
	if Title == nil || *Title == "" {
		return gorm.G[Tag](db).Find(dbCtx)
	}

	var tagsSearched []Tag // This should come back with one value...
	if withRecords {
		tags, err = gorm.G[Tag](db).Where("title = ?", *Title).Preload("Records", nil).Find(dbCtx)
	} else {
		tags, err = gorm.G[Tag](db).Where("title = ?", *Title).Find(dbCtx)
	}

	if err != nil {
		return
	}
	if len(tagsSearched) == 0 {
		err = huma.Error404NotFound(errorTagNotFound)
		return
	} else if len(tagsSearched) > 1 {
		err = huma.Error500InternalServerError(errorMoreTagsThanExpected)
		return
	}

	return

}
