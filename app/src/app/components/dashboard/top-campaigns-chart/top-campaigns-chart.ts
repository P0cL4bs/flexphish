import { Component, AfterViewInit, ElementRef, ViewChild } from '@angular/core'
import ApexCharts, { ApexOptions } from 'apexcharts'
import { CampaignAnalytics } from 'src/app/models/campaign-analytics.model'
import { ApiService } from 'src/app/services/api.service'

@Component({
  selector: 'top-campaigns-chart',
  templateUrl: './top-campaigns-chart.html'
})
export class TopCampaignsChart implements AfterViewInit {

  @ViewChild('chart') chart!: ElementRef

  constructor(private api: ApiService) { }

  ngAfterViewInit() {

    this.api.getCampaignAnalytics('year').subscribe({

      next: (analytics: CampaignAnalytics) => {

        const campaigns = analytics.top_campaigns

        const conversionRates = campaigns.map(c =>
          Number(c.conversion_rate.toFixed(1))
        )

        const options: ApexOptions = {

          chart: {
            type: 'bar',
            height: 320,
            toolbar: { show: false },
            foreColor: 'var(--color-base-content)'
          },

          plotOptions: {
            bar: {
              horizontal: true,
              borderRadius: 6,
              barHeight: '60%'
            }
          },

          series: [
            {
              name: 'Conversion Rate',
              data: conversionRates
            }
          ],

          xaxis: {
            categories: campaigns.map(c => c.name),
            max: 100
          },

          colors: [
            'var(--color-accent)'
          ],

          dataLabels: {
            enabled: true,
            formatter: (val) => `${val}%`
          },

          grid: {
            borderColor: 'hsl(var(--b3))'
          },

          legend: {
            horizontalAlign: 'left',
            labels: {
              colors: 'var(--color-base-content)'
            }
          },

          tooltip: {
            y: {
              formatter: (val) => `${val}%`
            },
            theme: 'dark'
          }

        }

        new ApexCharts(
          this.chart.nativeElement,
          options
        ).render()

      }

    })

  }

}