package backend

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

var ExportOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/export",
}

func Export(ctx context.Context, _ *struct{}) (output *BytesOutput, err error) {
	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)

	// --- Records ---
	records, err := GetRecords(nil, nil, nil, nil, []struct {
		q string
		h func(db gorm.PreloadBuilder) error
	}{
		{q: "Artifacts", h: func(db gorm.PreloadBuilder) error { return nil }},
		{q: "Tags", h: func(db gorm.PreloadBuilder) error { return nil }},
	}, nil)
	if err != nil {
		return
	}

	for i := range records {
		entity, e2 := records[i].ToEntity()
		if e2 != nil {
			err = e2
			return
		}
		data, e2 := json.MarshalIndent(entity, "", "  ")
		if e2 != nil {
			err = e2
			return
		}
		w, e2 := zw.Create("records/" + strconv.FormatUint(uint64(records[i].ID), 10) + ".json")
		if e2 != nil {
			err = e2
			return
		}
		w.Write(data)
	}

	// --- Artifacts (exclude previews) ---
	// Collect all preview IDs so we can skip them
	type previewRow struct {
		SmallPreviewID *uint
		LargePreviewID *uint
	}
	var previews []previewRow
	if tx := db.Model(&Artifact{}).Select("small_preview_id", "large_preview_id").Scan(&previews); tx.Error != nil {
		err = tx.Error
		return
	}
	previewIDs := map[uint]struct{}{}
	for _, p := range previews {
		if p.SmallPreviewID != nil {
			previewIDs[*p.SmallPreviewID] = struct{}{}
		}
		if p.LargePreviewID != nil {
			previewIDs[*p.LargePreviewID] = struct{}{}
		}
	}

	artifacts, err := gorm.G[Artifact](db).Find(dbCtx)
	if err != nil {
		return
	}

	for _, a := range artifacts {
		if _, isPreview := previewIDs[a.ID]; isPreview {
			continue
		}

		// Metadata JSON
		type artifactMeta struct {
			ID               uint    `json:"id"`
			ContentType      *string `json:"contentType,omitempty"`
			OriginalFilename *string `json:"originalFilename,omitempty"`
			RecordID         *uint   `json:"recordId,omitempty"`
		}
		meta := artifactMeta{
			ID:               a.ID,
			ContentType:      a.ContentType,
			OriginalFilename: a.OriginalFilename,
			RecordID:         a.RecordID,
		}
		metaData, e2 := json.MarshalIndent(meta, "", "  ")
		if e2 != nil {
			err = e2
			return
		}
		id := strconv.FormatUint(uint64(a.ID), 10)
		w, e2 := zw.Create("artifacts/" + id + ".json")
		if e2 != nil {
			err = e2
			return
		}
		w.Write(metaData)

		// Binary data
		if a.Data != nil && len(*a.Data) > 0 {
			ext := ".bin"
			if a.OriginalFilename != nil && filepath.Ext(*a.OriginalFilename) != "" {
				ext = filepath.Ext(*a.OriginalFilename)
			} else if a.ContentType != nil {
				switch *a.ContentType {
				case "image/webp":
					ext = ".webp"
				case "image/png":
					ext = ".png"
				case "image/jpeg":
					ext = ".jpg"
				}
			}
			w, e2 = zw.Create("artifacts/" + id + ext)
			if e2 != nil {
				err = e2
				return
			}
			w.Write(*a.Data)
		}
	}

	if err = zw.Close(); err != nil {
		return
	}

	output = &BytesOutput{
		ContentType: "application/zip",
		Body:        buf.Bytes(),
	}
	return
}
