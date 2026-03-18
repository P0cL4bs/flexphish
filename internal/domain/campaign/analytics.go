package campaign

type EventTypeMetric struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

type TimelineMetric struct {
	CampaignID   int64  `json:"campaign_id"`
	CampaignName string `json:"campaign_name"`
	Period       string `json:"period"`
	Count        int64  `json:"count"`
}

type TopCampaignMetric struct {
	CampaignID     int64   `json:"campaign_id"`
	Name           string  `json:"name"`
	Clicked        int64   `json:"clicked"`
	Submitted      int64   `json:"submitted"`
	ConversionRate float64 `json:"conversion_rate"`
}

type CampaignAnalytics struct {
	TotalCampaigns      int64 `json:"total_campaigns"`
	ActiveCampaigns     int64 `json:"active_campaigns"`
	EventsCaptured      int64 `json:"events_captured"`
	CredentialsCaptured int64 `json:"credentials_captured"`

	EventTypes   []EventTypeMetric   `json:"event_types"`
	Timeline     []TimelineMetric    `json:"timeline"`
	TopCampaigns []TopCampaignMetric `json:"top_campaigns"`
}
