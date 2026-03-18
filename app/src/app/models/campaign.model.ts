import { CampaignEvent } from "./campaign-event.model";
import { CampaignResult } from "./campaign-result.model";

export type CampaignStatus =
    | 'draft'
    | 'scheduled'
    | 'active'
    | 'stopped'
    | 'completed'
    | 'cancelled';

export interface Campaign {
    id: number;
    name: string;
    subdomain: string;
    status: CampaignStatus;

    launch_date?: string;
    completed_date?: string;

    created_at: string;
    updated_at: string;

    template_id: string;

    track_opens: boolean;
    track_clicks: boolean;
    track_geo_location: boolean;
    track_device_info: boolean;
    track_ip: boolean;
    track_user_agent: boolean;
    track_referrer: boolean;
    dev_mode: boolean;

    enable_fingerprinting: boolean;

    total_sent: number;
    total_opened: number;
    total_clicked: number;
    total_submitted: number;

    unique_opened: number;
    unique_clicked: number;
    unique_submitted: number;

    is_archived: boolean;
    deleted_at?: string;

    results?: CampaignResult[];
    events?: CampaignEvent[];
}


export interface CreateCampaignRequest {
    name: string;
    template_id: string;
    subdomain: string;
    dev_mode: boolean;
}

export interface UpdateCampaignRequest {
    name: string;
    template_id: string;
    dev_mode: boolean;
}