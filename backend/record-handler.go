package backend

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

type ListRecordsInput struct {
	ID            uint `query:"id" example:"1" doc:"ID to get" required:"false"`
	ChildrenDepth int  `query:"childrenDepth" example:"2" doc:"Depth to search for children, negative values mean unlimited search" required:"false"`
	ParentDepth   int  `query:"parentDepth" example:"2" doc:"Depth to search for parents, negative values mean unlimited search" required:"false"`
}

type RecordOutput struct {
	Body Record
}

type RecordsOutput struct {
	Body []Record
}

var GetRecordOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/v2/record/{id}",
}

func GetRecord(ctx context.Context, input *struct {
	ID uint `path:"id" example:"1" doc:"ID to delete"`
}) (output *RecordOutput, err error) {
	var records []Record
	records, err = GetRecords(&input.ID, nil, nil)
	if err != nil {
		return
	}
	output = &RecordOutput{
		Body: records[0],
	}
	return
}

var ListRecordsOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/v2/records",
}

func ListRecords(ctx context.Context, input *ListRecordsInput) (output *RecordsOutput, err error) {
	var records []Record
	records, err = GetRecordsFriendly(ctx, input.ID, input.ChildrenDepth, input.ParentDepth)
	output = &RecordsOutput{
		Body: records,
	}
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
	output = &RecordOutput{
		Body: record,
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
		err = huma.Error404NotFound(errorRecordNotFound)
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
		err = huma.Error404NotFound(errorRecordNotFound)
	}
	output = &EmptyOutput{}
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
