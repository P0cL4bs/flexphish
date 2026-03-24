import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { CampaignDetail } from 'src/app/models/campaign-detail.model';
import { CampaignStatus } from 'src/app/models/campaign.model';
import { ApiService } from 'src/app/services/api.service';
import { faAndroid, faApple, faWindows, faLinux, faChrome, faFirefox, faSafari, faEdge } from '@fortawesome/free-brands-svg-icons';
import { faQuestionCircle } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeModule } from '@fortawesome/angular-fontawesome';

import {
  faPlay,
  faStop,
  faCircleCheck,
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
import { CampaignTarget } from 'src/app/models/campaign-target.model';
import { CampaignResult } from 'src/app/models/campaign-result.model';
import { CampaignEvent } from 'src/app/models/campaign-event.model';
import { LucideAngularModule } from 'lucide-angular';
import { GroupedSelectGroup, GroupedSingleSelect } from '../shared/grouped-single-select/grouped-single-select';

type CampaignDetailTab = 'overview' | 'delivery' | 'results';

@Component({
  selector: 'app-campaign-detail-view',
  imports: [CommonModule, RouterModule, FormsModule, FontAwesomeModule, LucideAngularModule, GroupedSingleSelect],
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
    schedule_enabled: false,
    schedule_date: '',
    schedule_time: '',
    schedule_timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC',
  };

  faPlay = faPlay;
  faStop = faStop;
  faComplete = faCircleCheck;
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
  eventSearchTerm: string = '';
  dispatchSearchTerm: string = '';
  activeTab: CampaignDetailTab = 'overview';
  tabLoading = false;
  tabLoadingTab: CampaignDetailTab | null = null;

  config!: Config;

  resultToDelete: any = null;
  expandedResultId: number | null = null;
  emailDeliveryPollingId: ReturnType<typeof setInterval> | null = null;
  tabLoadingTimer: ReturnType<typeof setTimeout> | null = null;
  devModeErrorMessage = 'Email sending is not allowed while development mode is enabled.';
  availableTimezones: string[] = [];
  private filteredTargetsCache: CampaignTarget[] = [];
  private filteredTargetsCacheTerm = '';
  private filteredTargetsCacheSource: CampaignTarget[] | null = null;
  private filteredResultsCache: CampaignResult[] = [];
  private filteredResultsCacheTerm = '';
  private filteredResultsCacheSource: CampaignResult[] | null = null;
  filteredCampaignTargets: CampaignTarget[] = [];
  filteredResults: CampaignResult[] = [];
  private eventsByResultCache = new Map<number, CampaignEvent[]>();
  private eventsByResultSource: CampaignEvent[] | null = null;
  private deliveryStatsCacheSourceTargets: CampaignTarget[] | null = null;
  private deliveryStatsCacheSourceGroups: Group[] | null = null;
  private deliveryStatsCache: {
    expected: number;
    sent: number;
    failed: number;
    pending: number;
    opened: number;
    clicked: number;
    submitted: number;
  } = { expected: 0, sent: 0, failed: 0, pending: 0, opened: 0, clicked: 0, submitted: 0 };

  setActiveTab(tab: CampaignDetailTab) {
    if (this.activeTab === tab) {
      return;
    }

    this.activeTab = tab;
    this.startTabLoading(tab);

    if (!this.campaignId) {
      return;
    }

    this.router.navigate(this.getRouteForTab(tab));
  }

  goToResultsTab(target?: CampaignTarget) {
    if (!this.campaignId) return;

    if (this.activeTab !== 'results') {
      this.activeTab = 'results';
      this.startTabLoading('results');
    }

    const queryParams: Record<string, string> = {};
    if (target?.result?.session_id) {
      queryParams['session'] = target.result.session_id;
    }
    if (target?.result?.id) {
      queryParams['expand'] = String(target.result.id);
    }

    this.router.navigate(
      this.getRouteForTab('results'),
      { queryParams }
    );
  }

  toggleResult(id: number) {
    this.expandedResultId =
      this.expandedResultId === id ? null : id;
  }
  selectedMetadata: any = null;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private apiService: ApiService,
    private toastr: ToastService
  ) { }

  ngOnInit(): void {
    this.campaignId = Number(this.route.snapshot.paramMap.get('id'));
    this.syncActiveTabFromRoute();
    this.applyResultsStateFromQuery();
    this.availableTimezones = this.getAvailableTimezones();
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
          this.recomputeFilteredCampaignTargets();
          this.recomputeFilteredResults();
          this.syncEmailDeliveryPolling();

          if (this.campaign.template_id) {
            this.loadTemplate(this.campaign.template_id);
          }

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

  ngOnDestroy(): void {
    this.stopEmailDeliveryPolling();
    if (this.tabLoadingTimer) {
      clearTimeout(this.tabLoadingTimer);
      this.tabLoadingTimer = null;
    }
  }

  isTabLoading(tab: CampaignDetailTab): boolean {
    return this.activeTab === tab && this.tabLoading && this.tabLoadingTab === tab;
  }

  private startTabLoading(tab: CampaignDetailTab): void {
    if (this.tabLoadingTimer) {
      clearTimeout(this.tabLoadingTimer);
      this.tabLoadingTimer = null;
    }

    this.tabLoading = true;
    this.tabLoadingTab = tab;

    this.tabLoadingTimer = setTimeout(() => {
      if (this.activeTab !== tab) {
        return;
      }

      this.tabLoading = false;
      this.tabLoadingTab = null;
      this.tabLoadingTimer = null;
    }, 3000);
  }

  private syncActiveTabFromRoute() {
    const routePath = this.route.snapshot.routeConfig?.path || '';

    if (routePath.includes('results')) {
      this.activeTab = 'results';
      return;
    }

    if (routePath.includes('target-delivery') || routePath.includes('target-develiry')) {
      this.activeTab = 'delivery';
      return;
    }

    this.activeTab = 'overview';
  }

  private getRouteForTab(tab: CampaignDetailTab): (string | number)[] {
    if (tab === 'delivery') {
      return ['/campaigns', this.campaignId, 'target-delivery'];
    }
    if (tab === 'results') {
      return ['/campaigns', this.campaignId, 'results'];
    }
    return ['/campaigns', this.campaignId];
  }

  private applyResultsStateFromQuery() {
    const session = (this.route.snapshot.queryParamMap.get('session') || '').trim();
    const expandRaw = this.route.snapshot.queryParamMap.get('expand');
    const expand = expandRaw ? Number(expandRaw) : null;

    if (session) {
      this.eventSearchTerm = session;
    }
    if (expand && Number.isFinite(expand)) {
      this.expandedResultId = expand;
    }
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

  onEditTemplateSelected(value: string | number | null): void {
    this.editCampaignData.template_id = typeof value === 'string' ? value : '';
  }

  onEditSMTPProfileSelected(value: string | number | null): void {
    this.editCampaignData.smtp_profile_id = this.toNumberValue(value);
  }

  onEditEmailTemplateSelected(value: string | number | null): void {
    this.editCampaignData.email_template_id = this.toNumberValue(value);
  }

  get templateSelectGroups(): GroupedSelectGroup[] {
    const options = this.templates.map(template => {
      const rawCategory = template.category || template.info?.category || '';
      const category = rawCategory.trim() || 'Uncategorized';
      const label = template.name?.trim() || template.filename;

      return {
        group: category,
        label,
        value: template.filename,
        description: label !== template.filename ? template.filename : undefined,
        searchText: (template.tags || template.info?.tags || []).join(' ')
      };
    });

    return this.groupSelectOptions(options);
  }

  get smtpProfileSelectGroups(): GroupedSelectGroup[] {
    const options = this.smtpProfiles.map(profile => ({
      group: profile.is_global ? 'Global profiles' : 'Personal profiles',
      label: profile.name,
      value: profile.id,
      description: `${profile.host}:${profile.port}`,
      searchText: `${profile.host} ${profile.username} ${profile.from_email || ''}`
    }));

    return this.groupSelectOptions(options);
  }

  get emailTemplateSelectGroups(): GroupedSelectGroup[] {
    const options = this.emailTemplates.map(template => ({
      group: (template.category || '').trim() || (template.is_global ? 'Global templates' : 'Uncategorized'),
      label: template.name,
      value: template.id,
      description: template.subject,
      searchText: template.subject
    }));

    return this.groupSelectOptions(options);
  }

  private groupSelectOptions(
    options: Array<{ group: string; label: string; value: string | number; description?: string; searchText?: string }>
  ): GroupedSelectGroup[] {
    const grouped = new Map<string, GroupedSelectGroup>();

    for (const option of options) {
      const groupLabel = option.group.trim() || 'Other';
      if (!grouped.has(groupLabel)) {
        grouped.set(groupLabel, { label: groupLabel, options: [] });
      }

      grouped.get(groupLabel)?.options.push({
        label: option.label,
        value: option.value,
        description: option.description,
        searchText: option.searchText,
      });
    }

    return Array.from(grouped.values())
      .sort((a, b) => a.label.localeCompare(b.label))
      .map(group => ({
        ...group,
        options: [...group.options].sort((a, b) => a.label.localeCompare(b.label))
      }));
  }

  private toNumberValue(value: string | number | null): number | null {
    if (typeof value === 'number' && Number.isFinite(value)) {
      return value;
    }

    if (typeof value === 'string' && value.trim() !== '') {
      const parsed = Number(value);
      return Number.isFinite(parsed) ? parsed : null;
    }

    return null;
  }

  getConversion(): number {
    if (!this.campaign?.total_clicked) return 0;
    return Math.min(100, Math.round(
      (this.campaign.total_submitted / this.campaign.total_clicked) * 100
    ));
  }

  getCampaignTargets(): CampaignTarget[] {
    return this.campaign?.campaign_targets || [];
  }

  getExpectedTargetsCount(): number {
    return this.getDeliveryStats().expected;
  }

  getCampaignTargetsSentCount(): number {
    return this.getDeliveryStats().sent;
  }

  getCampaignTargetsFailedCount(): number {
    return this.getDeliveryStats().failed;
  }

  getCampaignTargetsPendingCount(): number {
    return this.getDeliveryStats().pending;
  }

  getCampaignTargetsOpenedCount(): number {
    return this.getDeliveryStats().opened;
  }

  getCampaignTargetsClickedCount(): number {
    return this.getDeliveryStats().clicked;
  }

  getCampaignTargetsSubmittedCount(): number {
    return this.getDeliveryStats().submitted;
  }

  isEmailDeliveryInProgress(): boolean {
    if (!this.campaign?.send_emails) return false;

    const targets = this.getCampaignTargets();
    const pending = this.getCampaignTargetsPendingCount();
    const sent = this.getCampaignTargetsSentCount();
    const failed = this.getCampaignTargetsFailedCount();

    if (targets.length === 0) return this.campaign.status === 'active';

    return pending > 0 || sent + failed < targets.length;
  }

  private syncEmailDeliveryPolling(): void {
    if (this.isEmailDeliveryInProgress()) {
      this.startEmailDeliveryPolling();
      return;
    }
    this.stopEmailDeliveryPolling();
  }

  private startEmailDeliveryPolling(): void {
    if (this.emailDeliveryPollingId) return;

    this.emailDeliveryPollingId = setInterval(() => {
      this.apiService.getCampaignById(this.campaignId).subscribe({
        next: (data) => {
          this.campaign = data;
          if (!this.isEmailDeliveryInProgress()) {
            this.stopEmailDeliveryPolling();
          }
        },
        error: () => {
          this.stopEmailDeliveryPolling();
        }
      });
    }, 3000);
  }

  private stopEmailDeliveryPolling(): void {
    if (!this.emailDeliveryPollingId) return;

    clearInterval(this.emailDeliveryPollingId);
    this.emailDeliveryPollingId = null;
  }

  getTargetDisplayName(campaignTarget: CampaignTarget): string {
    const firstName = campaignTarget.target?.first_name?.trim() || '';
    const lastName = campaignTarget.target?.last_name?.trim() || '';
    const fullName = `${firstName} ${lastName}`.trim();
    return fullName || '-';
  }

  getTargetStatusBadge(status: string): string {
    switch (status) {
      case 'sent':
        return 'badge-success';
      case 'failed':
        return 'badge-error';
      case 'pending':
      default:
        return 'badge-warning';
    }
  }

  isEmailOpened(target: CampaignTarget): boolean {
    return !!target.opened_at;
  }

  getTargetInteractionLabel(target: CampaignTarget): string {
    if (target.submitted_at) return 'Submitted';
    if (target.clicked_at) return 'Clicked';
    if (target.opened_at) return 'Opened';
    if (target.email_sent_at) return 'Sent';
    return 'Pending';
  }

  getTargetInteractionBadge(target: CampaignTarget): string {
    const label = this.getTargetInteractionLabel(target);
    switch (label) {
      case 'Submitted':
        return 'badge-success';
      case 'Clicked':
        return 'badge-info';
      case 'Opened':
        return 'badge-secondary';
      case 'Sent':
        return 'badge-primary';
      case 'Pending':
      default:
        return 'badge-warning';
    }
  }

  getUrl(): string {
    const scheme = (this.config?.campaign?.url_scheme || 'https').toLowerCase();
    return `${scheme}://${this.campaign?.subdomain}.${this.config.campaign.base_domain}?test_mode_token=${this.config.security.test_mode_token}`;
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
    const events = this.campaign?.events || [];
    if (this.eventsByResultSource !== events) {
      this.eventsByResultSource = events;
      this.eventsByResultCache.clear();
      for (const event of events) {
        const key = event.result_id;
        if (key == null) continue;
        const list = this.eventsByResultCache.get(key) || [];
        list.push(event);
        this.eventsByResultCache.set(key, list);
      }
    }

    return this.eventsByResultCache.get(resultId) || [];
  }

  onDispatchSearchChanged() {
    this.recomputeFilteredCampaignTargets();
  }

  onEventSearchChanged() {
    this.recomputeFilteredResults();
  }

  getFilteredResults() {
    const results = this.campaign?.results || [];
    const term = this.eventSearchTerm.trim().toLowerCase();

    if (
      this.filteredResultsCacheSource === results &&
      this.filteredResultsCacheTerm === term
    ) {
      return this.filteredResultsCache;
    }

    if (!term) {
      this.filteredResultsCacheSource = results;
      this.filteredResultsCacheTerm = term;
      this.filteredResultsCache = results;
      return this.filteredResultsCache;
    }

    this.filteredResultsCache = results.filter(result => {
      return (
        result.email?.toLowerCase().includes(term) ||
        result.username?.toLowerCase().includes(term) ||
        result.session_id?.toLowerCase().includes(term) ||
        result.ip?.toLowerCase().includes(term) ||
        result.user_agent?.toLowerCase().includes(term) ||
        result.status?.toLowerCase().includes(term)
      );
    });
    this.filteredResultsCacheSource = results;
    this.filteredResultsCacheTerm = term;
    return this.filteredResultsCache;
  }

  private getFilteredCampaignTargetsComputed(): CampaignTarget[] {
    const targets = this.getCampaignTargets();
    const term = this.dispatchSearchTerm.trim().toLowerCase();

    if (
      this.filteredTargetsCacheSource === targets &&
      this.filteredTargetsCacheTerm === term
    ) {
      return this.filteredTargetsCache;
    }

    if (!term) {
      this.filteredTargetsCacheSource = targets;
      this.filteredTargetsCacheTerm = term;
      this.filteredTargetsCache = targets;
      return this.filteredTargetsCache;
    }

    this.filteredTargetsCache = targets.filter(target => {
      const fullName = this.getTargetDisplayName(target).toLowerCase();
      const email = (target.target?.email || '').toLowerCase();
      const position = (target.target?.position || '').toLowerCase();
      const status = (target.status || '').toLowerCase();
      const token = (target.token || '').toLowerCase();
      const interaction = this.getTargetInteractionLabel(target).toLowerCase();
      const resultStatus = (target.result?.status || '').toLowerCase();
      const sessionID = (target.result?.session_id || '').toLowerCase();

      return (
        fullName.includes(term) ||
        email.includes(term) ||
        position.includes(term) ||
        status.includes(term) ||
        token.includes(term) ||
        interaction.includes(term) ||
        resultStatus.includes(term) ||
        sessionID.includes(term)
      );
    });
    this.filteredTargetsCacheSource = targets;
    this.filteredTargetsCacheTerm = term;
    return this.filteredTargetsCache;
  }

  trackByTargetId(_: number, target: CampaignTarget): number {
    return target.id;
  }

  trackByResultId(_: number, result: CampaignResult): number {
    return result.id;
  }

  trackByEventId(_: number, event: CampaignEvent): number {
    return event.id;
  }

  private getDeliveryStats() {
    const targets = this.getCampaignTargets();
    const groups = this.campaign?.groups || [];
    if (
      this.deliveryStatsCacheSourceTargets === targets &&
      this.deliveryStatsCacheSourceGroups === groups
    ) {
      return this.deliveryStatsCache;
    }

    let expected = targets.length;
    if (groups.length > 0) {
      const emails = new Set<string>();
      for (const group of groups) {
        for (const target of group.targets || []) {
          const email = (target.email || '').trim().toLowerCase();
          if (email) emails.add(email);
        }
      }
      if (emails.size > 0) {
        expected = emails.size;
      }
    }

    let sent = 0;
    let failed = 0;
    let explicitPending = 0;
    let opened = 0;
    let clicked = 0;
    let submitted = 0;

    for (const target of targets) {
      if (target.status === 'sent') sent++;
      else if (target.status === 'failed') failed++;
      else if (target.status === 'pending') explicitPending++;

      if (target.opened_at) opened++;
      if (target.clicked_at) clicked++;
      if (target.submitted_at) submitted++;
    }

    const resolved = sent + failed + explicitPending;
    const inferredPending = Math.max(expected - resolved, 0);

    this.deliveryStatsCache = {
      expected,
      sent,
      failed,
      pending: explicitPending + inferredPending,
      opened,
      clicked,
      submitted
    };
    this.deliveryStatsCacheSourceTargets = targets;
    this.deliveryStatsCacheSourceGroups = groups;
    return this.deliveryStatsCache;
  }

  private recomputeFilteredCampaignTargets() {
    this.filteredCampaignTargets = this.getFilteredCampaignTargetsComputed();
  }

  private recomputeFilteredResults() {
    this.filteredResults = this.getFilteredResults();
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
    const launchDateParts = this.parseLaunchDateToLocalParts(this.campaign.launch_date);

    this.editCampaignData = {
      name: this.campaign.name,
      template_id: this.campaign.template_id,
      dev_mode: this.campaign.dev_mode,
      group_ids: (this.campaign.groups || []).map(group => group.id),
      smtp_profile_id: this.campaign.smtp_profile_id ?? null,
      email_template_id: this.campaign.email_template_id ?? null,
      schedule_enabled: this.campaign.status === 'scheduled',
      schedule_date: this.campaign.status === 'scheduled' ? launchDateParts.date : '',
      schedule_time: this.campaign.status === 'scheduled' ? launchDateParts.time : '',
      schedule_timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC',
    };

    this.loadTemplates();
    this.loadGroups();
    this.loadSMTPProfiles();
    this.loadEmailTemplates();

    const modal = document.getElementById('edit_campaign_modal') as HTMLDialogElement;
    modal?.showModal();
  }

  openRescheduleCampaignModal() {
    this.openEditCampaignModal();

    this.editCampaignData.schedule_enabled = true;
    if (!this.editCampaignData.schedule_date || !this.editCampaignData.schedule_time) {
      const suggested = new Date(Date.now() + 10 * 60 * 1000);
      const yyyy = suggested.getFullYear();
      const mm = String(suggested.getMonth() + 1).padStart(2, '0');
      const dd = String(suggested.getDate()).padStart(2, '0');
      const hh = String(suggested.getHours()).padStart(2, '0');
      const mi = String(suggested.getMinutes()).padStart(2, '0');
      this.editCampaignData.schedule_date = `${yyyy}-${mm}-${dd}`;
      this.editCampaignData.schedule_time = `${hh}:${mi}`;
    }
  }

  saveCampaignEdit() {
    if (this.campaign.status === 'active' && this.editCampaignData.schedule_enabled) {
      this.toastr.show('Stop the campaign before scheduling a new start time', 'warning');
      return;
    }

    const scheduledStatus: CampaignStatus | undefined = this.editCampaignData.schedule_enabled
      ? 'scheduled'
      : (this.campaign.status === 'scheduled' ? 'draft' : undefined);

    const payload = {
      name: this.editCampaignData.name,
      template_id: this.editCampaignData.template_id,
      dev_mode: this.editCampaignData.dev_mode,
      group_ids: this.editCampaignData.group_ids,
      smtp_profile_id: this.editCampaignData.smtp_profile_id ?? 0,
      email_template_id: this.editCampaignData.email_template_id ?? 0,
      send_emails: this.editCampaignData.smtp_profile_id != null && this.editCampaignData.email_template_id != null,
      status: scheduledStatus,
      scheduled_start_at: this.editCampaignData.schedule_enabled
        ? `${this.editCampaignData.schedule_date}T${this.editCampaignData.schedule_time}`
        : (this.campaign.status === 'scheduled' ? '' : undefined),
      scheduled_timezone: this.editCampaignData.schedule_enabled
        ? this.editCampaignData.schedule_timezone
        : undefined,
    };

    if (this.editCampaignData.schedule_enabled && (!this.editCampaignData.schedule_date || !this.editCampaignData.schedule_time)) {
      this.toastr.show('Scheduled start requires date and time', 'warning');
      return;
    }

    this.apiService.updateCampaign(this.campaign.id, payload)
      .subscribe({
        next: (updated) => {

          this.campaign = updated;

          const modal = document.getElementById('edit_campaign_modal') as HTMLDialogElement;
          modal?.close();

          this.loadCampaign();
        },
        error: (err) => {
          const message = err?.error?.error || "Failed to update campaign";
          this.toastr.show(message, "error");
        }
      });

  }

  private parseLaunchDateToLocalParts(launchDate?: string): { date: string, time: string } {
    if (!launchDate) {
      return { date: '', time: '' };
    }

    const date = new Date(launchDate);
    if (Number.isNaN(date.getTime())) {
      return { date: '', time: '' };
    }

    const yyyy = date.getFullYear();
    const mm = String(date.getMonth() + 1).padStart(2, '0');
    const dd = String(date.getDate()).padStart(2, '0');
    const hh = String(date.getHours()).padStart(2, '0');
    const mi = String(date.getMinutes()).padStart(2, '0');

    return {
      date: `${yyyy}-${mm}-${dd}`,
      time: `${hh}:${mi}`,
    };
  }

  private getAvailableTimezones(): string[] {
    const fallback = [
      'UTC',
      'America/Sao_Paulo',
      'America/Bahia',
      'America/New_York',
      'Europe/London',
    ];

    try {
      const fn = (Intl as any).supportedValuesOf;
      if (typeof fn === 'function') {
        const list = fn.call(Intl, 'timeZone') as string[];
        if (Array.isArray(list) && list.length > 0) {
          return list;
        }
      }
    } catch {
    }

    return fallback;
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
    if (this.campaign.dev_mode && this.campaign.send_emails) {
      this.openDevModeErrorModal();
      return;
    }

    this.apiService.startCampaign(this.campaign.id).subscribe({
      next: (c) => {
        this.campaign = c;
        this.syncEmailDeliveryPolling();
      },
      error: (err) => {
        if (err?.error?.error?.includes('dev_mode')) {
          this.openDevModeErrorModal();
          return;
        }
        this.toastr.show(err.error?.error, "error")
      }
    });
  }

  stopCampaign() {
    if (!this.campaign) return;

    this.apiService.stopCampaign(this.campaign.id).subscribe({
      next: (c) => this.campaign = c,
      error: (err) => alert(err.message)
    });
  }

  completeCampaign() {
    if (!this.campaign) return;

    this.apiService.completeCampaign(this.campaign.id).subscribe({
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

  openDevModeErrorModal() {
    const modal = document.getElementById('dev_mode_error_modal') as HTMLDialogElement;
    modal?.showModal();
  }

  closeDevModeErrorModal() {
    const modal = document.getElementById('dev_mode_error_modal') as HTMLDialogElement;
    modal?.close();
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
          if (this.expandedResultId === resultId) {
            this.expandedResultId = null;
          }
          this.resultToDelete = null;
          this.closeDeleteModal();
          this.loadCampaign();
        },

        error: (err) => {
          console.error(err);
          alert("Failed to delete result");
        }

      });

  }
}
