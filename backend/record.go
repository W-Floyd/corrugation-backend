package backend

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

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

	Embedding     *[]byte `json:"-"` // JSON of embedding data
	EmbeddingHash *string `json:"-"` // Hash of JSON of embedding data (to allow caching)

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
	records, err := GetRecords(nil, nil, nil, nil, nil, []string{"id", "embedding", "embedding_hash", "title", "label", "description"})
	if err != nil {
		return
	}

	e = map[uint][]float64{}

	for _, record := range records {
		r := &record
		if r.EmbeddingHash == nil || r.Embedding == nil {
			err = r.GenerateEmbeddings()
			if err != nil {
				return
			}

			_, err = gorm.G[Record](db).Where("id = ?", r.ID).Update(dbCtx, "embedding", *r.Embedding)
			if err != nil {
				return
			}
			_, err = gorm.G[Record](db).Where("id = ?", r.ID).Update(dbCtx, "embedding_hash", r.EmbeddingHash)
			if err != nil {
				return
			}
		}

		var singleE []float64
		singleE, ok := embeddingsCache[*r.EmbeddingHash]
		if !ok {
			singleE = []float64{}
			subErr := json.Unmarshal(*r.Embedding, &singleE)
			if subErr != nil {
				err = subErr
				return
			}
			embeddingsCache[*r.EmbeddingHash] = singleE
		}
		e[r.ID] = singleE
	}
	return
}

func (r *Record) GenerateEmbeddings() (err error) {
	search := []string{}
	if r.Title != nil && *r.Title != "" {
		search = append(search, *r.Title)
	}
	if r.Label != nil && *r.Label != "" {
		search = append(search, *r.Label)
	}
	if r.Description != nil && *r.Description != "" {
		search = append(search, *r.Description)
	}

	log.Println("hgmmm")

	e, err := GenerateEmbeddings(strings.Join(search, " - "))
	if err != nil {
		return
	}

	hash, jsonData, err := e.MarshalEmbeddings()
	if err != nil {
		return
	}
	r.Embedding = &jsonData
	r.EmbeddingHash = &hash

	return

}
