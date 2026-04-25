package backend

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/conditional"
	"gorm.io/gorm"
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

	err = a.Store(ctx, f)
	if err != nil {
		Log.Error(err)
		return
	}

	Broadcast()

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
	conditional.Params
	ID uint `path:"id" example:"1" doc:"Artifact ID to get"`
}) (output *BytesOutput, err error) {

	artifact, err := GetArtifactFromDB(input.ID)
	if err != nil {
		return
	}

	embedCtx := context.WithoutCancel(ctx)
	go func() {
		i, iErr := artifact.GetInterface()
		if iErr != nil {
			return
		}
		img, ok := i.(*Image)
		if !ok {
			return
		}
		existing, _ := gorm.G[Embedding](db).Where("artifact_id = ? AND embed_model = ?", artifact.ID, infinityImageModel).Find(dbCtx)
		if len(existing) == 0 {
			if genErr := img.GenerateEmbeddings(embedCtx); genErr != nil {
				Log.Errorw("embedding generation failed", "error", genErr)
			}
		}
	}()

	etag := fmt.Sprintf(`"%d"`, artifact.UpdatedAt.UnixMilli())

	if input.HasConditionalParams() {
		if err = input.PreconditionFailed(etag, artifact.UpdatedAt); err != nil {
			return
		}
	}

	i, err := artifact.GetInterface()

	ob, err := i.GetSmallPreviewContents()
	if err != nil {
		return
	}

	output = &BytesOutput{}
	output.Body = *ob
	output.ContentType = http.DetectContentType(output.Body)
	output.CacheControl = "public, max-age=604800"
	output.ETag = etag

	return
}
