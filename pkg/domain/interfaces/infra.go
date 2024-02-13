package interfaces

import (
	"context"

	"github.com/m-mizutani/drone/pkg/domain/model"
	"github.com/m-mizutani/drone/pkg/domain/types"
)

type BigQuery interface {
	CreateOrUpdateSchema(ctx context.Context, tableName string, target any) error
	Insert(ctx context.Context, tableName string, data any) error
}

type Database interface {
	PutImportLog(ctx context.Context, id types.FeedID, log *model.ImportLog) error
	GetLatestImportLog(ctx context.Context, id types.FeedID) (*model.ImportLog, error)
}
