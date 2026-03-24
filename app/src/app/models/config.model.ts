export interface Config {
    server: {
        host: string
        dashboard_port: number
        api_port: number
        campaign_port: number
    }

    session: {
        cookie_name: string
        cookie_domain: string
        cookie_secure: boolean
        cookie_http_only: boolean
        ttl: string
    }

    campaign: {
        base_domain: string
        url_scheme: 'http' | 'https' | string
        subdomain_mode: boolean
    }

    email_scheduler: {
        enabled: boolean
        poll_interval_seconds: number
        emails_per_minute: number
        batch_size: number
        batch_pause_ms: number
        max_parallel_campaigns: number
    }

    security: {
        test_mode_token: string
    }

    template_dir: string
    template_assets_dir: string
}
