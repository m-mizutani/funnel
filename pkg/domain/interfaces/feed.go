package interfaces

import "context"

type Feed interface {
	Import(ctx context.Context, bq BigQuery) error
}
