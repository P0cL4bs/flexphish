import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, RouterModule } from '@angular/router';
import { CampaignDetail } from 'src/app/models/campaign-detail.model';
import { ApiService } from 'src/app/services/api.service';
import { faAndroid, faApple, faWindows, faLinux, faChrome, faFirefox, faSafari, faEdge } from '@fortawesome/free-brands-svg-icons';
import { faQuestionCircle } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeModule } from '@fortawesome/angular-fontawesome';

import {
  faPlay,
  faStop,
  faArchive,
  faPen,
  faTrash,
  faChartLine
} from '@fortawesome/free-solid-svg-icons';
import { Template, TemplateMetadata } from 'src/app/models/template.model';
import { Config } from 'src/app/models/config.model';
import { ToastService } from 'src/app/services/toast.service';
import { Group } from 'src/app/models/group.model';
import { SMTPProfile } from 'src/app/models/smtp.model';
import { EmailTemplate } from 'src/app/models/email-template.model';


@Component({
  selector: 'app-campaign-detail-view',
  imports: [CommonModule, RouterModule, FormsModule, FontAwesomeModule],
  templateUrl: './campaign-detail-view.html',
  styleUrl: './campaign-detail-view.css'
})
export class CampaignDetailView {
  campaignId!: number;
  campaign!: CampaignDetail;
  loading = true;
  faAndroid = faAndroid;
  faApple = faApple;
  faWindows = faWindows;
  faLinux = faLinux;
  faChrome = faChrome;
  faFirefox = faFirefox;
  faSafari = faSafari;
  faEdge = faEdge;
  faQuestionCircle = faQuestionCircle;

  editCampaignData = {
    name: '',
    template_id: '',
    dev_mode: false,
    group_ids: [] as number[],
    smtp_profile_id: null as number | null,
    email_template_id: null as number | null,
  };

  faPlay = faPlay;
  faStop = faStop;
  faArchive = faArchive;
  faEdit = faPen;
  faDelete = faTrash;
  faStats = faChartLine;

  template?: Template;
  templates: TemplateMetadata[] = [];
  loadingTemplates = false;
  groups: Group[] = [];
  loadingGroups = false;
  smtpProfiles: SMTPProfile[] = [];
  loadingSMTPProfiles = false;
  emailTemplates: EmailTemplate[] = [];
  loadingEmailTemplates = false;
  searchTerm: string = '';

  config!: Config;

  resultToDelete: any = null;
  expandedResultId: number | null = null;

  toggleResult(id: number) {
    this.expandedResultId =
      this.expandedResultId === id ? null : id;
  }
  selectedMetadata: any = null;

  constructor(
    private route: ActivatedRoute,
    private apiService: ApiService,
    private toastr: ToastService
  ) { }

  ngOnInit(): void {
    this.campaignId = Number(this.route.snapshot.paramMap.get('id'));
    this.loadCampaign();
    this.apiService.getConfigs().subscribe({
      next: (data) => {
        this.config = data
      },
      error: (err) => {
        const message = err?.error?.error || "Failed to load configs";
        this.toastr.show(message, "error");
      }
    })
  }
  loadCampaign() {
    this.apiService.getCampaignById(this.campaignId)
      .subscribe({
        next: (data) => {

          this.campaign = data;

          if (this.campaign.template_id) {
            this.loadTemplate(this.campaign.template_id);
          }

          console.log(this.campaign);
          this.loading = false;

        },
        error: (err) => {
          console.error(err);
          this.loading = false;
          const message = err?.error?.error || "Failed to get campaign";
          this.toastr.show(message, "error");
        }
      });
  }

  loadTemplate(templateId: string) {

    this.apiService.getTemplateById(templateId)
      .subscribe({
        next: (data) => {
          this.template = data;
        },
        error: (err) => {
          console.error("Error loading template", err);
          const message = err?.error?.error || "Error loading template";
          this.toastr.show(message, "error");
        }
      });

  }
  loadTemplates() {

    this.loadingTemplates = true;

    this.apiService.getTemplatesList()
      .subscribe({
        next: (data) => {
          this.templates = data;
          this.loadingTemplates = false;
        },
        error: (err) => {
          console.error("Error loading templates", err);
          this.loadingTemplates = false;
          const message = err?.error?.error || "Error loading templates";
          this.toastr.show(message, "error");
        }
      });

  }

  loadGroups() {
    this.loadingGroups = true;

    this.apiService.getGroups().subscribe({
      next: (data) => {
        this.groups = data;
        this.loadingGroups = false;
      },
      error: (err) => {
        console.error("Error loading groups", err);
        this.loadingGroups = false;
        const message = err?.error?.error || "Error loading groups";
        this.toastr.show(message, "error");
      }
    });
  }

  loadSMTPProfiles() {
    this.loadingSMTPProfiles = true;

    this.apiService.getSMTPProfiles().subscribe({
      next: (data) => {
        this.smtpProfiles = data;
        this.loadingSMTPProfiles = false;
      },
      error: (err) => {
        console.error("Error loading smtp profiles", err);
        this.loadingSMTPProfiles = false;
        const message = err?.error?.error || "Error loading smtp profiles";
        this.toastr.show(message, "error");
      }
    });
  }

  loadEmailTemplates() {
    this.loadingEmailTemplates = true;

    this.apiService.getEmailTemplates().subscribe({
      next: (data) => {
        this.emailTemplates = data;
        this.loadingEmailTemplates = false;
      },
      error: (err) => {
        console.error("Error loading email templates", err);
        this.loadingEmailTemplates = false;
        const message = err?.error?.error || "Error loading email templates";
        this.toastr.show(message, "error");
      }
    });
  }
  getConversion(): number {
    if (!this.campaign?.total_clicked) return 0;
    return Math.round(
      (this.campaign.total_submitted / this.campaign.total_clicked) * 100
    );
  }

  getUrl(): string {
    return `http://${this.campaign?.subdomain}.${this.config.campaign.base_domain}?test_mode_token=${this.config.security.test_mode_token}`;
  }
  getEventBadge(type: string): string {
    switch (type) {
      case 'submit':
        return 'badge-success';
      case 'click':
        return 'badge-warning';
      case 'open':
        return 'badge-info';
      case 'visit':
      case 'page_view':
        return 'badge-secondary';
      case 'redirect':
        return 'badge-accent';
      case 'error':
        return 'badge-error';
      default:
        return 'badge-neutral';
    }
  }
  getEventsByResult(resultId: number) {
    if (!this.campaign?.events) return [];

    const events = this.campaign.events.filter(
      ev => ev.result_id === resultId
    );

    return this.filterEvents(events);
  }

  shortenUserAgent(ua?: string): string {
    if (!ua) return '-';
    const s = ua.toLowerCase();

    const patterns = [
      { name: 'Chrome', re: /chrome\/\d+/ },
      { name: 'Firefox', re: /firefox\/\d+/ },
      { name: 'Safari', re: /safari\/\d+/ },
      { name: 'Edge', re: /edg\// },
      { name: 'Android', re: /android/ },
      { name: 'iPhone', re: /iphone/ },
      { name: 'iPad', re: /ipad/ },
      { name: 'Windows', re: /windows nt/ },
      { name: 'MacOS', re: /mac os x/ },
      { name: 'Linux', re: /linux/ },
    ];

    for (const p of patterns) {
      if (p.re.test(s)) return p.name;
    }

    return 'Unknown';
  }

  getUAIcon(ua?: string) {
    const name = this.shortenUserAgent(ua);
    switch (name) {
      case 'Android': return this.faAndroid;
      case 'iPhone':
      case 'iPad':
      case 'MacOS': return this.faApple;
      case 'Windows': return this.faWindows;
      case 'Linux': return this.faLinux;
      case 'Chrome': return this.faChrome;
      case 'Firefox': return this.faFirefox;
      case 'Safari': return this.faSafari;
      case 'Edge': return this.faEdge;
      default: return this.faQuestionCircle;
    }
  }

  openMetadataModal(metadata: any) {
    if (!metadata) return;

    try {
      this.selectedMetadata =
        typeof metadata === 'string'
          ? JSON.parse(metadata)
          : metadata;
    } catch {
      this.selectedMetadata = metadata;
    }

    const modal = document.getElementById('metadata_modal') as HTMLDialogElement;
    modal?.showModal();
  }

  closeMetadataModal() {
    const modal = document.getElementById('metadata_modal') as HTMLDialogElement;
    modal?.close();
  }

  openEditCampaignModal() {

    this.editCampaignData = {
      name: this.campaign.name,
      template_id: this.campaign.template_id,
      dev_mode: this.campaign.dev_mode,
      group_ids: (this.campaign.groups || []).map(group => group.id),
      smtp_profile_id: this.campaign.smtp_profile_id ?? null,
      email_template_id: this.campaign.email_template_id ?? null,
    };

    this.loadTemplates();
    this.loadGroups();
    this.loadSMTPProfiles();
    this.loadEmailTemplates();

    const modal = document.getElementById('edit_campaign_modal') as HTMLDialogElement;
    modal?.showModal();
  }

  saveCampaignEdit() {

    const payload = {
      ...this.editCampaignData,
      smtp_profile_id: this.editCampaignData.smtp_profile_id ?? 0,
      email_template_id: this.editCampaignData.email_template_id ?? 0,
      send_emails: this.editCampaignData.smtp_profile_id != null && this.editCampaignData.email_template_id != null,
    };

    this.apiService.updateCampaign(this.campaign.id, payload)
      .subscribe({
        next: (updated) => {

          this.campaign = updated;

          const modal = document.getElementById('edit_campaign_modal') as HTMLDialogElement;
          modal?.close();
        },
        error: (err) => {
          const message = err?.error?.error || "Falied to update campagin";
          this.toastr.show(message, "error");
        }
      });

  }

  isEditGroupSelected(groupId: number): boolean {
    return this.editCampaignData.group_ids.includes(groupId);
  }

  toggleEditGroupSelection(groupId: number, checked: boolean): void {
    if (checked) {
      if (!this.isEditGroupSelected(groupId)) {
        this.editCampaignData.group_ids = [...this.editCampaignData.group_ids, groupId];
      }
      return;
    }

    this.editCampaignData.group_ids = this.editCampaignData.group_ids.filter(id => id !== groupId);
  }

  removeEditGroupSelection(groupId: number): void {
    this.editCampaignData.group_ids = this.editCampaignData.group_ids.filter(id => id !== groupId);
  }

  getEditSelectedGroups(): Group[] {
    return this.groups.filter(group => this.editCampaignData.group_ids.includes(group.id));
  }

  getEditGroupsDropdownLabel(): string {
    const count = this.editCampaignData.group_ids.length;
    if (count === 0) {
      return 'Select groups';
    }
    if (count === 1) {
      return '1 group selected';
    }
    return `${count} groups selected`;
  }

  startCampaign() {
    if (!this.campaign) return;

    this.apiService.startCampaign(this.campaign.id).subscribe({
      next: (c) => this.campaign = c,
      error: (err) => alert(err.message)
    });
  }

  stopCampaign() {
    if (!this.campaign) return;

    this.apiService.stopCampaign(this.campaign.id).subscribe({
      next: (c) => this.campaign = c,
      error: (err) => alert(err.message)
    });
  }

  archiveCampaign() {
    if (!this.campaign) return;

    this.apiService.archiveCampaign(this.campaign.id).subscribe({
      next: (c) => this.campaign = c,
      error: (err) => alert(err.message)
    });
  }

  deleteCampaign() {

    if (!this.campaign) return;

    this.apiService.deleteCampaign(this.campaign.id)
      .subscribe({
        next: () => {

          const modal = document.getElementById('delete_campaign_modal') as HTMLDialogElement;
          modal?.close();

          window.location.href = '/campaigns';

        },
        error: (err) => {
          alert(err.message);
        }
      });

  }

  openDeleteModal() {

    const modal = document.getElementById('delete_campaign_modal') as HTMLDialogElement;
    modal?.showModal();

  }
  confirmDelete(result: any, event: Event) {
    event.stopPropagation();

    this.resultToDelete = result;

    const modal = document.getElementById('deleteResultModal') as HTMLDialogElement;
    modal.showModal();
  }

  closeDeleteModal() {
    const modal = document.getElementById('deleteResultModal') as HTMLDialogElement;
    modal.close();
  }

  deleteResult() {

    if (!this.resultToDelete) return;

    const resultId = this.resultToDelete.id;
    const campaignId = this.campaign.id;

    this.apiService.deleteResult(campaignId, resultId)
      .subscribe({

        next: () => {

          this.campaign.results = this.campaign.results.filter(
            (r: any) => r.id !== resultId
          );

          this.closeDeleteModal();
        },

        error: (err) => {
          console.error(err);
          alert("Failed to delete result");
        }

      });

  }
  filterEvents(events: any[]) {
    if (!this.searchTerm) return events;

    const term = this.searchTerm.toLowerCase();

    return events.filter(ev => {
      return (
        ev.type?.toLowerCase().includes(term) ||
        ev.step?.toLowerCase().includes(term) ||
        ev.path?.toLowerCase().includes(term) ||
        ev.user_agent?.toLowerCase().includes(term) ||
        ev.ip?.toLowerCase().includes(term) ||
        ev.referrer?.toLowerCase().includes(term) ||
        JSON.stringify(ev.metadata || '').toLowerCase().includes(term)
      );
    });
  }
}
