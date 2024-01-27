package config

import (
	"context"

	"github.com/m-mizutani/drone/pkg/infra/firestore"
	"github.com/urfave/cli/v2"
)

type Firestore struct {
	projectID  string
	databaseID string
}

func (x *Firestore) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "firestore-project-id",
			Usage:       "Firestore project ID",
			Destination: &x.projectID,
			EnvVars:     []string{"DRONE_FIRESTORE_PROJECT_ID"},
			Required:    true,
		},
		&cli.StringFlag{
			Name:        "firestore-database-id",
			Usage:       "Firestore database ID",
			Destination: &x.databaseID,
			EnvVars:     []string{"DRONE_FIRESTORE_DATABASE_ID"},
			Required:    true,
		},
	}
}

func (x *Firestore) Configure(ctx context.Context) (*firestore.Client, error) {
	return firestore.New(ctx, x.projectID, x.databaseID)
}
