import { GroupTarget } from "./group.model";

export type CampaignTargetStatus = 'pending' | 'sent' | 'failed';

export interface CampaignTarget {
    id: number;
    campaign_id: number;
    target_id: number;
    token: string;
    status: CampaignTargetStatus | string;
    email_sent_at?: string;
    opened_at?: string;
    clicked_at?: string;
    submitted_at?: string;
    ip?: string;
    user_agent?: string;
    target?: GroupTarget;
    created_at: string;
    updated_at: string;
}
