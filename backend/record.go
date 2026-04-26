package backend

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

const maxSearchDepth = 100

type RecordInput struct {
	Quantity        *uint       `required:"false"`
	ReferenceNumber *string     `required:"false"`
	Labeled         *bool       `required:"false"`
	Title           *string     `required:"false"`
	Description     *string     `required:"false"`
	Tags            []*TagInput `required:"false"`
	ParentID        *uint       `required:"false"`
	Artifacts       []*uint     `required:"false"`
}

type Record struct {
	gorm.Model

	Quantity        *uint   `json:",omitempty"`
	ReferenceNumber *string `json:",omitempty" gorm:"uniqueIndex"`
	Labeled         bool    `json:"Labeled"`
	Title           *string `json:",omitempty" gorm:"index"`
	Description     *string `json:",omitempty"`
	Tags            []*Tag  `json:",omitempty" gorm:"many2many:record_tags;"`

	Artifacts []*Artifact `json:",omitempty"`

	ParentID *uint   `json:",omitempty"`
	Parent   *Record `gorm:"foreignKey:ParentID" json:"-"`

	OwnerID *uint `json:",omitempty"`
	Owner   *User `gorm:"foreignKey:OwnerID" json:"-"`

	SearchConfidenceImage *float64 `gorm:"-" json:",omitempty"`
	SearchConfidenceText  *float64 `gorm:"-" json:",omitempty"`
}

func (i *RecordInput) Convert() (o Record, err error) {
	o.Quantity = i.Quantity
	o.ReferenceNumber = i.ReferenceNumber
	if i.Labeled != nil {
		o.Labeled = *i.Labeled
	}
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
	if record.ReferenceNumber != nil && *record.ReferenceNumber != "" {
		output += " (" + *record.ReferenceNumber + ")"
	}
	return
}

// GetRecordEmbeddings returns text embeddings for the given record IDs.
// Only jobs for those specific IDs are enqueued and waited on, so partial
// reflects completeness within the requested scope only.
func GetRecordEmbeddings(ctx context.Context, scopedIDs []uint) (e map[uint][]float64, partial bool, err error) {
	uc, _ := loadUser(UsernameFromContext(ctx))
	textModel, _, _, _ := effectiveInfinityConfig(uc)

	var embeddings []Embedding
	if err = db.Where("record_id IN ? AND embed_model = ?", scopedIDs, textModel).Find(&embeddings).Error; err != nil {
		return
	}

	e = map[uint][]float64{}
	embeddedIDs := map[uint]bool{}

	for _, emb := range embeddings {
		if emb.RecordID == nil {
			continue
		}
		var vec []float64
		if cached, ok := embeddingsCache.Load(emb.Hash); ok {
			vec = cached.(Embeddings)
		} else {
			if err = json.Unmarshal(emb.Data, &vec); err != nil {
				return
			}
			embeddingsCache.Store(emb.Hash, Embeddings(vec))
		}
		e[*emb.RecordID] = vec
		embeddedIDs[*emb.RecordID] = true
	}

	enqueuedIDs := generateMissingRecordEmbeddings(ctx, scopedIDs, embeddedIDs, "search")
	if len(enqueuedIDs) > 0 {
		if WaitForEmbeddingJobs(ctx, JobTypeRecord, enqueuedIDs, textModel) {
			for _, id := range enqueuedIDs {
				reloaded, reloadErr := gorm.G[Embedding](db).Where("record_id = ? AND embed_model = ?", id, textModel).Find(dbCtx)
				if reloadErr != nil || len(reloaded) == 0 {
					continue
				}
				var vec []float64
				if cached, ok := embeddingsCache.Load(reloaded[0].Hash); ok {
					vec = cached.(Embeddings)
				} else if jsonErr := json.Unmarshal(reloaded[0].Data, &vec); jsonErr == nil {
					embeddingsCache.Store(reloaded[0].Hash, Embeddings(vec))
				}
				if vec != nil {
					e[id] = vec
				}
			}
		} else {
			partial = true
		}
	}

	return
}

// generateMissingRecordEmbeddings enqueues embedding jobs for record IDs not in embeddedIDs.
// Returns the IDs that were enqueued.
func generateMissingRecordEmbeddings(ctx context.Context, recordIDs []uint, embeddedIDs map[uint]bool, source string) []uint {
	uc, _ := loadUser(UsernameFromContext(ctx))
	textModel, _, _, _ := effectiveInfinityConfig(uc)
	var ownerID *uint
	if uc.ID > 0 {
		ownerID = &uc.ID
	}
	username := UsernameFromContext(ctx)

	var enqueued []uint
	for _, id := range recordIDs {
		if !embeddedIDs[id] {
			EnqueueEmbeddingJob(JobTypeRecord, id, ownerID, username, textModel, source)
			enqueued = append(enqueued, id)
		}
	}
	Log.Infow("generateMissingRecordEmbeddings", "source", source, "username", username, "total", len(recordIDs), "enqueued", len(enqueued))
	return enqueued
}

type RecordResponse struct {
	ID                    uint        `json:"ID"`
	CreatedAt             *time.Time  `json:"CreatedAt,omitempty"`
	UpdatedAt             *time.Time  `json:"UpdatedAt,omitempty"`
	Quantity              *uint       `json:",omitempty"`
	ReferenceNumber       *string     `json:",omitempty"`
	Labeled               bool        `json:"Labeled"`
	Title                 *string     `json:",omitempty"`
	Description           *string     `json:",omitempty"`
	Tags                  []*Tag      `json:",omitempty"`
	Artifacts             []*Artifact `json:",omitempty"`
	ParentID              *uint       `json:",omitempty"`
	SearchConfidenceImage *float64    `json:",omitempty"`
	SearchConfidenceText  *float64    `json:",omitempty"`
}

func toRecordResponse(r Record, timestamps bool) RecordResponse {
	resp := RecordResponse{
		ID:                    r.ID,
		Quantity:              r.Quantity,
		ReferenceNumber:       r.ReferenceNumber,
		Labeled:               r.Labeled,
		Title:                 r.Title,
		Description:           r.Description,
		Tags:                  r.Tags,
		Artifacts:             r.Artifacts,
		ParentID:              r.ParentID,
		SearchConfidenceImage: r.SearchConfidenceImage,
		SearchConfidenceText:  r.SearchConfidenceText,
	}
	if timestamps {
		resp.CreatedAt = &r.Model.CreatedAt
		resp.UpdatedAt = &r.Model.UpdatedAt
	}
	return resp
}

func recordEmbeddingText(r Record) string {
	parts := []string{}
	if r.Title != nil && *r.Title != "" {
		parts = append(parts, *r.Title)
	}
	if r.ReferenceNumber != nil && *r.ReferenceNumber != "" {
		parts = append(parts, *r.ReferenceNumber)
	}
	if r.Description != nil && *r.Description != "" {
		parts = append(parts, *r.Description)
	}
	return strings.Join(parts, " - ")
}

func (r *Record) GenerateEmbeddings(ctx context.Context) (vec Embeddings, err error) {
	text := recordEmbeddingText(*r)
	if text == "" {
		return
	}

	uc, _ := loadUser(UsernameFromContext(ctx))
	textModel, _, _, _ := effectiveInfinityConfig(uc)

	var fullInput string
	vec, fullInput, err = GenerateTextDocumentEmbeddingsCtx(ctx, text)
	if err != nil {
		return
	}

	id := r.ID
	err = saveEmbedding(&id, nil, vec, textModel, fullInput)
	if err == nil {
		Log.Infof("embedding: record %d indexed with model %s", id, textModel)
	}
	return
}
