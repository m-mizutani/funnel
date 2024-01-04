package bq

import (
	"context"
	"time"

	"github.com/m-mizutani/pacman/pkg/domain/interfaces"
	"github.com/m-mizutani/pacman/pkg/domain/model"
)

type mockRecord struct {
	TableName string    `bigquery:"table_name"`
	Timestamp time.Time `bigquery:"timestamp"`
}

type Mock struct {
	InsertedData []any
	Records      []*model.ImportLog
}

var _ interfaces.BigQuery = &Mock{}

func NewMock() *Mock {
	return &Mock{}
}

func (x *Mock) Migrate(ctx context.Context, tableName string, schema any) error {
	return nil
}

func (x *Mock) Insert(ctx context.Context, tableName string, data any) error {
	x.InsertedData = append(x.InsertedData, data)
	return nil
}

func (x *Mock) PutImportLog(ctx context.Context, log *model.ImportLog) error {
	x.Records = append(x.Records, log)
	return nil
}

func (x *Mock) GetLatestImportLog(ctx context.Context, tableName string) (*model.ImportLog, error) {
	if len(x.Records) == 0 {
		return nil, nil
	}

	var latest *model.ImportLog
	for _, rec := range x.Records {
		if rec.TableName == tableName {
			if latest == nil || latest.Timestamp.Before(rec.Timestamp) {
				latest = rec
			}
		}
	}

	return latest, nil
}
