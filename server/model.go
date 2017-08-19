package server

import (
	"time"
)

// Entity は適当なモデルです。
type Entity struct {
	ID        int64     `json:"id" datastore:"-" goon:"id"`
	Name      string    `json:"name" datastore:",noindex"`
	CreatedAt time.Time `json:"createdAt"`
}
