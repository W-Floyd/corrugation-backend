package backend

import (
	"context"
	"log"

	"gorm.io/gorm"
)

func BackfillEmbeddings() {
	backfillRecordEmbeddings()
	backfillArtifactEmbeddings()
}

func backfillRecordEmbeddings() {
	records, err := GetRecords(dbCtx, nil, nil, nil, nil, nil, []string{"id", "title", "label", "description", "last_modified_by"})
	if err != nil {
		log.Printf("backfill: failed to fetch records: %v", err)
		return
	}

	byUser := map[string][]Record{}
	for _, r := range records {
		username := ""
		if r.LastModifiedBy != nil {
			username = *r.LastModifiedBy
		}
		byUser[username] = append(byUser[username], r)
	}

	for username, userRecords := range byUser {
		cfg, _ := loadUserConfig(username)
		textModel, _, _, docPrefix := effectiveInfinityConfig(cfg)
		ctx := context.WithValue(dbCtx, usernameContextKey, username)
		backfillRecordEmbeddingsForUser(ctx, textModel, docPrefix, userRecords)
	}
}

func backfillRecordEmbeddingsForUser(ctx context.Context, textModel, docPrefix string, records []Record) {
	recordIDs := make([]uint, len(records))
	for i, r := range records {
		recordIDs[i] = r.ID
	}

	embeddings, err := gorm.G[Embedding](db).Where("record_id IN ? AND embed_model = ?", recordIDs, textModel).Find(dbCtx)
	if err != nil {
		log.Printf("backfill: failed to fetch embeddings for model %q: %v", textModel, err)
		return
	}
	storedHash := map[uint]string{}
	for _, e := range embeddings {
		if e.RecordID != nil {
			storedHash[*e.RecordID] = e.Hash
		}
	}

	embeddedIDs := map[uint]bool{}
	for _, r := range records {
		text := recordEmbeddingText(r)
		if text == "" {
			continue
		}
		if storedHash[r.ID] == InputHash(docPrefix+text) {
			embeddedIDs[r.ID] = true
		}
	}

	generateMissingRecordEmbeddings(ctx, recordIDs, embeddedIDs)
}

func backfillArtifactEmbeddings() {
	embeddings, err := gorm.G[Embedding](db).Where("artifact_id IS NOT NULL AND embed_model = ?", infinityImageModel).Find(dbCtx)
	if err != nil {
		log.Printf("backfill: failed to fetch artifact embeddings: %v", err)
		return
	}
	embeddedIDs := map[uint]bool{}
	for _, e := range embeddings {
		if e.ArtifactID != nil {
			embeddedIDs[*e.ArtifactID] = true
		}
	}

	artifacts, err := gorm.G[Artifact](db).Select("id").Find(dbCtx)
	if err != nil {
		log.Printf("backfill: failed to fetch artifacts: %v", err)
		return
	}
	artifactIDs := make([]uint, len(artifacts))
	for i, a := range artifacts {
		artifactIDs[i] = a.ID
	}

	generateMissingArtifactEmbeddings(artifactIDs, embeddedIDs)
}
