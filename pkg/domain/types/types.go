package types

type FeedID string

func (x FeedID) String() string { return string(x) }

const (
	FeedOTXSubscribed FeedID = "otx-subscribed"
	FeedAbuseChFeodo  FeedID = "abuse.ch-feodo"
)
