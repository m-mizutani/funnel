package otx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/m-mizutani/drone/pkg/domain/model"
	"github.com/m-mizutani/drone/pkg/domain/types"
	"github.com/m-mizutani/drone/pkg/infra"
	"github.com/m-mizutani/drone/pkg/utils"
	"github.com/m-mizutani/goerr"
)

type Subscribed struct {
	apiKey  string
	baseURL *url.URL
}

func NewSubscribed(apiKey string) *Subscribed {
	return &Subscribed{
		apiKey:  apiKey,
		baseURL: utils.Must1(url.Parse("https://otx.alienvault.com")),
	}
}

const (
	initialPeriod = 24 * time.Hour * 30
)

func (x *Subscribed) Import(ctx context.Context, clients *infra.Clients) error {
	const (
		pulseTable = "otx_pulses"
	)

	if err := clients.BigQuery().Migrate(ctx, pulseTable, &PulseLog{}); err != nil {
		return goerr.Wrap(err, "Fail to migrate pulse table")
	}

	var since time.Time
	if log, err := clients.Database().GetLatestImportLog(ctx, types.FeedOTXSubscribed); err != nil {
		return goerr.Wrap(err, "Fail to get latest time of pulse table")
	} else if log != nil {
		since = log.Timestamp
	} else {
		since = time.Now().Add(-initialPeriod)
	}

	utils.Logger().Info("Start to import Subscribed", "since", since)

	sinceText := since.Format("2006-01-02T15:04:05.999+00:00")
	target := *x.baseURL
	target.Path = "/api/v1/pulses/subscribed"
	queryParam := url.Values{}
	queryParam.Add("limit", "50")
	queryParam.Add("modified_since", sinceText)
	target.RawQuery = queryParam.Encode()

	var latest *time.Time

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
		if err != nil {
			return goerr.Wrap(err, "Fail to create request")
		}

		req.Header.Set("X-OTX-API-KEY", x.apiKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return goerr.Wrap(err, "Fail to get Subscribed")
		}
		defer resp.Body.Close()

		var apiResp SubscribedResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			return goerr.Wrap(err, "Fail to decode response body")
		}

		var pulseLogs []PulseLog
		for _, pulse := range apiResp.Results {
			created, err := time.Parse("2006-01-02T15:04:05.999999", pulse.Created)
			if err != nil {
				return goerr.Wrap(err, "Fail to parse created time").With("time", pulse.Created)
			}

			// 2023-12-30T15:02:44.778000
			modified, err := time.Parse("2006-01-02T15:04:05.999999", pulse.Modified)
			if err != nil {
				return goerr.Wrap(err, "Fail to parse modified time").With("time", pulse.Modified)
			}
			if latest == nil || latest.Before(modified) {
				latest = &modified
			}

			pulseLogs = append(pulseLogs, PulseLog{
				Pulse:    pulse,
				Created:  created,
				Modified: modified,
			})
		}
		utils.Logger().Info("Subscribed",
			"count", apiResp.Count,
			"next", apiResp.Next,
			"len(results)", len(apiResp.Results),
			"len(pulseLogs)", len(pulseLogs),
		)

		if err := clients.BigQuery().Insert(ctx, pulseTable, pulseLogs); err != nil {
			return goerr.Wrap(err, "Fail to insert pulse logs")
		}

		nextURL, err := url.Parse(apiResp.Next)
		if err != nil {
			return goerr.Wrap(err, "Fail to parse next URL")
		}
		target = *nextURL

		if apiResp.Next == "" {
			break
		}
	}

	if latest != nil {
		log := model.ImportLog{
			TableName:  pulseTable,
			ImportedAt: time.Now(),
			Timestamp:  *latest,
		}
		if err := clients.Database().PutImportLog(ctx, types.FeedOTXSubscribed, &log); err != nil {
			return goerr.Wrap(err, "Fail to put latest time")
		}
	}

	return nil
}

type SubscribedResponse struct {
	Count            int64       `json:"count" bigquery:"count"`
	Next             string      `json:"next" bigquery:"next"`
	PrefetchPulseIds bool        `json:"prefetch_pulse_ids" bigquery:"prefetch_pulse_ids"`
	Previous         interface{} `json:"previous" bigquery:"previous"`
	Results          []Pulse     `json:"results" bigquery:"results"`
}

type Pulse struct {
	Adversary   string   `json:"adversary" bigquery:"adversary"`
	AttackIds   []string `json:"attack_ids" bigquery:"attack_ids"`
	AuthorName  string   `json:"author_name" bigquery:"author_name"`
	Created     string   `json:"created" bigquery:"-"`
	Description string   `json:"description" bigquery:"description"`
	// ExtractSource     []interface{} `json:"extract_source" bigquery:"extract_source"`
	ID                string      `json:"id" bigquery:"id"`
	Indicators        []Indicator `json:"indicators" bigquery:"indicators"`
	Industries        []string    `json:"industries" bigquery:"industries"`
	MalwareFamilies   []string    `json:"malware_families" bigquery:"malware_families"`
	Modified          string      `json:"modified" bigquery:"-"`
	MoreIndicators    bool        `json:"more_indicators" bigquery:"more_indicators"`
	Name              string      `json:"name" bigquery:"name"`
	Public            int64       `json:"public" bigquery:"public"`
	References        []string    `json:"references" bigquery:"references"`
	Revision          int64       `json:"revision" bigquery:"revision"`
	Tags              []string    `json:"tags" bigquery:"tags"`
	TargetedCountries []string    `json:"targeted_countries" bigquery:"targeted_countries"`
	Tlp               string      `json:"tlp" bigquery:"tlp"`
}

type PulseLog struct {
	Pulse
	Created  time.Time `bigquery:"created"`
	Modified time.Time `bigquery:"modified"`
}

type Indicator struct {
	Content     string `json:"content" bigquery:"content"`
	Created     string `json:"created" bigquery:"created"`
	Description string `json:"description" bigquery:"description"`
	Expiration  string `json:"expiration" bigquery:"expiration"`
	ID          int64  `json:"id" bigquery:"id"`
	Indicator   string `json:"indicator" bigquery:"indicator"`
	IsActive    int64  `json:"is_active" bigquery:"is_active"`
	Role        string `json:"role" bigquery:"role"`
	Title       string `json:"title" bigquery:"title"`
	Type        string `json:"type" bigquery:"type"`
}
