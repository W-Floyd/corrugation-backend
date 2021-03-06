package cmd

import (
	"bytes"
	"math"

	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/gabriel-vasile/mimetype"
	"github.com/labstack/echo/v4"
)

func listArtifacts(c echo.Context) error {

	ch := make(chan struct{})

	artifacts := d.KeysPrefix("artifacts/", ch)

	artifactSlice := chanToSlice(artifacts).([]string)

	for i, name := range artifactSlice {
		artifactSlice[i] = strings.TrimPrefix(name, "artifacts/")
	}

	close(ch)

	return c.JSON(http.StatusOK, artifactSlice)
}

func downloadArtifact(c echo.Context) error {

	id := c.Param("id")

	// If we have the file exactly, give it to them
	if d.Has("artifacts/" + id) {
		return c.File(d.BasePath + "/artifacts/" + id)
	} else { //If we don't have the file exactly, look for a file named that plus an extension

		key := "artifacts/" + id + ".webp"

		if d.Has(key) {
			return c.File(d.BasePath + "/" + key)
		}

		return c.JSON(http.StatusNotFound, "Artifact "+id+" not found")

	}

}

func deleteArtifact(c echo.Context) error {
	id := c.Param("id")

	iID, err := strconv.Atoi(id)

	if err != nil {
		return err
	}

	aID := ArtifactID(iID)

	d.Erase(store.Artifacts[aID].Path)

	return c.JSON(http.StatusNoContent, "Artifact "+id+" does not exist")
}

func uploadArtifact(c echo.Context) error {

	store.LastArtifactID += 1

	// Read form fields
	path := "/artifacts/"

	//-----------
	// Read file
	//-----------
	checkFormFiles([]string{
		"file",
	}, c)

	// Source
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	fullFile, err := ioutil.ReadAll(src)

	mType := mimetype.Detect(fullFile)
	if err != nil {
		return err
	}

	fileName := strconv.Itoa(int(store.LastArtifactID)) + mType.Extension()
	location := path + fileName

	isImage := strings.HasPrefix(mType.String(), "image/")

	if isImage {

		imgSize := 625 * 1000

		img, err := imaging.Decode(bytes.NewBuffer(fullFile), imaging.AutoOrientation(true))
		if err != nil {
			log.Println(err)
			return err
		}

		if img.Bounds().Dx()*img.Bounds().Dy() > imgSize {
			ratio := float64(img.Bounds().Dx()) / float64(img.Bounds().Dy())
			scaler := math.Sqrt(float64(imgSize) / (ratio * float64(img.Bounds().Dy()*img.Bounds().Dy())))
			img = imaging.Resize(img, int(float64(img.Bounds().Dx())*scaler), int(float64(img.Bounds().Dy())*scaler), imaging.Lanczos)
		}

		buf := new(bytes.Buffer)

		webp.Encode(buf, img, &webp.Options{Quality: 70})

		fullFile, err = ioutil.ReadAll(buf)
		if err != nil {
			log.Println(err)
			return err
		}

		fileName = strconv.Itoa(int(store.LastArtifactID)) + ".webp"
		location = path + fileName

	}

	store.Artifacts[store.LastArtifactID] = Artifact{
		Path:  location,
		ID:    store.LastArtifactID,
		Image: isImage,
	}

	if err != nil {
		return err
	}

	err = d.Write(location, fullFile)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, strconv.Itoa(int(store.LastArtifactID)))
}
