package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/m-mizutani/drone/pkg/domain/interfaces"
	"github.com/m-mizutani/drone/pkg/domain/model"
	"github.com/m-mizutani/drone/pkg/domain/types"
	"github.com/m-mizutani/goerr"
)

type Client struct {
	client     *firestore.Client
	projectID  string
	databaseID string
}

const (
	importLogTable = "import_logs"
)

func New(ctx context.Context, projectID, databaseID string) (*Client, error) {
	client, err := firestore.NewClientWithDatabase(ctx, projectID, databaseID)

	if err != nil {
		return nil, goerr.Wrap(err, "failed to create firestore client")
	}

	return &Client{
		client:     client,
		projectID:  projectID,
		databaseID: databaseID,
	}, nil
}

// GetLatestImportLog implements interfaces.Database.
func (x *Client) GetLatestImportLog(ctx context.Context, id types.FeedID) (*model.ImportLog, error) {
	doc, err := x.client.Collection(importLogTable).Doc(id.String()).Get(ctx)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return nil, goerr.Wrap(err, "failed to get import log").With("id", id)
		}

		return nil, nil
	}

	var log model.ImportLog
	if err := doc.DataTo(&log); err != nil {
		return nil, goerr.Wrap(err, "failed to convert import log").With("id", id)
	}

	return &log, nil
}

// PutImportLog implements interfaces.Database.
func (x *Client) PutImportLog(ctx context.Context, id types.FeedID, log *model.ImportLog) error {
	// Insert import log if ImportedAt is latest
	doc := x.client.Collection(importLogTable).Doc(id.String())
	if err := x.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		snapshot, err := tx.Get(doc)
		if err != nil {
			if status.Code(err) != codes.NotFound {
				return goerr.Wrap(err, "failed to get import log").With("id", id)
			}
		}

		if snapshot.Exists() {
			var oldLog model.ImportLog
			if err := snapshot.DataTo(&oldLog); err != nil {
				return goerr.Wrap(err, "failed to convert import log").With("id", id)
			}

			if oldLog.ImportedAt.After(log.ImportedAt) {
				return nil
			}
		}

		if err := tx.Set(doc, log); err != nil {
			return goerr.Wrap(err, "failed to create import log").With("id", id)
		}

		return nil
	}); err != nil {
		return goerr.Wrap(err, "failed to put import log").With("id", id)
	}

	return nil
}

// func hashNamespace(input types.Namespace) string {
// 	hash := sha512.New()
// 	hash.Write([]byte(input))
// 	hashed := hash.Sum(nil)
// 	return hex.EncodeToString(hashed)
// }

func (x *Client) Close() error {
	if err := x.client.Close(); err != nil {
		return goerr.Wrap(err, "failed to close firestore client")
	}
	return nil
}

var _ interfaces.Database = &Client{}
