package model

import "time"

type ImportLog struct {
	LatestRecord time.Time
	CheckedAt    time.Time
}
