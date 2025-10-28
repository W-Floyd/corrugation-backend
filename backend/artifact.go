package backend

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chai2010/webp"
	"github.com/danielgtaylor/huma/v2"
	"github.com/disintegration/imaging"
	"gorm.io/gorm"
)

type ArtifactInterface interface {
	Store(file huma.FormFile) (err error)
	GetOriginalContents() (output *[]byte, err error)
	GetSmallPreviewContents() (output *[]byte, err error)
	GetLargePreviewContents() (output *[]byte, err error)
	GetOriginalFilename() (output string, err error)
	GetContentType() (output string, err error)
	GetID() (output uint)
	GenerateEmbeddings() (err error)
}

type Artifact struct {
	gorm.Model

	Data             *[]byte
	OriginalFilename *string
	ContentType      *string

	SmallPreviewID *uint
	SmallPreview   *Artifact `gorm:"foreignKey:SmallPreviewID"`
	LargePreviewID *uint
	LargePreview   *Artifact `gorm:"foreignKey:LargePreviewID"`

	RecordID *uint

	Embedding     *[]byte // JSON of embedding data
	EmbeddingHash *string // Hash of JSON of embedding data (to allow caching)

}

func GetArtifactFromDB(ID uint) (artifact Artifact, err error) {
	var artifacts []Artifact
	artifacts, err = gorm.G[Artifact](db).Where("id = ?", ID).Find(dbCtx)
	if err != nil {
		return
	} else if len(artifacts) > 1 {
		err = huma.Error500InternalServerError(errorMoreArtifactsThanExpected)
		return
	} else if len(artifacts) == 0 {
		err = huma.Error404NotFound(errorArtifactNotFound)
		return
	}
	artifact = artifacts[0]
	return
}

func (a *Artifact) GetInterface() (output ArtifactInterface, err error) {
	if a.ContentType == nil {
		err = huma.Error415UnsupportedMediaType("empty content type")
	} else if strings.HasPrefix(*a.ContentType, "image/") {
		i := Image(*a)
		output = &i
	} else {
		switch filepath.Ext(*a.OriginalFilename) {
		case ".png", ".jpeg", ".jpg", ".webp":
			i := Image(*a)
			output = &i
		default:
			err = huma.Error415UnsupportedMediaType("unsupported content type " + *a.ContentType)
			return
		}

	}
	return
}

type Image Artifact

func (i *Image) Store(file huma.FormFile) (err error) {

	b, err := io.ReadAll(file)
	if err != nil {
		return
	}

	c := http.Client{}
	c.Get(infinityAddress + "")

	i.Data = &b

	err = i.GenerateEmbeddings()
	if err != nil {
		return
	}

	i.OriginalFilename = &file.Filename
	i.ContentType = &file.ContentType

	err = i.ComputePreviews()
	if err != nil {
		return
	}

	a := Artifact(*i)

	err = gorm.G[Artifact](db).Create(dbCtx, &a)
	if err != nil {
		return
	}

	*i = Image(a)

	return
}

func (i *Image) computePreview(size int, quality float32) (o *Artifact, err error) {
	preview, err := i.ComputePreview(size, quality)
	if err != nil {
		return
	}
	ct := http.DetectContentType(*preview)
	o = &Artifact{
		Data:        preview,
		ContentType: &ct,
	}
	err = gorm.G[Artifact](db).Create(dbCtx, o)
	return
}

func (i *Image) computeSmallPreview() (o *Artifact, err error) {
	o, err = i.computePreview(625*1000, 70)
	if err != nil {
		return
	}
	i.SmallPreview = o
	i.SmallPreviewID = &o.ID
	if i.ID > 0 {
		var n int
		// n, err := gorm.G[Artifact](db).Where("id = ?", i.ID).Update(dbCtx, "small_preview", *o)
		// if err != nil {
		// 	return
		// } else if n != 1 {
		// 	err = errors.New("affected " + strconv.Itoa(n) + " image small_preview")
		// 	return
		// }
		n, err = gorm.G[Artifact](db).Where("id = ?", i.ID).Update(dbCtx, "small_preview_id", o.ID)
		if err != nil {
			return
		} else if n != 1 {
			err = errors.New("affected " + strconv.Itoa(n) + " image small_preview_id")
			return
		}
	}
	return
}

func (i *Image) computeLargePreview() (o *Artifact, err error) {
	o, err = i.computePreview(1250*1000, 75)
	if err != nil {
		return
	}
	i.LargePreview = o
	i.LargePreviewID = &o.ID
	if i.ID > 0 {
		var n int
		// n, err := gorm.G[Artifact](db).Where("id = ?", i.ID).Update(dbCtx, "large_preview", *o)
		// if err != nil {
		// 	return
		// } else if n != 1 {
		// 	err = errors.New("affected " + strconv.Itoa(n) + " image large_preview")
		// 	return
		// }
		n, err = gorm.G[Artifact](db).Where("id = ?", i.ID).Update(dbCtx, "large_preview_id", o.ID)
		if err != nil {
			return
		} else if n != 1 {
			err = errors.New("affected " + strconv.Itoa(n) + " image large_preview_id")
			return
		}
	}
	return
}

func (i *Image) ComputePreviews() (err error) {

	_, err = i.computeSmallPreview()
	if err != nil {
		return
	}
	_, err = i.computeLargePreview()

	return
}

func (i *Image) ComputePreview(maximumPixelCount int, quality float32) (output *[]byte, err error) {

	img, err := imaging.Decode(bytes.NewBuffer(*i.Data), imaging.AutoOrientation(true))
	if err != nil {
		return
	}

	if img.Bounds().Dx()*img.Bounds().Dy() > maximumPixelCount {
		ratio := float64(img.Bounds().Dx()) / float64(img.Bounds().Dy())
		scaler := math.Sqrt(float64(maximumPixelCount) / (ratio * float64(img.Bounds().Dy()*img.Bounds().Dy())))
		img = imaging.Resize(img, int(float64(img.Bounds().Dx())*scaler), int(float64(img.Bounds().Dy())*scaler), imaging.Lanczos)
	}

	buf := new(bytes.Buffer)

	webp.Encode(buf, img, &webp.Options{Quality: quality})

	o, err := io.ReadAll(buf)
	output = &o
	return
}

func (i *Image) GetOriginalContents() (output *[]byte, err error) {
	if i.Data == nil || len(*i.Data) == 0 {
		err = errors.New("no data in image")
		return
	}
	output = i.Data
	return
}
func (i *Image) GetSmallPreviewContents() (output *[]byte, err error) {
	if i.SmallPreview == nil || i.SmallPreview.Data == nil || len(*i.SmallPreview.Data) == 0 {
		if i.SmallPreviewID != nil {
			var a Artifact
			a, err = GetArtifactFromDB(*i.SmallPreviewID)
			output = a.Data
			return
		} else {
			_, err = i.computeSmallPreview()
			if err != nil {
				return
			}
			output = i.SmallPreview.Data
			return
		}
	}
	output = i.SmallPreview.Data
	return
}
func (i *Image) GetLargePreviewContents() (output *[]byte, err error) {
	if i.LargePreview == nil || i.LargePreview.Data == nil || len(*i.LargePreview.Data) == 0 {
		if i.LargePreviewID != nil {
			var a Artifact
			a, err = GetArtifactFromDB(*i.LargePreviewID)
			output = a.Data
			return
		} else {
			_, err = i.computeLargePreview()
			if err != nil {
				return
			}
			output = i.LargePreview.Data
			return
		}
	}
	output = i.LargePreview.Data
	return
}
func (i *Image) GetOriginalFilename() (output string, err error) {
	if i.OriginalFilename == nil {
		err = errors.New("no filename in image")
		return
	}
	output = *i.OriginalFilename
	return
}
func (i *Image) GetContentType() (output string, err error) {
	if i.ContentType == nil {
		err = errors.New("no content type associated with image")
		return
	}
	output = *i.ContentType
	return
}
func (i *Image) GetID() (output uint) {
	return i.ID
}

func GetArtifactEmbeddings() (e map[uint]struct {
	embedding []float64
	artifact  *Artifact
}, err error) {
	artifacts, err := gorm.G[Artifact](db).Select("id", "embedding", "embedding_hash", "record_id").Find(dbCtx)
	if err != nil {
		return
	}

	e = map[uint]struct {
		embedding []float64
		artifact  *Artifact
	}{}

	for _, a := range artifacts {
		if a.Embedding == nil {
			continue
		}
		var singleE []float64
		singleE, ok := embeddingsCache[*a.EmbeddingHash]
		if !ok {
			singleE = []float64{}
			subErr := json.Unmarshal(*a.Embedding, &singleE)
			if subErr != nil {
				err = subErr
				return
			}
			embeddingsCache[*a.EmbeddingHash] = singleE
		}
		e[a.ID] = struct {
			embedding []float64
			artifact  *Artifact
		}{
			embedding: singleE,
			artifact:  &a,
		}
	}

	return

}
