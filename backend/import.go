package backend

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

type legacyMetadata struct {
	Quantity       int      `json:"quantity"`
	Owners         []string `json:"owners"`
	Tags           []string `json:"tags"`
	LastModified   string   `json:"lastmodified"`
	LastModifiedBy string   `json:"lastmodifiedby"`
	IsLabeled      bool     `json:"islabeled"`
}

type legacyEntity struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Artifacts   []int          `json:"artifacts"`
	Location    int            `json:"location"`
	Metadata    legacyMetadata `json:"metadata"`
}

type legacyStore struct {
	Entities map[string]legacyEntity `json:"entities"`
}

var ImportOperation = huma.Operation{
	Method:        http.MethodPost,
	Path:          "/api/import",
	DefaultStatus: http.StatusOK,
}

type ImportResult struct {
	EntitiesImported  int `json:"entitiesImported"`
	ArtifactsImported int `json:"artifactsImported"`
}

func Import(ctx context.Context, input *struct {
	Reset   bool `query:"reset" doc:"Clear all existing data before importing"`
	RawBody huma.MultipartFormFiles[struct {
		File huma.FormFile `form:"file" required:"true"`
	}]
}) (output *struct{ Body ImportResult }, err error) {

	f := input.RawBody.Data().File

	result, err := ImportFromReader(ctx, f, input.Reset)
	if err != nil {
		return
	}

	output = &struct{ Body ImportResult }{Body: result}
	return
}

// ImportFromReader imports legacy data from a tar.gz reader.
func ImportFromReader(ctx context.Context, r io.Reader, reset bool) (result ImportResult, err error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		err = huma.Error400BadRequest("not a gzip file: " + err.Error())
		return
	}
	defer gz.Close()

	// Read all tar entries into memory
	artifactData := map[int][]byte{}
	var store legacyStore

	tr := tar.NewReader(gz)
	for {
		hdr, e := tr.Next()
		if e == io.EOF {
			break
		}
		if e != nil {
			err = e
			return
		}
		name := hdr.Name
		base := filepath.Base(name)

		if base == "store.json" {
			data, e := io.ReadAll(tr)
			if e != nil {
				err = e
				return
			}
			if e := json.Unmarshal(data, &store); e != nil {
				err = huma.Error400BadRequest("invalid store.json: " + e.Error())
				return
			}
			continue
		}

		dir := filepath.Base(filepath.Dir(name))
		if dir == "artifacts" && strings.HasSuffix(base, ".webp") {
			idStr := strings.TrimSuffix(base, ".webp")
			id, e := strconv.Atoi(idStr)
			if e != nil {
				continue
			}
			data, e := io.ReadAll(tr)
			if e != nil {
				err = e
				return
			}
			artifactData[id] = data
		}
	}

	if reset {
		tables := []string{"record_tags", "artifacts", "records", "tags"}
		for _, table := range tables {
			if e := db.Exec("DELETE FROM " + table).Error; e != nil {
				err = e
				return
			}
			db.Exec("DELETE FROM sqlite_sequence WHERE name = ?", table)
		}
	}

	if store.Entities == nil {
		err = huma.Error400BadRequest("store.json not found or empty")
		return
	}

	// Import artifacts first (entities reference them by ID)
	ct := "image/webp"
	for id, data := range artifactData {
		d := make([]byte, len(data))
		copy(d, data)
		filename := fmt.Sprintf("%d.webp", id)
		a := Artifact{
			Data:             &d,
			ContentType:      &ct,
			OriginalFilename: &filename,
		}
		a.ID = uint(id)
		if e := db.Create(&a).Error; e != nil {
			// Skip if already exists
			continue
		}
		result.ArtifactsImported++
	}

	// Import entities
	for _, le := range store.Entities {
		r := Record{}
		r.ID = uint(le.ID)
		r.Description = strPtr(le.Description)

		if le.Metadata.IsLabeled {
			r.Label = strPtr(le.Name)
		} else {
			r.Title = strPtr(le.Name)
		}

		if le.Location != 0 {
			v := uint(le.Location)
			r.ParentID = &v
		}

		if le.Metadata.Quantity != 0 {
			v := uint(le.Metadata.Quantity)
			r.Quantity = &v
		}

		if le.Metadata.LastModifiedBy != "" {
			r.LastModifiedBy = strPtr(le.Metadata.LastModifiedBy)
		}

		// Link artifacts
		for _, aid := range le.Artifacts {
			var a Artifact
			if tx := db.First(&a, aid); tx.Error == nil {
				r.Artifacts = append(r.Artifacts, &a)
			}
		}

		// Tags
		for _, tagTitle := range le.Metadata.Tags {
			if tagTitle == "" {
				continue
			}
			var found []Tag
			if tx := db.Where("title = ?", tagTitle).Find(&found); tx.Error == nil && len(found) > 0 {
				r.Tags = append(r.Tags, &found[0])
			} else {
				t := Tag{Title: tagTitle}
				if e := gorm.G[Tag](db).Create(dbCtx, &t); e == nil {
					r.Tags = append(r.Tags, &t)
				}
			}
		}

		if e := db.Create(&r).Error; e != nil {
			continue
		}
		result.EntitiesImported++
	}

	Broadcast()
	return
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
