import { Component, AfterViewInit, ElementRef, ViewChild } from '@angular/core'
import ApexCharts, { ApexOptions } from 'apexcharts'
import { LucideAngularModule } from "lucide-angular"
import { CampaignAnalytics } from 'src/app/models/campaign-analytics.model'
import { ApiService } from 'src/app/services/api.service'

@Component({
  selector: 'app-campaign-events-chart',
  standalone: true,
  imports: [LucideAngularModule],
  templateUrl: './campaign-events-chart.html',
  styleUrl: './campaign-events-chart.css'
})
export class CampaignEventsChart implements AfterViewInit {

  @ViewChild('chart') chartElement!: ElementRef

  constructor(private api: ApiService) { }

  ngAfterViewInit() {

    this.api.getCampaignAnalytics('year').subscribe({

      next: (analytics: CampaignAnalytics) => {

        const timeline = analytics.timeline

        const grouped: Record<string, number> = {}

        timeline.forEach(t => {

          if (!grouped[t.campaign_name]) {
            grouped[t.campaign_name] = 0
          }

          grouped[t.campaign_name] += t.count

        })

        const categories = Object.keys(grouped)
        const seriesData = Object.values(grouped)

        const options: ApexOptions = {

          chart: {
            type: 'bar',
            height: 320,
            toolbar: { show: false },
            foreColor: 'var(--color-base-content)'
          },

          series: [
            {
              name: 'Events',
              data: seriesData
            }
          ],

          colors: ['var(--color-primary)'],

          xaxis: {
            categories: categories,
            labels: {
              style: {
                colors: 'var(--color-base-content)'
              }
            }
          },

          yaxis: {
            labels: {
              style: {
                colors: 'var(--color-base-content)'
              }
            }
          },

          grid: {
            borderColor: 'var(--color-base-300)'
          },

          plotOptions: {
            bar: {
              borderRadius: 8,
              columnWidth: '50%'
            }
          },

          dataLabels: {
            enabled: false
          },

          tooltip: {
            theme: 'dark'
          }

        }

        const chart = new ApexCharts(
          this.chartElement.nativeElement,
          options
        )

        chart.render()

      }

    })

  }

}