package backend

import (
	"errors"

	"gorm.io/gorm"
)

type Embedding struct {
	gorm.Model

	RecordID   *uint  `gorm:"index"`
	ArtifactID *uint  `gorm:"index"`
	EmbedModel string `gorm:"not null;index"`
	Data       []byte `gorm:"not null"`
	Hash       string `gorm:"not null"`
}

func saveEmbedding(recordID *uint, artifactID *uint, e Embeddings, model string, input string) error {
	if recordID == nil && artifactID == nil {
		return errors.New("saveEmbedding: both recordID and artifactID are nil")
	}

	hash, data, err := e.MarshalEmbeddings(input)
	if err != nil {
		return err
	}

	var existing []Embedding
	if recordID != nil {
		existing, err = gorm.G[Embedding](db).Where("record_id = ? AND embed_model = ?", *recordID, model).Find(dbCtx)
	} else {
		existing, err = gorm.G[Embedding](db).Where("artifact_id = ? AND embed_model = ?", *artifactID, model).Find(dbCtx)
	}
	if err != nil {
		return err
	}

	if len(existing) > 0 {
		if _, err = gorm.G[Embedding](db).Where("id = ?", existing[0].ID).Update(dbCtx, "data", data); err != nil {
			return err
		}
		_, err = gorm.G[Embedding](db).Where("id = ?", existing[0].ID).Update(dbCtx, "hash", hash)
		return err
	}

	entry := Embedding{
		RecordID:   recordID,
		ArtifactID: artifactID,
		EmbedModel: model,
		Data:       data,
		Hash:       hash,
	}
	return gorm.G[Embedding](db).Create(dbCtx, &entry)
}
