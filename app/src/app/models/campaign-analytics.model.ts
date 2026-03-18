export interface EventTypeMetric {
    type: string
    count: number
}

export interface TimelineMetric {
    campaign_id: number
    campaign_name: string
    period: string
    count: number
}

export interface TopCampaignMetric {
    campaign_id: number
    name: string
    clicked: number
    submitted: number
    conversion_rate: number
}

export interface CampaignAnalytics {
    total_campaigns: number
    active_campaigns: number
    events_captured: number
    credentials_captured: number

    event_types: EventTypeMetric[]
    timeline: TimelineMetric[]
    top_campaigns: TopCampaignMetric[]
}