package interfaces

import (
	"context"

	"github.com/m-mizutani/drone/pkg/domain/model"
)

type BigQuery interface {
	Migrate(ctx context.Context, tableName string, schema any) error
	Insert(ctx context.Context, tableName string, data any) error
	PutImportLog(ctx context.Context, log *model.ImportLog) error
	GetLatestImportLog(ctx context.Context, tableName string) (*model.ImportLog, error)
}
