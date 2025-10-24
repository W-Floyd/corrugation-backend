package backend

import (
	"bytes"
	"errors"
	"io"
	"math"
	"net/http"
	"strings"

	"github.com/chai2010/webp"
	"github.com/danielgtaylor/huma/v2"
	"github.com/disintegration/imaging"
	"gorm.io/gorm"
)

type ArtifactInterface interface {
	Store(file huma.FormFile) (err error)
	GetOriginalContents() (output []byte, err error)
	GetSmallPreviewContents() (output []byte, err error)
	GetLargePreviewContents() (output []byte, err error)
	GetOriginalFilename() (output string, err error)
	GetContentType() (output string, err error)
	GetID() (output uint)
}

type Artifact struct {
	gorm.Model

	Data             []byte
	OriginalFilename *string
	ContentType      *string

	SmallPreviewID *uint
	SmallPreview   *Artifact `gorm:"foreignKey:SmallPreviewID"`
	LargePreviewID *uint
	LargePreview   *Artifact `gorm:"foreignKey:LargePreviewID"`

	RecordID *uint
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
		err = huma.Error415UnsupportedMediaType("unsupported content type " + *a.ContentType)
	}
	return
}

type Image Artifact

func (i *Image) Store(file huma.FormFile) (err error) {
	b, err := io.ReadAll(file)
	if err != nil {
		return
	}

	i.Data = b
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

func (i *Image) ComputePreviews() (err error) {

	smallPreview, err := i.ComputePreview(625*1000, 70)
	if err != nil {
		return
	}
	largePreview, err := i.ComputePreview(1250*1000, 75)
	if err != nil {
		return
	}

	ctSmall := http.DetectContentType(smallPreview)
	ctLarge := http.DetectContentType(largePreview)

	smallPreviewImage := &Artifact{
		Data:        smallPreview,
		ContentType: &ctSmall,
	}
	largePreviewImage := &Artifact{
		Data:        largePreview,
		ContentType: &ctLarge,
	}

	err = gorm.G[Artifact](db).Create(dbCtx, smallPreviewImage)
	if err != nil {
		return
	}
	err = gorm.G[Artifact](db).Create(dbCtx, largePreviewImage)
	if err != nil {
		return
	}

	i.LargePreview = largePreviewImage
	i.LargePreviewID = &largePreviewImage.ID
	i.SmallPreview = smallPreviewImage
	i.SmallPreviewID = &smallPreviewImage.ID

	return
}

func (i *Image) ComputePreview(maximumPixelCount int, quality float32) (output []byte, err error) {

	img, err := imaging.Decode(bytes.NewBuffer(i.Data), imaging.AutoOrientation(true))
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

	output, err = io.ReadAll(buf)
	return
}

func (i *Image) GetOriginalContents() (output []byte, err error) {
	if len(i.Data) == 0 {
		err = errors.New("no data in image")
		return
	}
	output = i.Data
	return
}
func (i *Image) GetSmallPreviewContents() (output []byte, err error) {
	if i.SmallPreview == nil || len(i.SmallPreview.Data) == 0 {
		if i.SmallPreviewID != nil {
			var a Artifact
			a, err = GetArtifactFromDB(*i.SmallPreviewID)
			output = a.Data
			return
		} else {
			err = huma.Error500InternalServerError("no small preview in image")
			return
		}
	}
	output = i.SmallPreview.Data
	return
}
func (i *Image) GetLargePreviewContents() (output []byte, err error) {
	if i.LargePreview == nil || len(i.LargePreview.Data) == 0 {
		if i.LargePreviewID != nil {
			var a Artifact
			a, err = GetArtifactFromDB(*i.LargePreviewID)
			output = a.Data
			return
		} else {
			err = huma.Error500InternalServerError("no large preview in image")
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
