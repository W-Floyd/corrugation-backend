package backend

import (
	"context"
	"net/http"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

type ListRecordsInput struct {
	ID                  uint    `query:"id" example:"1" doc:"ID to get" required:"false"`
	ChildrenDepth       int     `query:"childrenDepth" example:"2" doc:"Depth to search for children, negative values mean unlimited search" required:"false" dependentRequired:"id"`
	ParentDepth         int     `query:"parentDepth" example:"2" doc:"Depth to search for parents, negative values mean unlimited search" required:"false" dependentRequired:"id"`
	Search              string  `query:"search" example:"Lamp" doc:"String to search embeddings with" required:"false"`
	SearchImage         bool    `query:"searchImage" doc:"Use image embeddings in search" required:"false"`
	SearchTextEmbedded  bool    `query:"searchTextEmbedded" doc:"Use text embeddings in search" required:"false"`
	SearchTextSubstring bool    `query:"searchTextSubstring" doc:"Use substring matching in search" required:"false"`
	MinImageScore       float64 `query:"minImageScore" doc:"Minimum image embedding score threshold" required:"false"`
	MinTextScore        float64 `query:"minTextScore" doc:"Minimum text score threshold" required:"false"`
	Timestamps          bool    `query:"timestamps" doc:"Include CreatedAt and UpdatedAt in response" required:"false"`
}

type RecordOutput struct {
	Body RecordResponse
}

type RecordsOutput struct {
	Body []RecordResponse
}

var GetRecordOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/v2/record/{id}",
}

func GetRecord(ctx context.Context, input *struct {
	ID         uint `path:"id" example:"1" doc:"ID to delete"`
	Timestamps bool `query:"timestamps" doc:"Include CreatedAt and UpdatedAt in response" required:"false"`
}) (output *RecordOutput, err error) {
	var records []Record
	records, err = GetRecords(ctx, &input.ID, nil, nil, nil, []struct {
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
	}, nil, false)
	if err != nil {
		return
	}
	output = &RecordOutput{
		Body: toRecordResponse(records[0], input.Timestamps),
	}
	return
}

var ListRecordsOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/v2/records",
}

func ListRecords(ctx context.Context, input *ListRecordsInput) (output *RecordsOutput, err error) {
	var records []Record
	s := NewRecordQuery(input.Search)
	search := &s
	s.SearchImage = input.SearchImage
	s.SearchTextEmbedded = input.SearchTextEmbedded
	s.SearchTextSubstring = input.SearchTextSubstring
	s.ChildrenDepth = input.ChildrenDepth
	s.ParentDepth = input.ParentDepth
	if input.MinImageScore > 0 {
		s.MinImageScore = input.MinImageScore
	}
	if input.MinTextScore > 0 {
		s.MinTextScore = input.MinTextScore
	}
	records, err = GetRecordsFriendly(ctx, input.ID, search)
	responses := make([]RecordResponse, len(records))
	for i, r := range records {
		responses[i] = toRecordResponse(r, input.Timestamps)
	}
	output = &RecordsOutput{Body: responses}
	return
}

var CreateRecordOperation = huma.Operation{
	Method: http.MethodPost,
	Path:   "/api/v2/record",
}

func CreateRecord(ctx context.Context, input *struct {
	Body RecordInput
}) (output *RecordOutput, err error) {
	record, err := input.Body.Convert()
	if err != nil {
		return
	}

	err = gorm.G[Record](db).Create(dbCtx, &record)
	if err != nil {
		return
	}
	if _, genErr := record.GenerateEmbeddings(ctx); genErr != nil {
		Log.Errorw("embedding generation failed", "error", genErr)
	}
	err = nil
	output = &RecordOutput{
		Body: toRecordResponse(record, true),
	}
	return
}

var DeleteRecordOperation = huma.Operation{
	Method: http.MethodDelete,
	Path:   "/api/v2/record/{id}",
}

func DeleteRecord(ctx context.Context, input *struct {
	ID uint `path:"id" example:"1" doc:"ID to delete"`
}) (output *EmptyOutput, err error) {

	chain := gorm.G[Record](db).Where("id = ?", input.ID)

	records, err := chain.Find(dbCtx)
	if err != nil {
		return
	}
	if len(records) == 0 {
		err = huma.Error404NotFound(errorRecordNotFound + " " + strconv.Itoa(int(input.ID)))
		return
	} else if len(records) > 1 {
		err = huma.Error500InternalServerError(errorMoreRecordsThanExpected)
		return
	}

	newParentID := records[0].ParentID

	// Children inherit parentID if available
	if newParentID != nil {
		_, err = gorm.G[Record](db).Where("parent_id = ?", input.ID).Update(dbCtx, "parent_id", *newParentID)
		if err != nil {
			return
		}
	} else {
		_, err = gorm.G[Record](db).Where("parent_id = ?", input.ID).Update(dbCtx, "parent_id", nil)
		if err != nil {
			return
		}
	}

	rowsAffected, err := gorm.G[Record](db).Where("id = ?", input.ID).Delete(dbCtx)
	if err != nil {
		return
	}
	if rowsAffected == 0 {
		err = huma.Error404NotFound(errorRecordNotFound + " " + strconv.Itoa(int(input.ID)))
	}
	output = &EmptyOutput{}
	return
}

var FlushStaleEmbeddingsOperation = huma.Operation{
	Method: http.MethodPost,
	Path:   "/api/v2/embeddings/flush",
}

func FlushStaleEmbeddings(ctx context.Context, _ *struct{}) (output *struct {
	Body struct {
		RecordsFlushed   int64 `json:"recordsFlushed"`
		ArtifactsFlushed int64 `json:"artifactsFlushed"`
	}
}, err error) {
	stale := "embed_model != ? AND embed_model != ?"

	recordsFlushed, err := gorm.G[Embedding](db).Where("record_id IS NOT NULL AND "+stale, infinityTextModel, infinityImageModel).Delete(dbCtx)
	if err != nil {
		return
	}
	artifactsFlushed, err := gorm.G[Embedding](db).Where("artifact_id IS NOT NULL AND "+stale, infinityTextModel, infinityImageModel).Delete(dbCtx)
	if err != nil {
		return
	}

	output = &struct {
		Body struct {
			RecordsFlushed   int64 `json:"recordsFlushed"`
			ArtifactsFlushed int64 `json:"artifactsFlushed"`
		}
	}{Body: struct {
		RecordsFlushed   int64 `json:"recordsFlushed"`
		ArtifactsFlushed int64 `json:"artifactsFlushed"`
	}{
		RecordsFlushed:   int64(recordsFlushed),
		ArtifactsFlushed: int64(artifactsFlushed),
	}}
	return
}

var VisualizeGraphRecordsOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/v2/records/visualize",
}

func VisualizeGraphRecords(ctx context.Context, input *ListRecordsInput) (output *BytesOutput, err error) {

	var graph string
	graph, err = GetRecordsGraphFriendlyNative(ctx, input.ID, input.ChildrenDepth, input.ParentDepth)
	if err != nil {
		return
	}

	output = &BytesOutput{ContentType: "text/html",
		Body: []byte(graph),
	}
	return
}
