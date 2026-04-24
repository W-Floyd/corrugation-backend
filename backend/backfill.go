package backend

import (
	"log"

	"gorm.io/gorm"
)

func BackfillEmbeddings() {
	backfillRecordEmbeddings()
	backfillArtifactEmbeddings()
}

func backfillRecordEmbeddings() {
	embeddings, err := gorm.G[Embedding](db).Where("record_id IS NOT NULL AND embed_model = ?", infinityTextModel).Find(dbCtx)
	if err != nil {
		log.Printf("backfill: failed to fetch record embeddings: %v", err)
		return
	}
	embeddedIDs := map[uint]bool{}
	for _, e := range embeddings {
		if e.RecordID != nil {
			embeddedIDs[*e.RecordID] = true
		}
	}

	records, err := GetRecords(nil, nil, nil, nil, nil, []string{"id"})
	if err != nil {
		log.Printf("backfill: failed to fetch records: %v", err)
		return
	}
	recordIDs := make([]uint, len(records))
	for i, r := range records {
		recordIDs[i] = r.ID
	}

	generateMissingRecordEmbeddings(recordIDs, embeddedIDs)
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
