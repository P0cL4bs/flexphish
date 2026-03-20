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

        const periods = this.getPeriodsForCurrentRange()
        const categories = this.getChartCategories(periods)

        // campanhas únicas
        const campaigns = [...new Set(
          timeline.map(t => t.campaign_name)
        )]

        // mapa rápido campaign+period -> count
        const map = new Map<string, number>()

        timeline.forEach(t => {
          const periodKey = this.normalizePeriodKey(t.period)

          const key = `${t.campaign_name}-${periodKey}`
          map.set(key, (map.get(key) ?? 0) + t.count)
        })

        const series = campaigns.map(name => {

          const data = periods.map(p => {
            const lookupPeriod = this.getLookupPeriodKey(p)
            return map.get(`${name}-${lookupPeriod}`) ?? 0
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
            categories,
            type: this.getXAxisType(),
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
              format: this.getTooltipFormat()
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
            xaxis: {
              categories,
              type: this.getXAxisType()
            }
          })

        }

      }

    })

  }

  changePeriod(period: 'day' | 'week' | 'month' | 'year') {
    this.period = period
    this.loadData()
  }

  private getPeriodsForCurrentRange(): string[] {
    const now = new Date()
    now.setSeconds(0, 0)

    if (this.period === 'day') {
      const end = new Date(now)
      end.setMinutes(0, 0, 0)
      const periods: string[] = []

      for (let i = 23; i >= 0; i--) {
        const d = new Date(end)
        d.setHours(end.getHours() - i)
        periods.push(this.formatHourKey(d))
      }

      return periods
    }

    if (this.period === 'week') {
      return this.getLastNDays(7)
    }

    if (this.period === 'month') {
      const start = new Date(now.getFullYear(), now.getMonth(), 1)
      const periods: string[] = []
      const cursor = new Date(start)

      while (cursor <= now) {
        periods.push(this.formatDateKey(cursor))
        cursor.setDate(cursor.getDate() + 1)
      }

      return periods
    }

    const periods: string[] = []
    const currentYear = now.getFullYear()
    const currentMonth = now.getMonth()
    for (let month = 0; month <= currentMonth; month++) {
      periods.push(this.formatMonthDateKey(new Date(currentYear, month, 1)))
    }
    return periods
  }

  private getLastNDays(days: number): string[] {
    const result: string[] = []
    const end = new Date()
    end.setHours(0, 0, 0, 0)

    for (let i = days - 1; i >= 0; i--) {
      const d = new Date(end)
      d.setDate(end.getDate() - i)
      result.push(this.formatDateKey(d))
    }

    return result
  }

  private normalizeDateKey(period: string): string {
    if (/^\d{4}-\d{2}-\d{2}$/.test(period)) {
      return period
    }

    const date = new Date(period)
    if (Number.isNaN(date.getTime())) {
      return period
    }

    return this.formatDateKey(date)
  }

  private normalizeHourKey(period: string): string {
    if (/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}$/.test(period)) {
      return `${period.slice(0, 13)}:00:00`
    }

    if (/^\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}$/.test(period)) {
      const iso = period.replace(' ', 'T')
      return `${iso.slice(0, 13)}:00:00`
    }

    if (/^\d{4}-\d{2}-\d{2}T\d{2}$/.test(period)) {
      return `${period}:00:00`
    }

    const date = new Date(period)
    if (Number.isNaN(date.getTime())) {
      return period
    }
    date.setMinutes(0, 0, 0)
    return this.formatHourKey(date)
  }

  private normalizeMonthKey(period: string): string {
    if (/^\d{4}-\d{2}$/.test(period)) {
      return period
    }

    if (/^\d{4}-\d{2}-\d{2}$/.test(period)) {
      return period.slice(0, 7)
    }

    const date = new Date(period)
    if (Number.isNaN(date.getTime())) {
      return period
    }

    return this.formatMonthKey(date)
  }

  private normalizePeriodKey(period: string): string {
    if (this.period === 'day') {
      return this.normalizeHourKey(period)
    }

    if (this.period === 'year') {
      return this.normalizeMonthKey(period)
    }

    return this.normalizeDateKey(period)
  }

  private getLookupPeriodKey(category: string): string {
    if (this.period === 'year') {
      return category.slice(0, 7)
    }
    return category
  }

  private getTooltipFormat(): string {
    if (this.period === 'day') {
      return 'HH:mm'
    }
    if (this.period === 'year') {
      return 'MMM yyyy'
    }
    return 'dd MMM yyyy'
  }

  private getChartCategories(periods: string[]): string[] {
    if (this.period === 'day') {
      return periods.map((period) => this.formatDayCategoryLabel(period))
    }

    if (this.period !== 'year') {
      return periods
    }

    return periods.map((period) => this.formatYearCategoryLabel(period))
  }

  private getXAxisType(): 'datetime' | 'category' {
    return this.period === 'year' || this.period === 'day' ? 'category' : 'datetime'
  }

  private formatYearCategoryLabel(period: string): string {
    const monthIndex = Number(period.slice(5, 7)) - 1
    const year = period.slice(0, 4)
    const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
    const month = monthNames[monthIndex] || period.slice(5, 7)
    return `${month}/${year}`
  }

  private formatDayCategoryLabel(period: string): string {
    const match = period.match(/T(\d{2}):/)
    if (match) {
      return `${match[1]}:00`
    }
    return period
  }

  private formatHourKey(date: Date): string {
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    const hour = String(date.getHours()).padStart(2, '0')
    return `${year}-${month}-${day}T${hour}:00:00`
  }

  private formatMonthKey(date: Date): string {
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    return `${year}-${month}`
  }

  private formatMonthDateKey(date: Date): string {
    return `${this.formatMonthKey(date)}-01`
  }

  private formatDateKey(date: Date): string {
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    return `${year}-${month}-${day}`
  }

}
