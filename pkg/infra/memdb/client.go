package memdb

import (
	"context"
	"sync"

	"github.com/m-mizutani/drone/pkg/domain/model"
	"github.com/m-mizutani/drone/pkg/domain/types"
)

type MemDB struct {
	latestLogs map[types.FeedID]*model.ImportLog
	rwLock     sync.RWMutex
}

func New() *MemDB {
	return &MemDB{
		latestLogs: map[types.FeedID]*model.ImportLog{},
	}
}

func (x *MemDB) GetLatestImportLog(ctx context.Context, id types.FeedID) (*model.ImportLog, error) {
	x.rwLock.RLock()
	defer x.rwLock.RUnlock()

	return x.latestLogs[id], nil
}

func (x *MemDB) PutImportLog(ctx context.Context, id types.FeedID, log *model.ImportLog) error {
	x.rwLock.Lock()
	defer x.rwLock.Unlock()

	if old, ok := x.latestLogs[id]; ok && old.ImportedAt.After(log.ImportedAt) {
		return nil
	}

	x.latestLogs[id] = log
	return nil
}
