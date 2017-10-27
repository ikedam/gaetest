package server

// エンティティの操作

import (
	"net/http"
	"strconv"
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
	g.FlushLocalCache()
	entity.ID = key.IntID()
	if err := g.Get(&entity); err != nil {
		// goon may log if configured inappropriately.
		log.Errorf(ctx, "Failed to re-get Entity: %v, key=%v", err, key)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(
		http.StatusOK,
		&entity,
	)
}

func handlerEntityPut(c echo.Context) error {
	ctx := appengine.NewContext(c.Request())
	g := goon.FromContext(ctx)

	idStr := c.Param("id")

	var id int64
	if _id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
		id = _id
	} else {
		log.Debugf(ctx, "Failed to parse id: %v: %v", idStr, err)
		return c.String(http.StatusNotFound, err.Error())
	}

	var entity Entity
	entity.ID = id

	if err := g.RunInTransaction(func(tg *goon.Goon) error {
		if err := g.Get(&entity); err == datastore.ErrNoSuchEntity {
			log.Debugf(ctx, "Not found: entity %v", id)
			return c.String(http.StatusNotFound, "Not Found")
		} else if err != nil {
			log.Errorf(ctx, "Failed to re-get Entity: %v", err)
			return c.String(http.StatusInternalServerError, err.Error())
		}

		if err := c.Bind(&entity); err != nil {
			log.Warningf(ctx, "Invalid request: %v", err)
			return c.String(http.StatusBadRequest, err.Error())
		}

		if _, err := g.Put(&entity); err != nil {
			log.Errorf(ctx, "Failed to put Entity: %v", err)
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return nil
	}, nil); err != nil {
		return err
	}
	return c.JSON(
		http.StatusOK,
		&entity,
	)
}
