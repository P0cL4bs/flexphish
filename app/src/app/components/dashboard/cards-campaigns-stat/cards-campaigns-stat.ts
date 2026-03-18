import { Component, OnInit } from '@angular/core'
import { LucideAngularModule } from "lucide-angular"
import { CommonModule } from '@angular/common'
import { CampaignAnalytics } from 'src/app/models/campaign-analytics.model'
import { ApiService } from 'src/app/services/api.service'

@Component({
  selector: 'app-cards-campaigns-stat',
  standalone: true,
  imports: [CommonModule, LucideAngularModule],
  templateUrl: './cards-campaigns-stat.html',
  styleUrl: './cards-campaigns-stat.css'
})
export class CardsCampaignsStat implements OnInit {

  analytics?: CampaignAnalytics
  loading = true

  constructor(private api: ApiService) { }

  ngOnInit(): void {
    this.loadStats()
  }

  loadStats() {

    this.api.getCampaignAnalytics('day').subscribe({
      next: (data) => {
        this.analytics = data
        this.loading = false
      },
      error: () => {
        this.loading = false
      }
    })

  }

}