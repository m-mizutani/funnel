package model

import "time"

type ImportLog struct {
	TableName  string    `bigquery:"table_name"`
	Timestamp  time.Time `bigquery:"timestamp"`
	ImportedAt time.Time `bigquery:"imported_at"`
}
