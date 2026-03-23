import { CampaignEvent } from "./campaign-event.model";
import { CampaignResult } from "./campaign-result.model";
import { CampaignTarget } from "./campaign-target.model";
import { EmailTemplate } from "./email-template.model";
import { Group } from "./group.model";
import { SMTPProfile } from "./smtp.model";

export type CampaignStatus =
    | 'draft'
    | 'scheduled'
    | 'active'
    | 'stopped'
    | 'completed'
    | 'cancelled';

export type EmailDispatchStatus =
    | 'idle'
    | 'queued'
    | 'processing'
    | 'completed'
    | 'failed';

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
    send_emails: boolean;
    email_dispatch_status?: EmailDispatchStatus | string;
    email_dispatch_queued_at?: string;
    email_dispatch_started_at?: string;
    email_dispatch_completed_at?: string;
    email_dispatch_last_attempt_at?: string;
    email_dispatch_last_error?: string;
    email_dispatch_total_targets?: number;
    email_dispatch_sent?: number;
    email_dispatch_failed?: number;
    email_dispatch_pending?: number;
    smtp_profile_id?: number;
    email_template_id?: number;
    smtp_profile?: SMTPProfile;
    email_template?: EmailTemplate;
    groups?: Group[];
    campaign_targets?: CampaignTarget[];

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
    group_ids?: number[];
    smtp_profile_id?: number;
    email_template_id?: number;
    send_emails?: boolean;
    scheduled_start_at?: string;
    scheduled_timezone?: string;
}

export interface UpdateCampaignRequest {
    name?: string;
    template_id?: string;
    dev_mode?: boolean;
    group_ids?: number[];
    smtp_profile_id?: number;
    email_template_id?: number;
    send_emails?: boolean;
    status?: CampaignStatus;
    scheduled_start_at?: string;
    scheduled_timezone?: string;
}
