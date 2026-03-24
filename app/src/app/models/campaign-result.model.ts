import { CampaignEvent } from "./campaign-event.model";

export type ResultStatus =
    | 'in_progress'
    | 'completed'
    | 'abandoned';

export interface CampaignResult {
    id: number;
    campaign_id: number;
    campaign_target_id?: number;

    session_id: string;

    email?: string;
    username?: string;
    password?: string;

    ip?: string;
    user_agent?: string;
    country?: string;
    city?: string;

    device?: string;
    os?: string;
    browser?: string;

    status: ResultStatus;

    first_seen: string;
    last_seen: string;

    events?: CampaignEvent[];

    created_at?: string;
    updated_at?: string;
}
