package infra_test

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/m-mizutani/drone/pkg/domain/interfaces"
	"github.com/m-mizutani/drone/pkg/domain/model"
	"github.com/m-mizutani/drone/pkg/domain/types"
	"github.com/m-mizutani/drone/pkg/infra/firestore"
	"github.com/m-mizutani/drone/pkg/infra/memdb"
	"github.com/m-mizutani/drone/pkg/utils"
	"github.com/m-mizutani/gt"
)

func TestFirestore(t *testing.T) {
	var (
		projectID  string
		databaseID string
	)

	if err := utils.LoadEnv(
		utils.Env("TEST_FIRESTORE_PROJECT_ID", &projectID),
		utils.Env("TEST_FIRESTORE_DATABASE_ID", &databaseID),
	); err != nil {
		t.Skipf("Skip test due to lack of env variables: %v", err)
	}

	ctx := context.Background()
	db := gt.R1(firestore.New(ctx, projectID, databaseID)).NoError(t)
	testDB(t, db)
}

func TestMemDB(t *testing.T) {
	testDB(t, memdb.New())
}

func testDB(t *testing.T, db interfaces.Database) {
	t.Run("basic", func(t *testing.T) {
		testBasic(t, db)
	})

	t.Run("random test", func(t *testing.T) {
		testRandomPut(t, db)
	})
}

func testBasic(t testing.TB, db interfaces.Database) {
	now := time.Now()
	log1 := model.ImportLog{
		TableName:  "test",
		ImportedAt: now,
		Timestamp:  now,
	}
	log2 := model.ImportLog{
		TableName:  "test",
		ImportedAt: now.Add(4 * time.Second),
		Timestamp:  now.Add(4 * time.Hour),
	}
	log3 := model.ImportLog{
		TableName:  "test",
		ImportedAt: now.Add(2 * time.Second),
		Timestamp:  now.Add(2 * time.Hour),
	}
	log4 := model.ImportLog{
		TableName:  "test",
		ImportedAt: now.Add(5 * time.Second),
		Timestamp:  now.Add(5 * time.Hour),
	}

	ctx := context.Background()
	gt.NoError(t, db.PutImportLog(ctx, "test", &log1))
	gt.NoError(t, db.PutImportLog(ctx, "test", &log2))
	gt.NoError(t, db.PutImportLog(ctx, "test", &log3))
	gt.NoError(t, db.PutImportLog(ctx, "test-2", &log4))

	log, err := db.GetLatestImportLog(ctx, "test")
	gt.NoError(t, err)
	gt.Equal(t, log.ImportedAt.Unix(), now.Add(4*time.Second).Unix())
}

func testRandomPut(t *testing.T, db interfaces.Database) {
	const (
		logCount = 100
		parallel = 10
	)
	var (
		feedID = types.FeedID(uuid.NewString())
	)

	now := time.Now()
	var logs []*model.ImportLog
	var maxTS time.Time
	for i := 0; i < logCount; i++ {
		maxTS = now.Add(time.Duration(i) * time.Second)
		logs = append(logs, &model.ImportLog{
			TableName:  "test",
			ImportedAt: maxTS,
			Timestamp:  now,
		})
	}

	// random shuffle logs
	for s := 0; s < 100; s++ {
		p := rand.Int() % logCount
		q := rand.Int() % logCount
		logs[p], logs[q] = logs[q], logs[p]
	}

	logCh := make(chan *model.ImportLog, logCount)
	for i := range logs {
		logCh <- logs[i]
	}
	close(logCh)

	ctx := context.Background()
	var wg sync.WaitGroup
	for n := 0; n < parallel; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for log := range logCh {
				gt.NoError(t, db.PutImportLog(ctx, feedID, log))
			}
		}()
	}

	wg.Wait()
	log := gt.R1(db.GetLatestImportLog(ctx, feedID)).NoError(t)
	gt.Equal(t, log.ImportedAt.Unix(), maxTS.Unix())
}
