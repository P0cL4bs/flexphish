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
  };
  templates: TemplateMetadata[] = [];
  loadingTemplates = false;

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

    this.api.createCampaign(this.newCampaign)
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
          };

          this.router.navigate(['/campaigns', campaign.id]);
        },
        error: (err) => {
          this.creating = false;
          this.errorMessage = err?.error?.error || "Failed to create a campaign";;
        }
      });
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

}