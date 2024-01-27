package infra

import (
	"github.com/m-mizutani/drone/pkg/domain/interfaces"
	"github.com/m-mizutani/drone/pkg/infra/memdb"
)

type Clients struct {
	db interfaces.Database
	bq interfaces.BigQuery
}

func New(options ...Option) *Clients {
	clients := &Clients{
		db: memdb.New(),
	}
	for _, opt := range options {
		opt(clients)
	}

	return clients
}

func (x *Clients) Database() interfaces.Database {
	return x.db
}

func (x *Clients) BigQuery() interfaces.BigQuery {
	return x.bq
}

type Option func(*Clients)

func WithDatabase(db interfaces.Database) Option {
	return func(x *Clients) {
		x.db = db
	}
}

func WithBigQuery(bq interfaces.BigQuery) Option {
	return func(x *Clients) {
		x.bq = bq
	}
}
