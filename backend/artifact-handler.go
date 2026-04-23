package backend

import (
	"context"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

var CreateArtifactOperation = huma.Operation{
	Method: http.MethodPost,
	Path:   "/api/v2/artifact",
}

func CreateArtifact(ctx context.Context, input *struct {
	RawBody huma.MultipartFormFiles[struct {
		File huma.FormFile `form:"file" required:"true"`
	}]
}) (output *UIntOutput, err error) {

	f := input.RawBody.Data().File

	var a ArtifactInterface

	if strings.HasPrefix(f.ContentType, "image/") {
		a = &Image{}
	} else {
		switch filepath.Ext(f.Filename) {
		case ".png", ".jpeg", ".jpg", ".webp":
			a = &Image{}
		default:
			err = huma.Error415UnsupportedMediaType("unsupported media type " + f.ContentType)
			return
		}
	}

	err = a.Store(f)
	if err != nil {
		log.Println(err)
		return
	}

	output = &UIntOutput{
		Body: a.GetID(),
	}

	return
}

var GetArtifactOperation = huma.Operation{
	Method: http.MethodGet,
	Path:   "/api/v2/artifact/{id}",
}

func GetArtifact(ctx context.Context, input *struct {
	ID uint `path:"id" example:"1" doc:"Artifact ID to get"`
}) (output *BytesOutput, err error) {

	artifact, err := GetArtifactFromDB(input.ID)
	if err != nil {
		return
	}

	i, err := artifact.GetInterface()
	if err != nil {
		return
	}

	output = &BytesOutput{}

	ob, err := i.GetSmallPreviewContents()
	if err != nil {
		return
	}

	output.Body = *ob
	output.ContentType = http.DetectContentType(output.Body)
	output.CacheControl = "public, max-age=604800"

	return
}
