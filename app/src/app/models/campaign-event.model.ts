export type EventType =
    | 'visit'
    | 'page_view'
    | 'submit'
    | 'click'
    | 'open'
    | 'redirect'
    | 'error';

export interface CampaignEvent {
    id: number;
    campaign_id: number;
    result_id?: number;

    type: EventType;
    step_id: string;

    path: string;
    ip: string;
    user_agent: string;
    referrer: string;

    metadata?: string;

    created_at: string;
}