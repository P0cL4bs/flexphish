import { Campaign } from './campaign.model';
import { CampaignResult } from './campaign-result.model';
import { CampaignEvent } from './campaign-event.model';

export interface CampaignDetail extends Campaign {
    results: CampaignResult[];
    events: CampaignEvent[];
}