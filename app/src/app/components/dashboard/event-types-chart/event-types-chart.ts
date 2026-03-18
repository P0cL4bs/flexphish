import { Component, AfterViewInit, ElementRef, ViewChild } from '@angular/core'
import ApexCharts, { ApexOptions } from 'apexcharts'
import { CampaignAnalytics } from 'src/app/models/campaign-analytics.model'
import { ApiService } from 'src/app/services/api.service'

@Component({
  selector: 'event-types-chart',
  templateUrl: './event-types-chart.html'
})
export class EventTypesChart implements AfterViewInit {

  @ViewChild('chart') chart!: ElementRef

  constructor(private api: ApiService) { }

  ngAfterViewInit() {

    this.api.getCampaignAnalytics('year').subscribe({

      next: (analytics: CampaignAnalytics) => {

        const events = analytics.event_types

        const series = events.map(e => e.count)
        const labels = events.map(e => this.formatLabel(e.type))

        const options: ApexOptions = {

          chart: {
            type: 'donut',
            height: 320
          },

          series: series,

          labels: labels,

          colors: [
            'var(--color-primary)',
            'var(--color-secondary)',
            'var(--color-accent)',
            'var(--color-info)',
            'var(--color-success)',
            'var(--color-warning)'
          ],

          legend: {
            position: 'bottom',
            labels: {
              colors: 'var(--color-base-content)'
            }
          },

          tooltip: {
            theme: 'dark'
          },

          dataLabels: {
            formatter: (val) => `${Number(val).toFixed(1)}%`
          },

          plotOptions: {
            pie: {
              donut: {
                size: '70%',
                labels: {
                  show: true,

                  name: {
                    show: true,
                    fontSize: '14px',
                    color: 'var(--color-base-content)',
                    offsetY: -10
                  },

                  value: {
                    show: true,
                    fontSize: '22px',
                    fontWeight: 600,
                    color: 'var(--color-base-content)',
                    offsetY: 10,
                    formatter: (val: number) => {
                      return val.toString()
                    }
                  },

                  total: {
                    show: true,
                    label: 'Events',
                    fontSize: '14px',
                    color: 'var(--color-base-content)',
                    formatter: (w) => {

                      const total = w.globals.seriesTotals.reduce(
                        (a: number, b: number) => a + b,
                        0
                      )

                      return total.toString()
                    }
                  }

                }
              }
            }
          }

        }

        new ApexCharts(
          this.chart.nativeElement,
          options
        ).render()

      }

    })

  }

  private formatLabel(type: string): string {

    const map: Record<string, string> = {

      page_view: 'Page View',
      redirect: 'Redirect',
      submit: 'Submit',
      error: 'Error'

    }

    return map[type] || type
  }

}