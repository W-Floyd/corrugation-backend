package backend

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/nao1215/markdown/mermaid/flowchart"
	"gorm.io/gorm"
)

func GetRecordsFriendly(ctx context.Context, inputID uint, inputChildrenDepth int, inputParentDepth int) (records []Record, err error) {
	var ID *uint
	var ChildrenDepth, ParentDepth *int
	if inputID != 0 {
		ID = &inputID
	}
	if inputChildrenDepth != 0 {
		ChildrenDepth = &inputChildrenDepth
	}
	if inputParentDepth != 0 {
		ParentDepth = &inputParentDepth
	}
	records, err = GetRecords(ID, ChildrenDepth, ParentDepth)
	return
}

func GetRecords(ID *uint, childrenDepth *int, parentDepth *int) (records []Record, err error) {
	if ID == nil {
		if childrenDepth != nil {
			err = errors.New("childrenDepth provided without an ID")
			return
		}
		return gorm.G[Record](db).Find(dbCtx)
	}

	var recordsSearched []Record // This should come back with one value...
	recordsSearched, err = gorm.G[Record](db).Where("id = ?", *ID).Find(dbCtx)
	if err != nil {
		return
	}
	if len(recordsSearched) == 0 {
		err = huma.Error404NotFound(errorRecordNotFound)
	}
	records = append(records, recordsSearched...)

	if childrenDepth != nil {
		var recordPtrs []*Record
		recordPtrs, err = GetChildrenRecurse(*ID, *childrenDepth, 0)

		for _, record := range recordPtrs {
			records = append(records, *record)
		}
	}

	if parentDepth != nil {
		parentSearchCurrentDepth := 0
		searchID := recordsSearched[0].ParentID
		for {
			if searchID == nil {
				break
			}
			parentSearchCurrentDepth += 1
			if *parentDepth > 0 && parentSearchCurrentDepth > *parentDepth {
				break
			} else if parentSearchCurrentDepth > maxSearchDepth {
				err = errors.New("exceeded max search depth on parent")
				return
			}

			var recordsSearched []Record
			recordsSearched, err = gorm.G[Record](db).Where("id = ?", *searchID).Find(dbCtx)
			if err != nil {
				return
			}
			if len(recordsSearched) > 0 {
				records = append(records, recordsSearched...)
				searchID = recordsSearched[0].ParentID
			} else {
				err = errors.New("found no record for " + strconv.FormatUint(uint64(*searchID), 10))
				return
			}

		}
	}

	return

}

func GetChildrenRecurse(parentID uint, searchDepth int, currentDepth int) (records []*Record, err error) {
	if currentDepth > maxSearchDepth {
		err = errors.New("exceeded max search depth on children")
		return
	} else if searchDepth > 0 && currentDepth >= searchDepth {
		return
	}

	var children []Record

	children, err = gorm.G[Record](db).Where("parent_id = ?", parentID).Find(dbCtx)
	if err != nil {
		return
	}
	for _, child := range children {
		records = append(records, &child)
		var subChildren []*Record
		subChildren, err = GetChildrenRecurse(child.ID, searchDepth, currentDepth+1)
		records = append(records, subChildren...)
	}

	return

}

func GetRecordsGraphFriendly(ctx context.Context, inputID uint, inputChildrenDepth int, inputParentDepth int) (graphOutput string, err error) {
	var records []Record
	records, err = GetRecordsFriendly(ctx, inputID, inputChildrenDepth, inputParentDepth)

	recordMap := make(map[uint]*Record)

	fc := flowchart.NewFlowchart(
		io.Discard,
		flowchart.WithTitle("mermaid flowchart builder"),
		flowchart.WithOrientalTopToBottom(),
	)

	for _, record := range records {
		recordMap[record.ID] = &record
	}

	for _, record := range records {
		if record.ParentID != nil {
			if _, ok := recordMap[*record.ParentID]; ok {
				fc.LinkWithArrowHead(
					recordMap[*record.ParentID].PrettyString(),
					recordMap[record.ID].PrettyString(),
				)
			}
		}
	}

	graphOutput = fc.String()
	return
}

func GetRecordsGraphFriendlyNative(ctx context.Context, inputID uint, inputChildrenDepth int, inputParentDepth int) (graphOutput string, err error) {
	var records []Record
	records, err = GetRecordsFriendly(ctx, inputID, inputChildrenDepth, inputParentDepth)
	if err != nil {
		return
	}
	recordMap := make(map[uint]*Record)
	childrenMap := make(map[uint][]uint)
	topLevel := []uint{}

	for _, record := range records {
		recordMap[record.ID] = &record

	}

	for _, record := range records {
		if record.ParentID != nil {
			if _, ok := recordMap[*record.ParentID]; ok {
				childrenMap[*record.ParentID] = append(childrenMap[*record.ParentID], record.ID)
			}
		}
	}

	for _, record := range records {
		if record.ParentID == nil {
			topLevel = append(topLevel, record.ID)
		}
	}

	children := []*opts.TreeData{}

	for _, tl := range topLevel {
		children = append(children, DescendTreeMap(tl, recordMap, childrenMap))
	}

	page := components.NewPage()
	page.AddCharts(
		treeBase(children),
	)

	b := bytes.NewBuffer([]byte{})

	err = page.Render(io.MultiWriter(b))
	if err != nil {
		return
	}
	graphOutput = b.String()
	return
}

func DescendTreeMap(
	rootID uint,
	recordMap map[uint]*Record,
	childrenMap map[uint][]uint,
) (output *opts.TreeData) {
	output = &opts.TreeData{
		Name: recordMap[rootID].PrettyString(),
		Children: func() (out []*opts.TreeData) {
			for _, child := range childrenMap[rootID] {
				out = append(out, DescendTreeMap(child, recordMap, childrenMap))
			}
			return
		}(),
	}
	return
}
func treeBase(treenodes []*opts.TreeData) *charts.Tree {
	graph := charts.NewTree()
	graph.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Width: "100%", Height: "95vh"}),
		//charts.WithTooltipOpts(opts.Tooltip{Show: false}),
	)
	var tree *charts.Tree

	directTreeNodes := []opts.TreeData{}

	for _, node := range treenodes {
		directTreeNodes = append(directTreeNodes, *node)
	}

	if len(directTreeNodes) == 1 {
		tree = graph.AddSeries("tree", directTreeNodes)
	} else {
		tree = graph.AddSeries("tree", []opts.TreeData{
			{
				Name:     topLevelName,
				Children: treenodes,
			},
		})
	}

	tree.
		SetSeriesOptions(
			charts.WithTreeOpts(
				opts.TreeChart{
					Layout:           "orthogonal",
					Orient:           "LR",
					InitialTreeDepth: -1,
					Leaves: &opts.TreeLeaves{
						Label: &opts.Label{Show: opts.Bool(true), Position: "right", Color: "Black"},
					},
				},
			),
			charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Position: "top", Color: "Black"}),
		)
	return graph
}
