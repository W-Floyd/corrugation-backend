package cmd

import (
	"net/http"
	"strconv"

	"github.com/imdario/mergo"
	"github.com/labstack/echo/v4"
)

func parseID(c echo.Context) (int, error) {

	id := c.Param("id")

	idInt, err := strconv.Atoi(id)

	if err != nil {
		return 0, c.JSON(http.StatusBadRequest, "Entity ID "+id+" could not be parsed as an integer")
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

	if err := updateModification(c, store.LastEntityID); err != nil {
		return err
	}

	updateStore()
	return c.JSON(http.StatusOK, strconv.Itoa(int(store.LastEntityID)))
}

func getEntity(c echo.Context) error {

	return checkEntity(c, func(c echo.Context, id EntityID) error {
		return c.JSONPretty(http.StatusOK, store.Entities[EntityID(id)], "  ")
	})

}

func deleteEntity(c echo.Context) error {

	return checkEntity(c, func(c echo.Context, id EntityID) error {
		delete(store.Entities, EntityID(id))
		updateStore()
		return c.NoContent(http.StatusOK)
	})

}

func listEntities(c echo.Context) error {

	idList := []string{}

	for key := range store.Entities {
		idList = append(idList, strconv.Itoa(int(key)))
	}

	return c.JSON(http.StatusOK, idList)
}

func checkEntity(c echo.Context, f func(c echo.Context, id EntityID) error) error {

	id, err := parseID(c)
	if err != nil {
		return err
	}

	eID := EntityID(id)

	if _, ok := store.Entities[eID]; ok {

		return f(c, eID)

	}

	return c.JSON(http.StatusNoContent, "Entity "+strconv.Itoa(id)+" does not exist")

}

func replaceEntity(c echo.Context) error {
	return checkEntity(c, func(c echo.Context, id EntityID) error {
		l := Entity{}

		if err := c.Bind(&l); err != nil {
			return err
		}

		store.Entities[id] = l

		if err := updateModification(c, id); err != nil {
			return err
		}

		updateStore()
		return c.JSON(http.StatusOK, store.Entities[id])
	})
}

func patchEntity(c echo.Context) error {
	return checkEntity(c, func(c echo.Context, id EntityID) error {
		l := Entity{}
		n := store.Entities[id]

		if err := c.Bind(&l); err != nil {
			return err
		}

		if err := mergo.Merge(&n, l, mergo.WithOverride); err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		store.Entities[id] = n

		if err := updateModification(c, id); err != nil {
			return err
		}

		updateStore()
		return c.JSON(http.StatusOK, store.Entities[id])
	})
}

func findContains(id EntityID) (out []EntityID) {
	for key, val := range store.Entities {
		if val.Location == id {
			out = append(out, key)
		}
	}
	return
}

func recurseContains(id EntityID) (out []EntityID) {
	for _, val := range findContains(id) {
		out = append(out, recurseContains(val)...)
	}
	out = append(out, id)
	return
}

func getContains(c echo.Context) error {

	return checkEntity(c, func(c echo.Context, id EntityID) error {
		contains := []EntityID{}

		if c.FormValue("recursive") == "true" {
			for _, val := range findContains(id) {
				contains = append(contains, recurseContains(val)...)
			}
		} else {
			contains = findContains(id)
		}

		return c.JSON(http.StatusOK, contains)
	})

}
