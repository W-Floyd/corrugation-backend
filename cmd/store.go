package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

type EntityID int
type ArtifactID int

type Metadata struct {
	Quantity       int      `json:"quantity"`
	Owners         []string `json:"owners"`
	Tags           []string `json:"tags"`
	LastModified   string   `json:"lastmodified"`
	LastModifiedBy string   `json:"lastmodifiedby"`
}

type Entity struct {
	ID          EntityID     `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Artifacts   []ArtifactID `json:"artifacts"`
	Location    EntityID     `json:"location"`
	Metadata    Metadata     `json:"metadata"`
}

type Store struct {
	Entities       map[EntityID]Entity `json:"entities"`
	LastEntityID   EntityID            `json:"lastentityid"`
	LastArtifactID ArtifactID          `json:"lastartifactID"`
}

func updateStore() {
	a, _ := json.MarshalIndent(store, "", "  ")
	d.Write("store.json", a)
}

func dumpStore(c echo.Context) error {
	return c.JSONPretty(http.StatusOK, store, "  ")
}

func updateModification(c echo.Context, eID EntityID) error {

	if _, ok := store.Entities[eID]; ok {
		e := store.Entities[eID]
		e.Metadata.LastModified = time.Now().UTC().Format("2006-01-02 15:04:05.000000") + " UTC"
		e.Metadata.LastModifiedBy = viper.GetString("username")
		e.ID = eID
		store.Entities[eID] = e
		return nil
	}

	return fmt.Errorf("cannot find entity")

}

func resetStore() error {
	store = *new(Store)
	store.Entities = map[EntityID]Entity{}
	updateStore()
	return nil
}
