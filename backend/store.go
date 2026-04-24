package backend

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/W-Floyd/corrugation/oldbackend"
	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

type Metadata struct {
	Quantity       *int      `json:"quantity" required:"false"`
	Owners         []*string `json:"owners" required:"false"`
	Tags           []*string `json:"tags" required:"false"`
	LastModified   *string   `json:"lastmodified" required:"false"`
	LastModifiedBy *string   `json:"lastmodifiedby" required:"false"`
	IsLabeled      *bool     `json:"islabeled" required:"false"`
}

type EntityOutput struct {
	Body EntityInput
}
type EntityPatch struct {
	ID          oldbackend.EntityID  `json:"id" required:"false"`
	Name        *string              `json:"name" required:"false"`
	Description *string              `json:"description" required:"false"`
	Artifacts   []*uint              `json:"artifacts" required:"false"`
	Location    *oldbackend.EntityID `json:"location" required:"false"`
	Metadata    *Metadata            `json:"metadata" required:"false"`
}

type EntityInput struct {
	ID          oldbackend.EntityID  `json:"id" required:"false"`
	Name        *string              `json:"name" required:"false"`
	Description *string              `json:"description" required:"false"`
	Artifacts   []*uint              `json:"artifacts" required:"false"`
	Location    *oldbackend.EntityID `json:"location" required:"false"`
	Metadata    *Metadata            `json:"metadata" required:"false"`
}

type StoreOutput struct {
	Body struct {
		Entities       map[oldbackend.EntityID]*EntityInput `json:"entities"`
		Artifacts      map[uint]oldbackend.FrontendArtifact `json:"artifacts"`
		LastArtifactID uint                                 `json:"lastartifactid"`
		StoreVersion   int                                  `json:"storeversion"`
	}
}

func (record *Record) ToEntity() (output *EntityInput, err error) {
	output =
		&EntityInput{
			ID: oldbackend.EntityID(record.ID),
			Name: func() *string {
				if record.Label != nil {
					return record.Label
				} else if record.Title != nil {
					return record.Title
				} else {
					v := strconv.Itoa(int(record.ID))
					return &v
				}
			}(),
			Description: func() *string {
				if record.Description == nil {
					v := ""
					return &v
				}
				return record.Description
			}(),
			Location: func() (output *oldbackend.EntityID) {
				var v oldbackend.EntityID
				if record.ParentID == nil {
					v = 0
				} else {
					v = oldbackend.EntityID(*record.ParentID)
				}
				return &v
			}(),
			Artifacts: func() (output []*uint) {
				for _, a := range record.Artifacts {
					output = append(output, &a.ID)
				}
				return
			}(),
			Metadata: &Metadata{
				Owners: []*string{},
				Quantity: func() *int {
					var v int
					if record.Quantity == nil {
						v = 0
					} else {
						v = int(*record.Quantity)
					}
					return &v
				}(),
				Tags: func() (out []*string) {
					for _, tag := range record.Tags {
						out = append(out, &tag.Title)
					}
					return
				}(),
				LastModified: func() *string {
					v := record.UpdatedAt.UTC().Format("2006-01-02 15:04:05.000000") + " UTC"
					return &v
				}(),
				IsLabeled: func() *bool {
					var v bool
					if record.Label != nil {
						v = true
						return &v
					} else {
						return nil
					}
				}(),
			},
		}
	return
}

var GetStoreOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/store",
}

func GetStore(ctx context.Context, input *struct{}) (output *StoreOutput, err error) {

	output = &StoreOutput{}

	output.Body.Entities = make(map[oldbackend.EntityID]*EntityInput)

	records, err := GetRecords(nil, nil, nil, nil, []struct {
		q string
		h func(db gorm.PreloadBuilder) error
	}{
		{
			q: "Artifacts",
			h: func(db gorm.PreloadBuilder) error {
				db.Select("id", "record_id")
				return nil
			},
		},
		{
			q: "Tags",
			h: func(db gorm.PreloadBuilder) error { return nil },
		},
	}, nil)
	if err != nil {
		return
	}

	var newest time.Time

	for _, record := range records {
		if record.UpdatedAt.After(newest) {
			newest = record.UpdatedAt
		}
		if record.CreatedAt.After(newest) {
			newest = record.CreatedAt
		}
		if record.DeletedAt.Time.After(newest) {
			newest = record.DeletedAt.Time
		}

		e, err := record.ToEntity()
		if err != nil {
			return output, err
		}
		output.Body.Entities[oldbackend.EntityID(record.ID)] = e
	}

	output.Body.StoreVersion = int(newest.Unix())

	var lastID *uint
	tx := db.Model(&Artifact{}).Unscoped().Select("MAX(id)").Scan(&lastID)
	if tx.Error != nil {
		err = tx.Error
		return
	}
	if lastID != nil {
		output.Body.LastArtifactID = *lastID
	}

	as, err := gorm.G[Artifact](db).Select("id", "content_type", "original_filename").Find(dbCtx)
	if err != nil {
		return
	}

	output.Body.Artifacts = map[uint]oldbackend.FrontendArtifact{}

	for _, a := range as {
		output.Body.Artifacts[a.ID] = oldbackend.FrontendArtifact{
			ID:   oldbackend.ArtifactID(a.ID),
			Path: "/api/artifact/" + strconv.FormatUint(uint64(a.ID), 10),
			Image: func() bool {
				i, _ := a.GetInterface()
				switch i.(type) {
				case *Image:
					return true
				default:
					return false
				}
			}(),
		}
	}

	return
}

type IntOutput struct {
	Body int
}

type UIntOutput struct {
	Body uint
}

type StringOutput struct {
	Body string
}

var GetStoreVersionOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/store/version",
}

func GetStoreVersion(ctx context.Context, input *struct{}) (output *UIntOutput, err error) {
	records, err := GetRecords(nil, nil, nil, nil, nil, []string{"updated_at", "created_at", "deleted_at"})
	if err != nil {
		return
	}

	var newest time.Time

	for _, record := range records {
		if record.UpdatedAt.After(newest) {
			newest = record.UpdatedAt
		}
		if record.CreatedAt.After(newest) {
			newest = record.CreatedAt
		}
		if record.DeletedAt.Time.After(newest) {
			newest = record.DeletedAt.Time
		}
	}
	output = &UIntOutput{
		Body: uint(newest.Unix()),
	}
	return
}

var GetFirstFreeIDOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/entity/find/firstfreeid",
}

func getFirstFreeID() (id *uint, err error) {
	tx := db.Model(&Record{}).Unscoped().Select("MAX(id)").Scan(&id)
	if tx.Error != nil {
		err = tx.Error
		return
	}
	if id == nil {
		var v uint = 1
		id = &v
	} else {
		*id++
	}
	return
}

func GetFirstFreeID(ctx context.Context, input *struct{}) (output *UIntOutput, err error) {
	output = &UIntOutput{}
	maxID, err := getFirstFreeID()
	output.Body = *maxID
	return
}

var CreateEntityOperation = huma.Operation{
	Method: http.MethodPost,
	Path:   "/api/entity",
}

func CreateEntity(ctx context.Context, input *struct {
	Body EntityInput
}) (output *StringOutput, err error) {

	record := &Record{}

	if input.Body.ID == 0 {
		var maxID *uint
		maxID, err = getFirstFreeID()
		if err != nil {
			return
		}
		record.ID = *maxID
	} else {
		record.ID = uint(input.Body.ID)
	}

	if input.Body.Metadata.IsLabeled != nil && *input.Body.Metadata.IsLabeled {
		record.Label = input.Body.Name
	} else {
		record.Title = input.Body.Name
	}

	record.Description = input.Body.Description

	if input.Body.Metadata != nil && input.Body.Metadata.Quantity != nil {
		v := uint(*input.Body.Metadata.Quantity)
		record.Quantity = &v
	}

	location := input.Body.Location
	if location != nil && *location != 0 {
		v := uint(*location)
		record.ParentID = &v
	}

	for _, a := range input.Body.Artifacts {
		if a == nil {
			log.Println("empty artifact!")
			continue
		}
		var artifact Artifact
		artifact, err = GetArtifactFromDB(*a)
		if err != nil {
			return
		}
		record.Artifacts = append(record.Artifacts, &artifact)
	}

	var foundTags []Tag
	var foundTag *Tag
	for _, tag := range input.Body.Metadata.Tags {
		foundTags, err = gorm.G[Tag](db).Where("title = ?", tag).Find(dbCtx)
		if err != nil {
			return
		} else if len(foundTags) > 1 {
			err = huma.Error500InternalServerError(errorMoreTagsThanExpected)
			return
		} else if len(foundTags) == 1 {
			foundTag = &foundTags[0]
		} else {
			newtag := Tag{
				Title: *tag,
			}
			err = gorm.G[Tag](db).Create(dbCtx, &newtag)
			if err != nil {
				return
			}
			foundTag = &newtag
		}
		record.Tags = append(record.Tags, foundTag)
	}

	err = gorm.G[Record](db).Create(dbCtx, record)
	if err != nil {
		return
	}

	Broadcast()

	output = &StringOutput{
		Body: strconv.FormatUint(uint64(record.ID), 10),
	}

	return
}

var GetAllEntitiesOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/entity",
}

type AllEntitiesOutput struct {
	Body map[oldbackend.EntityID]*EntityInput
}

func GetAllEntities(ctx context.Context, input *struct{}) (output *AllEntitiesOutput, err error) {
	records, err := GetRecords(nil, nil, nil, nil, []struct {
		q string
		h func(db gorm.PreloadBuilder) error
	}{
		{
			q: "Artifacts",
			h: func(db gorm.PreloadBuilder) error {
				db.Select("id", "record_id")
				return nil
			},
		},
		{
			q: "Tags",
			h: func(db gorm.PreloadBuilder) error { return nil },
		},
	}, nil)
	if err != nil {
		return
	}

	output = &AllEntitiesOutput{
		Body: make(map[oldbackend.EntityID]*EntityInput),
	}
	for _, record := range records {
		e, e2 := record.ToEntity()
		if e2 != nil {
			err = e2
			return
		}
		output.Body[oldbackend.EntityID(record.ID)] = e
	}
	return
}

var GetEntityOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/entity/{id}",
}

func GetEntity(ctx context.Context, input *struct {
	ID uint `path:"id" example:"1" doc:"ID to get"`
}) (output *EntityOutput, err error) {
	if input.ID == 0 {
		err = huma.Error500InternalServerError("id must not be 0")
	}

	records, err := GetRecords(&input.ID, nil, nil, nil, nil, nil)
	if err != nil {
		return
	}

	entity, err := records[0].ToEntity()
	if err != nil {
		return
	}

	if entity.Location == nil {
		v := oldbackend.EntityID(0)
		entity.Location = &v
	}

	output = &EntityOutput{
		Body: *entity,
	}

	return
}

var DeleteEntityOperation = huma.Operation{
	Method:       http.MethodDelete,
	Path:         "/api/entity/{id}",
	DefaultStatus: http.StatusOK,
}

type DeleteOutput struct {
	Status int
}

func DeleteEntity(ctx context.Context, input *struct {
	ID uint `path:"id" example:"1" doc:"ID to delete"`
}) (output *DeleteOutput, err error) {
	result := db.Where("id = ?", input.ID).Delete(&Record{})
	if result.Error != nil {
		err = result.Error
		return
	}
	output = &DeleteOutput{}
	if result.RowsAffected == 0 {
		output.Status = http.StatusNoContent
	} else {
		output.Status = http.StatusOK
		Broadcast()
	}
	return
}

var PatchEntityOperation = huma.Operation{
	Method: http.MethodPatch,
	Path:   "/api/entity/{id}",
}

func PatchEntity(ctx context.Context, input *struct {
	ID   uint `path:"id" example:"1" doc:"ID to update"`
	Body EntityPatch
}) (output *EntityOutput, err error) {

	records, err := GetRecords(&input.ID, nil, nil, nil, nil, nil)
	if err != nil {
		return
	}

	i := input.Body
	r := records[0]

	if i.Name != nil {
		if r.Label != nil {
			r.Label = i.Name
		} else {
			r.Title = i.Name
		}
	}

	if i.Description != nil {
		r.Description = i.Description
	}

	if i.Location != nil {
		v := uint(*i.Location)
		r.ParentID = &v
	}

	var artifact Artifact
	for _, a := range i.Artifacts {
		artifact, err = GetArtifactFromDB(*a)
		if err != nil {
			return
		}
		r.Artifacts = append(r.Artifacts, &artifact)
	}

	_, err = gorm.G[Record](db).Where("id = ?", r.ID).Updates(dbCtx, r)
	if err != nil {
		return
	}

	Broadcast()

	entity, err := r.ToEntity()
	if err != nil {
		return
	}

	output = &EntityOutput{
		Body: *entity,
	}

	return
}

var ResetStoreOperation = huma.Operation{
	Method: http.MethodPost,
	Path:   "/api/reset",
}

func ResetStore(ctx context.Context, input *struct{}) (output *EmptyOutput, err error) {
	if err = db.Unscoped().Where("1 = 1").Delete(&Record{}).Error; err != nil {
		return
	}
	if err = db.Unscoped().Where("1 = 1").Delete(&Artifact{}).Error; err != nil {
		return
	}
	if err = db.Unscoped().Where("1 = 1").Delete(&Tag{}).Error; err != nil {
		return
	}
	Broadcast()
	output = &EmptyOutput{}
	return
}

var CreateArtifactStoreOperation = huma.Operation{
	Method: http.MethodPost,
	Path:   "/api/artifact",
}

var GetArtifactStoreOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/artifact/{id}",
}
