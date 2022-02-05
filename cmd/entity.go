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
		return 0, c.String(http.StatusBadRequest, "Entity ID "+id+" could not be parsed as an integer")
	}

	return idInt, nil
}

func createEntity(c echo.Context) error {
	store.LastEntityID += 1

	loc := Entity{}
	if hasForm("name", c) {
		loc.Name = c.FormValue("name")
	}

	if hasForm("description", c) {
		loc.Description = c.FormValue("description")
	}

	store.Entities[store.LastEntityID] = loc

	updateStore()
	return c.String(http.StatusOK, strconv.Itoa(int(store.LastEntityID)))
}

func getEntity(c echo.Context) error {

	id, err := parseID(c)
	if err != nil {
		return err
	}

	if val, ok := store.Entities[EntityID(id)]; ok {
		return c.JSONPretty(http.StatusOK, val, "  ")
	}

	return c.String(http.StatusNotFound, "Entity "+strconv.Itoa(id)+" does not exist")

}

func deleteEntity(c echo.Context) error {

	id, err := parseID(c)
	if err != nil {
		return err
	}

	if _, ok := store.Entities[EntityID(id)]; ok {
		delete(store.Entities, EntityID(id))
		updateStore()
		return c.NoContent(http.StatusOK)
	}

	return c.String(http.StatusNoContent, "Entity "+strconv.Itoa(id)+" does not exist")

}

func listEntities(c echo.Context) error {

	idList := []string{}

	for key := range store.Entities {
		idList = append(idList, strconv.Itoa(int(key)))
	}

	return c.JSON(http.StatusOK, idList)
}

func updateEntity(c echo.Context) error {

	id, err := parseID(c)
	if err != nil {
		return err
	}

	eID := EntityID(id)

	if _, ok := store.Entities[eID]; ok {

		l := Entity{}

		if err := c.Bind(&l); err != nil {
			return err
		}

		store.Entities[eID] = l

		if err := updateModificationDate(eID); err != nil {
			return err
		}

		updateStore()
		return c.JSON(http.StatusOK, l)

	}

	return c.String(http.StatusNoContent, "Entity "+strconv.Itoa(id)+" does not exist")

}
