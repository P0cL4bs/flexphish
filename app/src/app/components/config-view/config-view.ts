import { Component, OnInit } from '@angular/core'
import { CommonModule } from '@angular/common'
import { FormsModule } from '@angular/forms'
import { ApiService } from '../../services/api.service'
import { Config } from '../../models/config.model'


function payloadToConfig(payload: Record<string, any>): Config {

  const config: any = {}

  Object.entries(payload).forEach(([key, value]) => {

    const parts = key.split(".")

    if (parts.length === 1) {

      config[key] = value
      return

    }

    const [section, field] = parts

    if (!config[section]) {
      config[section] = {}
    }

    config[section][field] = value

  })

  return config as Config
}

@Component({
  selector: 'app-config-view',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './config-view.html',
  styleUrl: './config-view.css'
})
export class ConfigView implements OnInit {

  config!: Config
  loading = false
  saving = false

  constructor(private api: ApiService) { }

  ngOnInit(): void {
    this.load()
  }

  load() {

    this.loading = true

    this.api.getConfigs().subscribe({
      next: (data) => {

        this.config = data
        if (!this.config.email_scheduler) {
          this.config.email_scheduler = {
            enabled: true,
            poll_interval_seconds: 15,
            emails_per_minute: 60,
            batch_size: 25,
            batch_pause_ms: 1000,
            max_parallel_campaigns: 3
          }
        }

        this.config.email_scheduler.poll_interval_seconds = this.config.email_scheduler.poll_interval_seconds || 15
        this.config.email_scheduler.emails_per_minute = this.config.email_scheduler.emails_per_minute || 60
        this.config.email_scheduler.batch_size = this.config.email_scheduler.batch_size || 25
        this.config.email_scheduler.batch_pause_ms = this.config.email_scheduler.batch_pause_ms ?? 1000
        this.config.email_scheduler.max_parallel_campaigns = this.config.email_scheduler.max_parallel_campaigns || 3
        this.loading = false

      },
      error: (err) => {

        console.error(err)
        this.loading = false

      }
    })

  }

  save() {

    if (!this.config) return

    this.saving = true

    const payload: Record<string, any> = {}

    Object.entries(this.config).forEach(([section, value]) => {

      if (typeof value === 'object' && value !== null) {

        Object.entries(value).forEach(([key, val]) => {
          payload[`${section}.${key}`] = val
        })

      } else {

        payload[section] = value

      }

    })

    this.api.updateConfigs(payload).subscribe({
      next: () => {

        this.saving = false
        console.log("Configs updated")

      },
      error: (err) => {

        console.error(err)
        this.saving = false

      }
    })

  }

}
