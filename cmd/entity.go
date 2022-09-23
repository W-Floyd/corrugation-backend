package cmd

import (
	"net/http"
	"sort"
	"strconv"

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

	loc := Entity{}

	if err := c.Bind(&loc); err != nil {
		return err
	}

	if hasForm("name", c) {
		loc.Name = c.FormValue("name")
	}

	if hasForm("description", c) {
		loc.Description = c.FormValue("description")
	}

	if loc.ID == 0 {
		loc.ID = store.LastEntityID() + 1
	}

	if _, exists := store.Entities[loc.ID]; exists {
		if store.Entities[loc.ID].Metadata.IsLabeled {
			return c.JSON(http.StatusConflict, "Entity exists with labeled entity "+strconv.Itoa(int(loc.ID)))
		}
		e := store.Entities[loc.ID]
		e.ID = store.LastEntityID() + 1
		store.Entities[e.ID] = e
	}

	store.Entities[loc.ID] = loc

	if err := updateModification(loc.ID); err != nil {
		return err
	}

	updateStore()
	return c.JSON(http.StatusOK, strconv.Itoa(int(loc.ID)))
}

func getEntity(c echo.Context) error {

	return checkEntity(c, func(c echo.Context, id EntityID) error {
		return c.JSON(http.StatusOK, store.Entities[EntityID(id)])
	})

}

func deleteEntity(c echo.Context) error {

	return checkEntity(c, func(c echo.Context, id EntityID) error {
		for key := range store.Entities {
			if store.Entities[key].Location == id {
				l := store.Entities[key]
				l.Location = store.Entities[id].Location
				store.Entities[key] = l
			}
		}
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

		if err := updateModification(id); err != nil {
			return err
		}

		updateStore()
		return c.JSON(http.StatusOK, store.Entities[id])
	})
}

func patchEntity(c echo.Context) error {
	return checkEntity(c, func(c echo.Context, id EntityID) error {
		n := store.Entities[id]

		if err := c.Bind(&n); err != nil {
			return err
		}

		store.Entities[id] = n

		if err := updateModification(id); err != nil {
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

func getContainsFull(c echo.Context) error {

	id, err := parseID(c)
	if err != nil {
		return err
	}

	eID := EntityID(id)

	contains := findContains(eID)

	l := []Entity{}
	for _, val := range contains {
		l = append(l, store.Entities[val])
	}

	return c.JSON(http.StatusOK, sortEntities(l))

}

func getContainsFullRecursive(c echo.Context) error {

	id, err := parseID(c)
	if err != nil {
		return err
	}

	eID := EntityID(id)

	contains := []EntityID{}

	for _, val := range findContains(eID) {
		contains = append(contains, recurseContains(val)...)
	}

	l := []Entity{}
	for _, val := range contains {
		l = append(l, store.Entities[val])
	}

	return c.JSON(http.StatusOK, sortEntities(l))

}

func getEntities(c echo.Context) error {
	return c.JSON(http.StatusOK, store.Entities)
}

func findEntitiesWithChildren(c echo.Context) error {

	m := map[EntityID][]EntityID{}

	for key, val := range store.Entities {
		m[val.Location] = append(m[val.Location], key)
	}

	l := []EntityID{}

	for key := range m {
		l = append(l, key)
	}

	return c.JSON(http.StatusOK, l)
}

func findEntitiesWithChildrenFull(c echo.Context) error {

	m := map[EntityID][]EntityID{}

	for key, val := range store.Entities {
		m[val.Location] = append(m[val.Location], key)
	}

	l := []Entity{}

	for key := range m {
		if val, ok := store.Entities[key]; ok {
			l = append(l, val)
		}
	}

	return c.JSON(http.StatusOK, sortEntities(l))
}

func sortEntities(s []Entity) (out []Entity) {
	out = s
	sort.SliceStable(out, func(p, q int) bool {
		if out[p].Name == out[q].Name {
			return out[p].Description < out[q].Description
		}
		return out[p].Name < out[q].Name
	})
	return
}

func firstId(c echo.Context) error {
	candidates := unlabeledIDs()
	candidates = append(candidates, emptyIDs()...)

	min := store.LastEntityID() + 1

	for _, val := range candidates {
		if val < min {
			min = val
		}
	}

	return c.JSON(http.StatusOK, min)

}
