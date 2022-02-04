package cmd

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func parseID(c echo.Context) (int, error) {

	id := c.Param("id")

	idInt, err := strconv.Atoi(id)

	if err != nil {
		return 0, c.String(http.StatusBadRequest, "Location ID "+id+" could not be parsed as an integer")
	}

	return idInt, nil
}

func createLocation(c echo.Context) error {
	store.LastLocationID += 1

	loc := Location{}
	if hasForm("name", c) {
		loc.Name = c.FormValue("name")
	}

	if hasForm("description", c) {
		loc.Description = c.FormValue("description")
	}

	store.Locations[store.LastLocationID] = loc

	updateStore()
	return c.String(http.StatusOK, strconv.Itoa(store.LastLocationID))
}

func getLocation(c echo.Context) error {

	id, err := parseID(c)
	if err != nil {
		return err
	}

	if val, ok := store.Locations[id]; ok {
		return c.JSONPretty(http.StatusOK, val, "  ")
	}

	return c.String(http.StatusNotFound, "Location "+strconv.Itoa(id)+" does not exist")

}

func deleteLocation(c echo.Context) error {

	id, err := parseID(c)
	if err != nil {
		return err
	}

	if _, ok := store.Locations[id]; ok {
		delete(store.Locations, id)
		updateStore()
		return c.NoContent(http.StatusOK)
	}

	return c.String(http.StatusNoContent, "Location "+strconv.Itoa(id)+" does not exist")

}

func listLocations(c echo.Context) error {

	idList := []string{}

	for key, _ := range store.Locations {
		idList = append(idList, strconv.Itoa(key))
	}

	return c.JSON(http.StatusOK, idList)
}

func updateLocation(c echo.Context) error {

	id, err := parseID(c)
	if err != nil {
		return err
	}

	if _, ok := store.Locations[id]; ok {

		l := new(Location)

		if err := c.Bind(l); err != nil {
			return err
		}

		store.Locations[id] = *l

		updateStore()
		return c.NoContent(http.StatusOK)

	}

	return c.String(http.StatusNoContent, "Location "+strconv.Itoa(id)+" does not exist")

}
