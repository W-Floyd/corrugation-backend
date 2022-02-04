package cmd

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/labstack/echo"
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

func findArtifact(c echo.Context, id string) (string, error) {

	ch := make(chan struct{})

	artifacts := d.KeysPrefix("artifacts/"+id, ch)

	artifactSlice := chanToSlice(artifacts).([]string)

	close(ch)

	if len(artifactSlice) < 1 {
		return "", c.String(http.StatusNotFound, "Artifact "+id+" not found")
	}

	if len(artifactSlice) > 1 {
		return "", c.String(http.StatusNotFound, "More than one artifact found for "+id)
	}

	return d.BasePath + "/" + artifactSlice[0], nil
}

func downloadArtifact(c echo.Context) error {

	id := c.Param("id")

	// If we have the file exactly, give it to them
	if d.Has("artifacts/" + id) {
		return c.File(d.BasePath + "/artifacts/" + id)
	} else { //If we don't have the file exactly, look for a file named that plus an extension
		path, err := findArtifact(c, id)
		if err != nil {
			return err
		}
		return c.File(path)
	}

}

func deleteArtifact(c echo.Context) error {
	id := c.Param("id")

	// If we have the file exactly, delete it
	if d.Has("artifacts/" + id) {
		d.Erase("artifacts/" + id)
		return c.NoContent(http.StatusOK)
	}

	return c.String(http.StatusNoContent, "Artifact "+id+" does not exist")
}

func uploadArtifact(c echo.Context) error {

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

	store.LastArtifactID += 1

	fileName := strconv.Itoa(store.LastArtifactID) + mType.Extension()
	location := path + fileName

	err = d.Write(location, fullFile)
	if err != nil {
		return err
	}

	return c.String(http.StatusOK, strconv.Itoa(store.LastArtifactID))
}
