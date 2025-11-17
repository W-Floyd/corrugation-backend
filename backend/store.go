package backend

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/W-Floyd/corrugation-backend/frontend"
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
	ID          frontend.EntityID  `json:"id" required:"false"`
	Name        *string            `json:"name" required:"false"`
	Description *string            `json:"description" required:"false"`
	Artifacts   []*uint            `json:"artifacts" required:"false"`
	Location    *frontend.EntityID `json:"location" required:"false"`
	Metadata    *Metadata          `json:"metadata" required:"false"`
}

type EntityInput struct {
	ID          frontend.EntityID  `json:"id"`
	Name        *string            `json:"name" required:"false"`
	Description *string            `json:"description" required:"false"`
	Artifacts   []*uint            `json:"artifacts" required:"false"`
	Location    *frontend.EntityID `json:"location" required:"false"`
	Metadata    *Metadata          `json:"metadata" required:"false"`
}

type StoreOutput struct {
	Body struct {
		Entities       map[frontend.EntityID]*EntityInput `json:"entities"`
		Artifacts      map[uint]frontend.FrontendArtifact `json:"artifacts"`
		LastArtifactID uint                               `json:"lastartifactid"`
		StoreVersion   int                                `json:"storeversion"`
	}
}

func (record *Record) ToEntity() (output *EntityInput, err error) {
	output =
		&EntityInput{
			ID: frontend.EntityID(record.ID),
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
			Location: func() (output *frontend.EntityID) {
				var v frontend.EntityID
				if record.ParentID == nil {
					v = 0
				} else {
					v = frontend.EntityID(*record.ParentID)
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

	output.Body.Entities = make(map[frontend.EntityID]*EntityInput)

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
		output.Body.Entities[frontend.EntityID(record.ID)] = e
	}

	output.Body.StoreVersion = int(newest.Unix())

	tx := db.Model(&Record{}).Unscoped().Select("MAX(id)").Find(&output.Body.LastArtifactID)
	if tx.Error != nil {
		err = tx.Error
		return
	}

	as, err := gorm.G[Artifact](db).Select("id", "content_type", "original_filename").Find(dbCtx)
	if err != nil {
		return
	}

	output.Body.Artifacts = map[uint]frontend.FrontendArtifact{}

	for _, a := range as {
		output.Body.Artifacts[a.ID] = frontend.FrontendArtifact{
			ID:   frontend.ArtifactID(a.ID),
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

func GetFirstFreeID(ctx context.Context, input *struct{}) (output *UIntOutput, err error) {
	output = &UIntOutput{}
	tx := db.Model(&Record{}).Unscoped().Select("MAX(id)").Find(&output.Body)
	if tx.Error != nil {
		err = tx.Error
		return
	}
	output.Body += 1
	return
}

var CreateEntityOperation = huma.Operation{
	Method: http.MethodPost,
	Path:   "/api/entity",
}

func CreateEntity(ctx context.Context, input *struct {
	Body EntityInput
}) (output *EntityInput, err error) {
	if input.Body.ID == 0 {
		err = huma.Error500InternalServerError("id must not be 0")
	}

	record := &Record{}

	record.ID = uint(input.Body.ID)
	if input.Body.Metadata.IsLabeled != nil && *input.Body.Metadata.IsLabeled {
		record.Label = input.Body.Name
	} else {
		record.Title = input.Body.Name
	}

	record.Description = input.Body.Description

	location := input.Body.Location
	if location != nil && *location != 0 {
		v := uint(*location)
		record.ParentID = &v
	}

	for _, a := range input.Body.Artifacts {
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

	output, err = record.ToEntity()
	if err != nil {
		return
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
		v := frontend.EntityID(0)
		entity.Location = &v
	}

	output = &EntityOutput{
		Body: *entity,
	}

	return
}

var DeleteEntityOperation = huma.Operation{
	Method: http.MethodDelete,
	Path:   "/api/entity/{id}",
}

var PatchEntityOperation = huma.Operation{
	Method: http.MethodPatch,
	Path:   "/api/entity/{id}",
}

func PatchEntity(ctx context.Context, input *struct {
	ID   uint `path:"id" example:"1" doc:"ID to update"`
	Body EntityPatch
}) (output *EntityInput, err error) {

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

	_, err = gorm.G[Record](db).Where("id = ?", r.ID).Updates(dbCtx, r)
	if err != nil {
		return
	}

	output, err = r.ToEntity()

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
