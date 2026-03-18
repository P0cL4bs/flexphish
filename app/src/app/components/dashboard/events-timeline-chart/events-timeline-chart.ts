import { Component, AfterViewInit, ElementRef, ViewChild } from '@angular/core'
import ApexCharts, { ApexOptions } from 'apexcharts'
import { CampaignAnalytics } from 'src/app/models/campaign-analytics.model'
import { ApiService } from 'src/app/services/api.service'

@Component({
  selector: 'events-timeline-chart',
  templateUrl: './events-timeline-chart.html'
})
export class EventsTimelineChart implements AfterViewInit {

  @ViewChild('chart') chart!: ElementRef

  chartInstance!: ApexCharts

  period: 'day' | 'week' | 'month' | 'year' = 'week'

  constructor(private api: ApiService) { }

  ngAfterViewInit() {
    this.loadData()
  }

  private loadData() {

    this.api.getCampaignAnalytics(this.period).subscribe({

      next: (analytics: CampaignAnalytics) => {

        const timeline = analytics.timeline

        // períodos únicos ordenados
        const periods = [...new Set(
          timeline.map(t => t.period)
        )].sort()

        // campanhas únicas
        const campaigns = [...new Set(
          timeline.map(t => t.campaign_name)
        )]

        // mapa rápido campaign+period -> count
        const map = new Map<string, number>()

        timeline.forEach(t => {
          map.set(`${t.campaign_name}-${t.period}`, t.count)
        })

        const series = campaigns.map(name => {

          const data = periods.map(p => {
            return map.get(`${name}-${p}`) ?? 0
          })

          return { name, data }

        })

        const options: ApexOptions = {

          chart: {
            type: 'area',
            height: 320,
            stacked: true,
            toolbar: { show: false },
            foreColor: 'var(--color-base-content)'
          },

          series,

          xaxis: {
            categories: periods,
            type: 'datetime',
            labels: {
              style: {
                colors: 'var(--color-base-content)'
              }
            }
          },

          legend: {
            position: 'top',
            horizontalAlign: 'left',
            labels: {
              colors: 'var(--color-base-content)'
            }
          },

          colors: [
            'var(--color-primary)',
            'var(--color-secondary)',
            'var(--color-accent)',
            'var(--color-info)',
            'var(--color-success)'
          ],

          stroke: {
            curve: 'smooth',
            width: 2
          },

          markers: {
            size: 3,
            strokeWidth: 2,
            strokeColors: 'hsl(var(--b1))'
          },

          tooltip: {
            theme: 'dark',
            x: {
              format: 'dd MMM yyyy'
            }
          },

          grid: {
            borderColor: 'var(--color-base-300)'
          },

          fill: {
            type: 'gradient',
            gradient: {
              shadeIntensity: 1,
              opacityFrom: 0.6,
              opacityTo: 0.15
            }
          }

        }

        if (!this.chartInstance) {

          this.chartInstance = new ApexCharts(
            this.chart.nativeElement,
            options
          )

          this.chartInstance.render()

        } else {

          this.chartInstance.updateOptions({
            series,
            xaxis: { categories: periods }
          })

        }

      }

    })

  }

  changePeriod(period: 'day' | 'week' | 'month' | 'year') {
    this.period = period
    this.loadData()
  }

}