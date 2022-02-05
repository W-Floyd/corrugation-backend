package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type EntityID int
type ArtifactID int

type Metadata struct {
	Owners       []string
	Tags         []string
	LastModified string
}

type Entity struct {
	Name        string
	Description string
	Artifacts   []ArtifactID
	Location    EntityID
	Metadata    Metadata
}

type Store struct {
	Entities       map[EntityID]Entity
	LastEntityID   EntityID
	LastArtifactID ArtifactID
}

func updateStore() {
	a, _ := json.MarshalIndent(store, "", "  ")
	d.Write("store.json", a)
}

func dumpStore(c echo.Context) error {
	return c.JSONPretty(http.StatusOK, store, "  ")
}

func updateModificationDate(eID EntityID) error {

	if _, ok := store.Entities[eID]; ok {
		e := store.Entities[eID]
		e.Metadata.LastModified = time.Now().UTC().Format("2006-01-02 15:04:05") + " UTC"
		store.Entities[eID] = e
		return nil
	}

	return fmt.Errorf("cannot find entity")

}
