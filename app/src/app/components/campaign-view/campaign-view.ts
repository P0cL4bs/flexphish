import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../services/api.service';
import { Campaign } from '../../models/campaign.model';
import { LucideAngularModule } from 'lucide-angular';
import { Template, TemplateMetadata } from 'src/app/models/template.model';
import { CampaignDetail } from 'src/app/models/campaign-detail.model';
import { Config } from 'src/app/models/config.model';
import { ToastService } from 'src/app/services/toast.service';
import { Group } from 'src/app/models/group.model';
import { SMTPProfile } from 'src/app/models/smtp.model';
import { EmailTemplate } from 'src/app/models/email-template.model';
import { CampaignTarget } from 'src/app/models/campaign-target.model';

@Component({
  selector: 'app-campaign-view',
  standalone: true,
  imports: [CommonModule, RouterModule, FormsModule, LucideAngularModule],
  templateUrl: './campaign-view.html',
  styleUrl: './campaign-view.css'
})
export class CampaignView implements OnInit {

  campaigns: Campaign[] = [];
  filteredCampaigns: Campaign[] = [];

  newCampaign = {
    name: '',
    template_id: '',
    subdomain: '',
    dev_mode: false,
    group_ids: [] as number[],
    smtp_profile_id: null as number | null,
    email_template_id: null as number | null,
  };
  templates: TemplateMetadata[] = [];
  loadingTemplates = false;
  groups: Group[] = [];
  loadingGroups = false;
  smtpProfiles: SMTPProfile[] = [];
  loadingSMTPProfiles = false;
  emailTemplates: EmailTemplate[] = [];
  loadingEmailTemplates = false;

  creating = false;
  errorMessage = '';

  loading = false;

  page = 1;
  pageSize = 10;
  total = 0;

  search = '';
  statusFilter = '';

  expandedRow: number | null = null;
  campaignDetails: { [id: number]: CampaignDetail } = {};
  loadingDetail: number | null = null;
  templateCache: { [id: string]: Template } = {};
  loadingTemplate: { [id: string]: boolean } = {};

  config!: Config

  constructor(private api: ApiService, private router: Router, private toastr: ToastService) { }

  ngOnInit(): void {
    this.loadCampaigns();
    this.loadTemplates();
    this.loadGroups();
    this.loadSMTPProfiles();
    this.loadEmailTemplates();
    this.api.getConfigs().subscribe({
      next: (data) => {
        this.config = data
      },
      error: (err) => {
      }
    })
    window.addEventListener('campagins:reload', () => {
      this.loadCampaigns();
      this.loadTemplates();
      this.loadGroups();
      this.loadSMTPProfiles();
      this.loadEmailTemplates();
    });
  }

  loadGroups() {
    this.loadingGroups = true;

    this.api.getGroups().subscribe({
      next: (data) => {
        this.groups = data;
        this.loadingGroups = false;
      },
      error: () => {
        this.loadingGroups = false;
      }
    });
  }

  loadSMTPProfiles() {
    this.loadingSMTPProfiles = true;

    this.api.getSMTPProfiles().subscribe({
      next: (data) => {
        this.smtpProfiles = data;
        this.loadingSMTPProfiles = false;
      },
      error: () => {
        this.loadingSMTPProfiles = false;
      }
    });
  }

  loadEmailTemplates() {
    this.loadingEmailTemplates = true;

    this.api.getEmailTemplates().subscribe({
      next: (data) => {
        this.emailTemplates = data;
        this.loadingEmailTemplates = false;
      },
      error: () => {
        this.loadingEmailTemplates = false;
      }
    });
  }
  loadTemplates() {
    this.loadingTemplates = true;

    this.api.getTemplatesList().subscribe({
      next: (data) => {
        this.templates = data;
        this.loadingTemplates = false;
      },
      error: () => {
        this.loadingTemplates = false;
      }
    });
  }

  onTemplateSelect(filename: string) {
    this.newCampaign.template_id = filename;
  }
  loadCampaigns(): void {
    this.loading = true;

    this.api.getCampaigns(this.page, this.pageSize).subscribe({
      next: (response) => {
        this.campaigns = response.data;
        this.total = response.total;
        this.applyFilters();
        this.loading = false;
      },
      error: () => {
        this.loading = false;
      }
    });
  }

  openCreateDialog() {
    const dialog = document.getElementById('create_campaign_modal') as HTMLDialogElement;
    dialog?.showModal();
  }

  closeCreateDialog() {
    const dialog = document.getElementById('create_campaign_modal') as HTMLDialogElement;
    dialog?.close();
  }


  createCampaign() {

    if (!this.newCampaign.name ||
      !this.newCampaign.template_id ||
      !this.newCampaign.subdomain) {

      this.errorMessage = 'All fields are required';
      return;
    }

    this.creating = true;
    this.errorMessage = '';

    const payload = {
      ...this.newCampaign,
      smtp_profile_id: this.newCampaign.smtp_profile_id ?? 0,
      email_template_id: this.newCampaign.email_template_id ?? 0,
      send_emails: this.newCampaign.smtp_profile_id != null && this.newCampaign.email_template_id != null,
    };

    this.api.createCampaign(payload)
      .subscribe({
        next: (campaign: Campaign) => {

          this.creating = false;

          this.closeCreateDialog();

          // reset form
          this.newCampaign = {
            name: '',
            template_id: '',
            subdomain: '',
            dev_mode: false,
            group_ids: [],
            smtp_profile_id: null,
            email_template_id: null,
          };

          this.router.navigate(['/campaigns', campaign.id]);
        },
        error: (err) => {
          this.creating = false;
          this.errorMessage = err?.error?.error || "Failed to create a campaign";;
        }
      });
  }

  isCreateGroupSelected(groupId: number): boolean {
    return this.newCampaign.group_ids.includes(groupId);
  }

  toggleCreateGroupSelection(groupId: number, checked: boolean): void {
    if (checked) {
      if (!this.isCreateGroupSelected(groupId)) {
        this.newCampaign.group_ids = [...this.newCampaign.group_ids, groupId];
      }
      return;
    }

    this.newCampaign.group_ids = this.newCampaign.group_ids.filter(id => id !== groupId);
  }

  removeCreateGroupSelection(groupId: number): void {
    this.newCampaign.group_ids = this.newCampaign.group_ids.filter(id => id !== groupId);
  }

  getCreateSelectedGroups(): Group[] {
    return this.groups.filter(group => this.newCampaign.group_ids.includes(group.id));
  }

  getCreateGroupsDropdownLabel(): string {
    const count = this.newCampaign.group_ids.length;
    if (count === 0) {
      return 'Select groups';
    }
    if (count === 1) {
      return '1 group selected';
    }
    return `${count} groups selected`;
  }

  applyFilters(): void {
    this.filteredCampaigns = this.campaigns.filter(c => {
      const matchesSearch =
        c.name.toLowerCase().includes(this.search.toLowerCase());

      const matchesStatus =
        this.statusFilter ? c.status === this.statusFilter : true;

      return matchesSearch && matchesStatus;
    });
    this.campaigns.forEach(c => {

      this.api.getCampaignById(c.id).subscribe(detail => {

        this.campaignDetails[c.id] = detail;

      });

    });
  }

  nextPage(): void {
    if (this.page * this.pageSize < this.total) {
      this.page++;
      this.loadCampaigns();
    }
  }

  prevPage(): void {
    if (this.page > 1) {
      this.page--;
      this.loadCampaigns();
    }
  }

  getStatusColor(status: string): string {
    switch (status) {
      case 'active': return 'bg-green-100 text-green-700';
      case 'draft': return 'bg-yellow-100 text-yellow-700';
      case 'paused': return 'bg-gray-100 text-gray-700';
      case 'completed': return 'bg-blue-100 text-blue-700';
      default: return 'bg-gray-100 text-gray-700';
    }
  }

  getTotalResults(c: Campaign): number {
    const detail = this.campaignDetails[c.id];
    if (!detail?.results?.length) return 0;
    return detail.results.length ?? 0;
  }

  getTotalClicked(c: any): number {
    return c?.total_clicked ?? 0;
  }

  getTotalOpened(c: any): number {
    return c?.total_opened ?? 0;
  }

  getConversionRate(c: any): number {
    const clicked = c?.total_clicked ?? 0;
    const submitted = c?.total_submitted ?? 0;

    if (clicked === 0) return 0;

    return Math.round((submitted / clicked) * 100);
  }

  toggleRow(campaign: Campaign, event: MouseEvent) {
    event.stopPropagation();
    if (this.expandedRow === campaign.id) {

      this.expandedRow = null;
      return;

    }

    this.expandedRow = campaign.id;

    if (!this.campaignDetails[campaign.id]) {

      this.loadingDetail = campaign.id;

      this.api.getCampaignById(campaign.id).subscribe({
        next: (detail) => {

          this.campaignDetails[campaign.id] = detail;

          this.loadingDetail = null;

        },
        error: () => {

          this.loadingDetail = null;

        }
      });

    }

  }

  getLastCapture(campaign: Campaign): Date | null {

    const detail = this.campaignDetails[campaign.id];

    if (!detail?.results?.length) return null;

    const last = detail.results.reduce((a, b) =>
      new Date(a.last_seen) > new Date(b.last_seen) ? a : b
    );

    return new Date(last.last_seen);

  }

  getTemplateById(templateId: string): Template | null {

    const template = this.templateCache[templateId];

    if (template) {
      return template;
    }

    if (!this.loadingTemplate[templateId]) {

      this.loadingTemplate[templateId] = true;

      this.api.getTemplateById(templateId).subscribe({

        next: (template) => {
          this.templateCache[templateId] = template;
          this.loadingTemplate[templateId] = false;
        },

        error: (err) => {
          console.error(err);
          this.loadingTemplate[templateId] = false;
        }

      });

    }

    return null;

  }

  private getCampaignDeliveryTargets(campaign: Campaign): CampaignTarget[] {
    return campaign.campaign_targets || [];
  }

  getEmailDeliveryLabel(campaign: Campaign): string {
    if (!campaign.send_emails) return 'Disabled';

    if (campaign.email_dispatch_status) {
      switch (campaign.email_dispatch_status) {
        case 'queued':
          return 'Queued';
        case 'processing':
          return 'In progress';
        case 'completed':
          return 'Complete';
        case 'failed':
          return 'Failed';
        case 'idle':
        default:
          break;
      }
    }

    const targets = this.getCampaignDeliveryTargets(campaign);
    if (targets.length === 0) return 'Pending';

    const pending = targets.filter(target => target.status === 'pending').length;
    if (pending > 0) return 'In progress';

    return 'Complete';
  }

  getEmailDeliveryBadgeClass(campaign: Campaign): string {
    const label = this.getEmailDeliveryLabel(campaign);
    switch (label) {
      case 'Disabled':
        return 'badge-ghost';
      case 'Pending':
        return 'badge-warning';
      case 'In progress':
        return 'badge-info';
      case 'Complete':
        return 'badge-success';
      case 'Queued':
        return 'badge-warning';
      case 'Failed':
        return 'badge-error';
      default:
        return 'badge-ghost';
    }
  }

}
