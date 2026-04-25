package backend

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

var GetEmbeddingProgressOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/v2/embeddings/progress",
}

type EmbeddingProgress struct {
	Pending    int64 `json:"pending"`
	Processing int64 `json:"processing"`
	Done       int64 `json:"done"`
	Failed     int64 `json:"failed"`
	Total      int64 `json:"total"`
}

func GetEmbeddingProgress(ctx context.Context, _ *struct{}) (output *struct{ Body EmbeddingProgress }, err error) {
	username := UsernameFromContext(ctx)
	uc, _ := loadUser(username)

	q := db.Model(&EmbeddingJob{})
	if username != "" && uc.ID > 0 {
		q = q.Where("owner_id = ?", uc.ID)
	}

	type statusCount struct {
		Status string
		Count  int64
	}
	var counts []statusCount
	if err = q.Select("status, COUNT(*) as count").Group("status").Scan(&counts).Error; err != nil {
		return
	}

	var p EmbeddingProgress
	for _, c := range counts {
		switch c.Status {
		case JobStatusPending:
			p.Pending = c.Count
		case JobStatusProcessing:
			p.Processing = c.Count
		case JobStatusDone:
			p.Done = c.Count
		case JobStatusFailed:
			p.Failed = c.Count
		}
	}
	p.Total = p.Pending + p.Processing + p.Done + p.Failed

	output = &struct{ Body EmbeddingProgress }{Body: p}
	return
}

var GetSearchEmbeddingProgressOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/v2/embeddings/search-progress",
}

type SearchEmbeddingProgress struct {
	Record struct {
		Complete []int64 `json:"complete"`
		Pending  []int64 `json:"pending"`
	} `json:"record"`
	Artifact struct {
		Complete []int64 `json:"complete"`
		Pending  []int64 `json:"pending"`
	} `json:"artifact"`
	Ready bool `json:"ready"`
}

func GetSearchEmbeddingProgress(ctx context.Context, input *struct {
	ID                 int  `query:"id" required:"false" default:"-1"`
	Global             bool `query:"global" required:"false"`
	ChildrenDepth      int  `query:"childrenDepth" required:"false"`
	SearchImage        bool `query:"searchImage" required:"false"`
	SearchTextEmbedded bool `query:"searchTextEmbedded" required:"false"`
}) (output *struct{ Body SearchEmbeddingProgress }, err error) {
	// Resolve scope
	var id *uint
	if !input.Global {
		if input.ID >= 0 {
			v := uint(input.ID)
			id = &v
		} else {
			var zero uint = 0
			id = &zero
		}
	}
	var childrenDepth *int
	if input.ChildrenDepth != 0 {
		childrenDepth = &input.ChildrenDepth
	}

	records, _, err := GetRecords(ctx, id, childrenDepth, nil, nil, []struct {
		q string
		h func(db gorm.PreloadBuilder) error
	}{
		{q: "Artifacts", h: func(db gorm.PreloadBuilder) error { db.Select("id", "record_id"); return nil }},
	}, nil)
	if err != nil {
		return
	}

	recordIDs := make([]uint, 0, len(records))
	artifactIDs := make([]uint, 0)
	for _, r := range records {
		recordIDs = append(recordIDs, r.ID)
		for _, a := range r.Artifacts {
			if a != nil {
				artifactIDs = append(artifactIDs, a.ID)
			}
		}
	}

	uc, _ := loadUser(UsernameFromContext(ctx))
	textModel, imageModel, _, _ := effectiveInfinityConfig(uc)

	var p SearchEmbeddingProgress

	if input.SearchTextEmbedded && len(recordIDs) > 0 {
		var indexed, pending int64
		db.Model(&Embedding{}).
			Where("record_id IN ? AND embed_model = ?", recordIDs, textModel).
			Count(&indexed)
		db.Model(&EmbeddingJob{}).
			Where("job_type = ? AND target_id IN ? AND embed_model = ? AND status IN ?",
				JobTypeRecord, recordIDs, textModel, []string{JobStatusPending, JobStatusProcessing}).
			Count(&pending)
		p.Record.Complete = append(p.Record.Complete, indexed)
		p.Record.Pending = append(p.Record.Pending, pending)
	}

	if input.SearchImage && len(artifactIDs) > 0 {
		var indexed, pending int64
		db.Model(&Embedding{}).
			Where("artifact_id IN ? AND embed_model = ?", artifactIDs, imageModel).
			Count(&indexed)
		db.Model(&EmbeddingJob{}).
			Where("job_type = ? AND target_id IN ? AND embed_model = ? AND status IN ?",
				JobTypeArtifact, artifactIDs, imageModel, []string{JobStatusPending, JobStatusProcessing}).
			Count(&pending)
		p.Artifact.Complete = append(p.Artifact.Complete, indexed)
		p.Artifact.Pending = append(p.Artifact.Pending, pending)
	}

	p.Ready = true
	if input.SearchTextEmbedded && len(p.Record.Pending) > 0 && p.Record.Pending[0] > 0 {
		p.Ready = false
	}
	if input.SearchImage && len(p.Artifact.Pending) > 0 && p.Artifact.Pending[0] > 0 {
		p.Ready = false
	}

	output = &struct{ Body SearchEmbeddingProgress }{Body: p}
	return
}
