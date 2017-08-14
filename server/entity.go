package server

// エンティティの操作

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/mjibson/goon"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

func handlerEntityListGet(c echo.Context) error {
	ctx := appengine.NewContext(c.Request())
	var entityList []Entity
	g := goon.FromContext(ctx)
	q := datastore.NewQuery("Entity").Order("-CreatedAt")
	if _, err := g.GetAll(q, &entityList); err != nil {
		log.Errorf(ctx, "Failed to query Entity: %v", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(
		http.StatusOK,
		&entityList,
	)
}

func handlerEntityPost(c echo.Context) error {
	ctx := appengine.NewContext(c.Request())
	var entity Entity
	if err := c.Bind(&entity); err != nil {
		log.Warningf(ctx, "Invalid request: %v", err)
		return c.String(http.StatusBadRequest, err.Error())
	}
	entity.ID = 0
	entity.CreatedAt = time.Now().UTC()
	g := goon.FromContext(ctx)

	key, err := g.Put(&entity)
	if err != nil {
		log.Errorf(ctx, "Failed to put Entity: %v", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if err := datastore.Get(ctx, key, &entity); err != nil {
		log.Errorf(ctx, "Failed to re-get Entity: %v, %v", key, err)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(
		http.StatusOK,
		&entity,
	)
}
