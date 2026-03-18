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