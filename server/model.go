package server

import (
	"time"
)

// Entity は適当なモデルです。
type Entity struct {
	ID            int64     `json:"id" datastore:"-" goon:"id" protectfor:"update"`
	Name          string    `json:"name" datastore:",noindex"`
	ScheduledDate time.Time `json:"scheduledDate" datastore:",noindex"`
	CreatedAt     time.Time `json:"createdAt" protectfor:"update"`
}
