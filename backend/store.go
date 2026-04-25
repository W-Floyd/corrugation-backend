package backend

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/W-Floyd/corrugation/oldbackend"
	"github.com/danielgtaylor/huma/v2"
	qrcode "github.com/skip2/go-qrcode"
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
					out = []*string{}
					for _, tag := range record.Tags {
						out = append(out, &tag.Title)
					}
					return
				}(),
				LastModified: func() *string {
					v := record.UpdatedAt.UTC().Format("2006-01-02 15:04:05.000000") + " UTC"
					return &v
				}(),
				LastModifiedBy: record.LastModifiedBy,
				IsLabeled: func() *bool {
					v := record.Label != nil
					return &v
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

	if u := UsernameFromContext(ctx); u != "" {
		record.LastModifiedBy = &u
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

var ListEntityIDsOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/entity/list",
}

type EntityIDListOutput struct {
	Body []string
}

func ListEntityIDs(ctx context.Context, input *struct{}) (output *EntityIDListOutput, err error) {
	records, err := GetRecords(nil, nil, nil, nil, nil, []string{"id"})
	if err != nil {
		return
	}
	output = &EntityIDListOutput{Body: []string{}}
	for _, r := range records {
		output.Body = append(output.Body, strconv.FormatUint(uint64(r.ID), 10))
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

	records, err := GetRecords(&input.ID, nil, nil, nil, []struct {
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
	var record Record
	if err = db.Where("id = ?", input.ID).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
			output = &DeleteOutput{Status: http.StatusNoContent}
		}
		return
	}

	if record.ParentID != nil {
		db.Model(&Record{}).Where("parent_id = ?", input.ID).Update("parent_id", *record.ParentID)
	} else {
		db.Model(&Record{}).Where("parent_id = ?", input.ID).Update("parent_id", nil)
	}

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

	records, err := GetRecords(&input.ID, nil, nil, nil, []struct {
		q string
		h func(db gorm.PreloadBuilder) error
	}{
		{q: "Artifacts", h: func(db gorm.PreloadBuilder) error { return nil }},
		{q: "Tags", h: func(db gorm.PreloadBuilder) error { return nil }},
	}, nil)
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

	if i.Artifacts != nil {
		r.Artifacts = nil
		var artifact Artifact
		for _, a := range i.Artifacts {
			artifact, err = GetArtifactFromDB(*a)
			if err != nil {
				return
			}
			r.Artifacts = append(r.Artifacts, &artifact)
		}
	}

	if i.Metadata != nil {
		if i.Metadata.Quantity != nil {
			v := uint(*i.Metadata.Quantity)
			r.Quantity = &v
		}
		if i.Metadata.Tags != nil {
			var newTags []*Tag
			var foundTags []Tag
			var foundTag *Tag
			for _, tag := range i.Metadata.Tags {
				foundTags, err = gorm.G[Tag](db).Where("title = ?", tag).Find(dbCtx)
				if err != nil {
					return
				} else if len(foundTags) > 1 {
					err = huma.Error500InternalServerError(errorMoreTagsThanExpected)
					return
				} else if len(foundTags) == 1 {
					foundTag = &foundTags[0]
				} else {
					newtag := Tag{Title: *tag}
					err = gorm.G[Tag](db).Create(dbCtx, &newtag)
					if err != nil {
						return
					}
					foundTag = &newtag
				}
				newTags = append(newTags, foundTag)
			}
			if err = db.Model(&r).Association("Tags").Replace(newTags); err != nil {
				return
			}
			r.Tags = newTags
		}
	}

	if u := UsernameFromContext(ctx); u != "" {
		r.LastModifiedBy = &u
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

var GetQRCodeOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/qrcode/{id}",
}

func GetQRCode(_ context.Context, input *struct {
	ID uint `path:"id"`
}) (output *BytesOutput, err error) {
	return GetEntityQRCode(nil, input)
}

var GetEntityQRCodeOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/entity/{id}/qrcode",
}

func GetEntityQRCode(_ context.Context, input *struct {
	ID uint `path:"id"`
}) (output *BytesOutput, err error) {
	png, err := qrcode.Encode(strconv.FormatUint(uint64(input.ID), 10), qrcode.Medium, 1024)
	if err != nil {
		return
	}
	output = &BytesOutput{
		ContentType:  "image/png",
		CacheControl: "public, max-age=31536000",
		Body:         png,
	}
	return
}

var GetNextIDOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/entity/find/nextid",
}

func GetNextID(ctx context.Context, _ *struct{}) (output *UIntOutput, err error) {
	return GetFirstFreeID(ctx, &struct{}{})
}

var GetFirstLabeledIDOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/entity/find/firstid",
}

func GetFirstLabeledID(ctx context.Context, _ *struct{}) (output *UIntOutput, err error) {
	var labeledIDs []uint
	if tx := db.Model(&Record{}).Where("label IS NOT NULL").Order("id ASC").Pluck("id", &labeledIDs); tx.Error != nil {
		err = tx.Error
		return
	}
	labeled := make(map[uint]struct{}, len(labeledIDs))
	for _, id := range labeledIDs {
		labeled[id] = struct{}{}
	}
	var next uint = 1
	for {
		if _, used := labeled[next]; !used {
			break
		}
		next++
	}
	output = &UIntOutput{Body: next}
	return
}

var FindLocationsOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/entity/find/locations",
}

func FindLocations(ctx context.Context, _ *struct{}) (output *struct{ Body []uint }, err error) {
	type row struct {
		ParentID *uint
	}
	var rows []row
	if tx := db.Model(&Record{}).Select("DISTINCT parent_id").Scan(&rows); tx.Error != nil {
		err = tx.Error
		return
	}
	ids := []uint{}
	hasNull := false
	for _, r := range rows {
		if r.ParentID == nil {
			hasNull = true
		} else {
			ids = append(ids, *r.ParentID)
		}
	}
	if hasNull {
		ids = append([]uint{0}, ids...)
	}
	output = &struct{ Body []uint }{Body: ids}
	return
}

var FindLocationsFullOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/entity/find/locations/full",
}

func FindLocationsFull(ctx context.Context, _ *struct{}) (output *struct{ Body []*EntityInput }, err error) {
	var parentIDs []uint
	if tx := db.Model(&Record{}).Where("parent_id IS NOT NULL").Distinct("parent_id").Pluck("parent_id", &parentIDs); tx.Error != nil {
		err = tx.Error
		return
	}
	entities := []*EntityInput{}
	for _, id := range parentIDs {
		var records []Record
		if tx := db.Where("id = ?", id).Preload("Artifacts").Preload("Tags").Find(&records); tx.Error != nil {
			err = tx.Error
			return
		}
		if len(records) == 1 {
			e, e2 := records[0].ToEntity()
			if e2 != nil {
				err = e2
				return
			}
			entities = append(entities, e)
		}
	}
	output = &struct{ Body []*EntityInput }{Body: entities}
	return
}

var GetChildrenFullOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/entity/find/children/{id}/full",
}

func childrenQuery(id uint) *gorm.DB {
	if id == 0 {
		return db.Where("parent_id IS NULL")
	}
	return db.Where("parent_id = ?", id)
}

func GetChildrenFull(ctx context.Context, input *struct {
	ID uint `path:"id"`
}) (output *struct{ Body []*EntityInput }, err error) {
	var children []Record
	if tx := childrenQuery(input.ID).Preload("Artifacts").Preload("Tags").Find(&children); tx.Error != nil {
		err = tx.Error
		return
	}
	output = &struct{ Body []*EntityInput }{Body: []*EntityInput{}}
	for i := range children {
		e, e2 := children[i].ToEntity()
		if e2 != nil {
			err = e2
			return
		}
		output.Body = append(output.Body, e)
	}
	return
}

var GetChildrenFullRecursiveOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/entity/find/children/{id}/full/recursive",
}

func getDescendants(parentID uint, out *[]*EntityInput) error {
	var children []Record
	if tx := childrenQuery(parentID).Preload("Artifacts").Preload("Tags").Find(&children); tx.Error != nil {
		return tx.Error
	}
	for i := range children {
		e, err := children[i].ToEntity()
		if err != nil {
			return err
		}
		*out = append(*out, e)
		if err := getDescendants(children[i].ID, out); err != nil {
			return err
		}
	}
	return nil
}

func GetChildrenFullRecursive(ctx context.Context, input *struct {
	ID uint `path:"id"`
}) (output *struct{ Body []*EntityInput }, err error) {
	entities := []*EntityInput{}
	if err = getDescendants(input.ID, &entities); err != nil {
		return
	}
	output = &struct{ Body []*EntityInput }{Body: entities}
	return
}

var GetEntityContainsOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/entity/{id}/contains",
}

func GetEntityContains(ctx context.Context, input *struct {
	ID uint `path:"id" example:"1" doc:"Parent entity ID"`
}) (output *struct{ Body []uint }, err error) {
	var children []Record
	tx := childrenQuery(input.ID).Select("id").Find(&children)
	if tx.Error != nil {
		err = tx.Error
		return
	}
	if len(children) == 0 {
		output = &struct{ Body []uint }{Body: nil}
		return
	}
	ids := make([]uint, len(children))
	for i, c := range children {
		ids[i] = c.ID
	}
	output = &struct{ Body []uint }{Body: ids}
	return
}

var ReplaceEntityOperation = huma.Operation{
	Method: http.MethodPut,
	Path:   "/api/entity/{id}",
}

func ReplaceEntity(ctx context.Context, input *struct {
	ID   uint `path:"id" example:"1" doc:"ID to replace"`
	Body EntityInput
}) (output *EntityOutput, err error) {

	records, err := GetRecords(&input.ID, nil, nil, nil, []struct {
		q string
		h func(db gorm.PreloadBuilder) error
	}{
		{q: "Artifacts", h: func(db gorm.PreloadBuilder) error { return nil }},
		{q: "Tags", h: func(db gorm.PreloadBuilder) error { return nil }},
	}, nil)
	if err != nil {
		return
	}

	r := records[0]
	i := input.Body

	if i.Metadata != nil && i.Metadata.IsLabeled != nil && *i.Metadata.IsLabeled {
		r.Label = i.Name
		r.Title = nil
	} else {
		r.Title = i.Name
		r.Label = nil
	}

	r.Description = i.Description

	if i.Location != nil && *i.Location != 0 {
		v := uint(*i.Location)
		r.ParentID = &v
	} else {
		r.ParentID = nil
	}

	if i.Metadata != nil && i.Metadata.Quantity != nil {
		v := uint(*i.Metadata.Quantity)
		r.Quantity = &v
	} else {
		r.Quantity = nil
	}

	r.Artifacts = nil
	for _, a := range i.Artifacts {
		if a == nil {
			continue
		}
		var artifact Artifact
		artifact, err = GetArtifactFromDB(*a)
		if err != nil {
			return
		}
		r.Artifacts = append(r.Artifacts, &artifact)
	}

	var newTags []*Tag
	if i.Metadata != nil {
		var foundTags []Tag
		var foundTag *Tag
		for _, tag := range i.Metadata.Tags {
			foundTags, err = gorm.G[Tag](db).Where("title = ?", tag).Find(dbCtx)
			if err != nil {
				return
			} else if len(foundTags) > 1 {
				err = huma.Error500InternalServerError(errorMoreTagsThanExpected)
				return
			} else if len(foundTags) == 1 {
				foundTag = &foundTags[0]
			} else {
				newtag := Tag{Title: *tag}
				err = gorm.G[Tag](db).Create(dbCtx, &newtag)
				if err != nil {
					return
				}
				foundTag = &newtag
			}
			newTags = append(newTags, foundTag)
		}
	}
	if err = db.Model(&r).Association("Tags").Replace(newTags); err != nil {
		return
	}
	r.Tags = newTags

	if u := UsernameFromContext(ctx); u != "" {
		r.LastModifiedBy = &u
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

	output = &EntityOutput{Body: *entity}
	return
}

var ResetStoreOperation = huma.Operation{
	Method: http.MethodPost,
	Path:   "/api/reset",
}

func ResetStore(ctx context.Context, input *struct{}) (output *EmptyOutput, err error) {
	tables := []string{"record_tags", "artifacts", "records", "tags"}
	for _, table := range tables {
		if err = db.Exec("DELETE FROM " + table).Error; err != nil {
			return
		}
		db.Exec("DELETE FROM sqlite_sequence WHERE name = ?", table)
	}
	Broadcast()
	output = &EmptyOutput{}
	return
}

var GetArtifactQRCodeOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/artifact/{id}/qrcode",
}

func GetArtifactQRCode(_ context.Context, input *struct {
	ID uint `path:"id"`
}) (output *BytesOutput, err error) {
	png, err := qrcode.Encode(strconv.FormatUint(uint64(input.ID), 10), qrcode.Medium, 1024)
	if err != nil {
		return
	}
	output = &BytesOutput{
		ContentType:  "image/png",
		CacheControl: "public, max-age=31536000",
		Body:         png,
	}
	return
}

var DeleteArtifactOperation = huma.Operation{
	Method:        http.MethodDelete,
	Path:          "/api/artifact/{id}",
	DefaultStatus: http.StatusOK,
}

func DeleteArtifact(ctx context.Context, input *struct {
	ID uint `path:"id"`
}) (output *DeleteOutput, err error) {
	var a Artifact
	if tx := db.Select("id", "small_preview_id", "large_preview_id").First(&a, input.ID); tx.Error != nil {
		output = &DeleteOutput{Status: http.StatusNoContent}
		return
	}
	previewIDs := []uint{}
	if a.SmallPreviewID != nil {
		previewIDs = append(previewIDs, *a.SmallPreviewID)
	}
	if a.LargePreviewID != nil {
		previewIDs = append(previewIDs, *a.LargePreviewID)
	}
	result := db.Unscoped().Where("id = ?", input.ID).Delete(&Artifact{})
	if result.Error != nil {
		err = result.Error
		return
	}
	if len(previewIDs) > 0 {
		db.Unscoped().Where("id IN ?", previewIDs).Delete(&Artifact{})
	}
	output = &DeleteOutput{Status: http.StatusOK}
	Broadcast()
	return
}

var ListArtifactsOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/artifact/list",
}

func ListArtifacts(ctx context.Context, _ *struct{}) (output *struct{ Body []string }, err error) {
	artifacts, err := gorm.G[Artifact](db).Select("id", "content_type").Find(dbCtx)
	if err != nil {
		return
	}
	names := []string{}
	for _, a := range artifacts {
		ext := ".bin"
		if a.ContentType != nil {
			switch *a.ContentType {
			case "image/webp":
				ext = ".webp"
			case "image/png":
				ext = ".png"
			case "image/jpeg":
				ext = ".jpg"
			}
		}
		names = append(names, strconv.FormatUint(uint64(a.ID), 10)+ext)
	}
	output = &struct{ Body []string }{Body: names}
	return
}

var CreateArtifactStoreOperation = huma.Operation{
	Method: http.MethodPost,
	Path:   "/api/artifact",
}

func CreateArtifactStore(ctx context.Context, input *struct {
	RawBody huma.MultipartFormFiles[struct {
		File huma.FormFile `form:"file" required:"true"`
	}]
}) (output *StringOutput, err error) {
	result, err := CreateArtifact(ctx, input)
	if err != nil {
		return
	}
	output = &StringOutput{Body: strconv.FormatUint(uint64(result.Body), 10)}
	return
}

var GetArtifactStoreOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/artifact/{id}",
}
