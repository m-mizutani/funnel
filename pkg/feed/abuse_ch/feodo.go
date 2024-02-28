package abuse_ch

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/m-mizutani/bqs"
	"github.com/m-mizutani/drone/pkg/domain/model"
	"github.com/m-mizutani/drone/pkg/domain/types"
	"github.com/m-mizutani/drone/pkg/infra"
	"github.com/m-mizutani/drone/pkg/utils"
	"github.com/m-mizutani/goerr"
)

type Feodo struct {
}

func NewFeodo() *Feodo {
	return &Feodo{}
}

const (
	feodoURL = "https://feodotracker.abuse.ch/downloads/ipblocklist.json"
)

type FeodoResponse struct {
	AsName     string `json:"as_name"`
	AsNumber   int64  `json:"as_number"`
	Country    string `json:"country"`
	FirstSeen  string `json:"first_seen"`
	Hostname   string `json:"hostname"`
	IPAddress  string `json:"ip_address"`
	LastOnline string `json:"last_online"`
	Malware    string `json:"malware"`
	Port       int64  `json:"port"`
	Status     string `json:"status"`
}

type FeodoRecord struct {
	FeodoResponse
	FirstSeen  time.Time
	LastOnline time.Time
}

func (f *Feodo) Import(ctx context.Context, clients *infra.Clients) error {
	const tableName = "abusech_feodo"

	schema, err := bqs.Infer(&FeodoRecord{})
	if err != nil {
		return goerr.Wrap(err, "Fail to infer schema")
	}

	if err := clients.BigQuery().CreateOrUpdateSchema(ctx, tableName, schema); err != nil {
		return goerr.Wrap(err, "Fail to migrate feodo table")
	}

	req, err := http.NewRequest("GET", feodoURL, nil)
	if err != nil {
		return goerr.Wrap(err, "Fail to create request").With("url", feodoURL)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return goerr.Wrap(err, "Fail to get response").With("url", feodoURL)
	}

	if resp.StatusCode != http.StatusOK {
		return goerr.Wrap(err, "Fail to get response").With("url", feodoURL).With("status", resp.StatusCode)
	}

	var data []FeodoResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return goerr.Wrap(err, "Fail to decode response").With("url", feodoURL)
	}

	log, err := clients.Database().GetLatestImportLog(ctx, types.FeedAbuseChFeodo)
	if err != nil {
		return goerr.Wrap(err, "Fail to get latest import log").With("feed", types.FeedAbuseChFeodo)
	}

	var latest *time.Time
	var newRecords []FeodoRecord
	for _, rec := range data {
		firstSeen, err := time.Parse("2006-01-02 15:04:05", rec.FirstSeen)
		if err != nil {
			return goerr.Wrap(err, "Fail to parse first_seen").With("first_seen", rec.FirstSeen)
		}
		lastOnline, err := time.Parse("2006-01-02", rec.LastOnline)
		if err != nil {
			return goerr.Wrap(err, "Fail to parse last_online").With("last_online", rec.LastOnline)
		}
		if log == nil || log.LatestRecord.Before(firstSeen) {
			newRecords = append(newRecords, FeodoRecord{
				FeodoResponse: rec,
				FirstSeen:     firstSeen,
				LastOnline:    lastOnline,
			})
		}
		if latest == nil || latest.Before(firstSeen) {
			latest = &firstSeen
		}
	}

	utils.Logger().Info("Imported Feodo", "new_records", len(newRecords))

	if len(newRecords) > 0 {
		if err := clients.BigQuery().Insert(ctx, tableName, newRecords); err != nil {
			return goerr.Wrap(err, "Fail to insert data").With("table", tableName)
		}
	}

	if latest != nil {
		if err := clients.Database().PutImportLog(ctx, types.FeedAbuseChFeodo, &model.ImportLog{
			LatestRecord: *latest,
			CheckedAt:    time.Now(),
		}); err != nil {
			return goerr.Wrap(err, "Fail to put import log").With("table", tableName)
		}
	}

	return nil
}
