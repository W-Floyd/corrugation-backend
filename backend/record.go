package backend

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

const maxSearchDepth = 100

type RecordInput struct {
	Quantity    *uint       `required:"false"`
	Label       *string     `required:"false"`
	Title       *string     `required:"false"`
	Description *string     `required:"false"`
	Tags        []*TagInput `required:"false"`
	ParentID    *uint       `required:"false"`
	Artifacts   []*uint     `required:"false"`
}

type Record struct {
	gorm.Model

	Quantity    *uint   `json:",omitempty"`
	Label       *string `json:",omitempty" gorm:"uniqueIndex"`
	Title       *string `json:",omitempty" gorm:"index"`
	Description *string `json:",omitempty"`
	Tags        []*Tag  `json:",omitempty" gorm:"many2many:record_tags;"`

	Artifacts []*Artifact `json:",omitempty"`

	ParentID *uint   `json:",omitempty"`
	Parent   *Record `gorm:"foreignKey:ParentID" json:"-"`

	LastModifiedBy *string `json:",omitempty"`

	SearchConfidenceImage *float64 `gorm:"-" json:",omitempty"`
	SearchConfidenceText  *float64 `gorm:"-" json:",omitempty"`
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
			err = huma.Error404NotFound(errorRecordNotFound + " " + strconv.Itoa(int(*i.ParentID)))
			return
		}
		o.ParentID = i.ParentID
	}

	var foundTags []Tag
	var foundTag *Tag

	for _, tag := range i.Tags {
		foundTags, err = gorm.G[Tag](db).Where("title = ?", tag.Title).Find(dbCtx)
		if err != nil {
			return
		} else if len(foundTags) > 1 {
			err = huma.Error500InternalServerError(errorMoreTagsThanExpected)
			return
		} else if len(foundTags) == 1 {
			foundTag = &foundTags[0]
		} else {
			var newtag Tag
			newtag, err = tag.Convert()
			if err != nil {
				return
			}
			err = gorm.G[Tag](db).Create(dbCtx, &newtag)
			if err != nil {
				return
			}
			foundTag = &newtag
		}
		o.Tags = append(o.Tags, foundTag)
	}

	for _, artifact := range i.Artifacts {
		var foundArtifact Artifact
		foundArtifact, err = GetArtifactFromDB(*artifact)
		if err != nil {
			return
		}
		o.Artifacts = append(o.Artifacts, &foundArtifact)
	}

	return

}

func (record *Record) PrettyString() (output string) {
	output = strconv.FormatUint(uint64(record.ID), 10)
	if record.Label != nil && *record.Label != "" {
		output += " (" + *record.Label + ")"
	}
	return
}

func GetRecordEmbeddings() (e map[uint][]float64, err error) {
	embeddings, err := gorm.G[Embedding](db).Where("record_id IS NOT NULL AND embed_model = ?", infinityTextModel).Find(dbCtx)
	if err != nil {
		return
	}

	e = map[uint][]float64{}
	embeddedIDs := map[uint]bool{}

	for _, emb := range embeddings {
		if emb.RecordID == nil {
			continue
		}
		var vec []float64
		if cached, ok := embeddingsCache[emb.Hash]; ok {
			vec = cached
		} else {
			if err = json.Unmarshal(emb.Data, &vec); err != nil {
				return
			}
			embeddingsCache[emb.Hash] = vec
		}
		e[*emb.RecordID] = vec
		embeddedIDs[*emb.RecordID] = true
	}

	records, err := GetRecords(nil, nil, nil, nil, nil, []string{"id", "title", "label", "description"})
	if err != nil {
		return
	}

	recordIDs := make([]uint, len(records))
	for i, r := range records {
		recordIDs[i] = r.ID
	}
	generateMissingRecordEmbeddings(recordIDs, embeddedIDs)

	for _, record := range records {
		if e[record.ID] != nil {
			continue
		}
		reloaded, reloadErr := gorm.G[Embedding](db).Where("record_id = ? AND embed_model = ?", record.ID, infinityTextModel).Find(dbCtx)
		if reloadErr != nil || len(reloaded) == 0 {
			continue
		}
		var vec []float64
		if cached, ok := embeddingsCache[reloaded[0].Hash]; ok {
			vec = cached
		} else {
			if jsonErr := json.Unmarshal(reloaded[0].Data, &vec); jsonErr != nil {
				continue
			}
			embeddingsCache[reloaded[0].Hash] = vec
		}
		e[record.ID] = vec
	}

	return
}

// generateMissingRecordEmbeddings generates text embeddings for record IDs not in embeddedIDs.
// Pass nil recordIDs to process all records.
func generateMissingRecordEmbeddings(recordIDs []uint, embeddedIDs map[uint]bool) {
	var candidates []uint
	for _, id := range recordIDs {
		if !embeddedIDs[id] {
			candidates = append(candidates, id)
		}
	}

	if len(candidates) == 0 {
		return
	}

	records, err := GetRecords(nil, nil, nil, nil, nil, []string{"id", "title", "label", "description"})
	if err != nil {
		log.Printf("embedding generation: failed to fetch records: %v", err)
		return
	}
	recordMap := map[uint]Record{}
	for _, r := range records {
		recordMap[r.ID] = r
	}

	for _, id := range candidates {
		r, ok := recordMap[id]
		if !ok {
			continue
		}
		if _, genErr := r.GenerateEmbeddings(); genErr != nil {
			log.Printf("embedding generation failed for record %d: %v", id, genErr)
		}
	}
}

type RecordResponse struct {
	ID                    uint        `json:"ID"`
	CreatedAt             *time.Time  `json:"CreatedAt,omitempty"`
	UpdatedAt             *time.Time  `json:"UpdatedAt,omitempty"`
	Quantity              *uint       `json:",omitempty"`
	Label                 *string     `json:",omitempty"`
	Title                 *string     `json:",omitempty"`
	Description           *string     `json:",omitempty"`
	Tags                  []*Tag      `json:",omitempty"`
	Artifacts             []*Artifact `json:",omitempty"`
	ParentID              *uint       `json:",omitempty"`
	LastModifiedBy        *string     `json:",omitempty"`
	SearchConfidenceImage *float64    `json:",omitempty"`
	SearchConfidenceText  *float64    `json:",omitempty"`
}

func toRecordResponse(r Record, timestamps bool) RecordResponse {
	resp := RecordResponse{
		ID:                    r.ID,
		Quantity:              r.Quantity,
		Label:                 r.Label,
		Title:                 r.Title,
		Description:           r.Description,
		Tags:                  r.Tags,
		Artifacts:             r.Artifacts,
		ParentID:              r.ParentID,
		LastModifiedBy:        r.LastModifiedBy,
		SearchConfidenceImage: r.SearchConfidenceImage,
		SearchConfidenceText:  r.SearchConfidenceText,
	}
	if timestamps {
		resp.CreatedAt = &r.Model.CreatedAt
		resp.UpdatedAt = &r.Model.UpdatedAt
	}
	return resp
}

func (r *Record) GenerateEmbeddings() (vec Embeddings, err error) {
	parts := []string{}
	if r.Title != nil && *r.Title != "" {
		parts = append(parts, *r.Title)
	}
	if r.Label != nil && *r.Label != "" {
		parts = append(parts, *r.Label)
	}
	if r.Description != nil && *r.Description != "" {
		parts = append(parts, *r.Description)
	}

	if len(parts) == 0 {
		return
	}

	vec, err = GenerateTextDocumentEmbeddings(strings.Join(parts, " - "))
	if err != nil {
		return
	}

	id := r.ID
	err = saveEmbedding(&id, nil, vec, infinityTextModel)
	return
}
