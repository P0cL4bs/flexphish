package config

import "flexphish/internal/cli"

func ApplyCLIOverrides(opts *cli.CLIOptions) {

	if opts.Host != "" {
		SetConfig("server.host", opts.Host)
	}

	if opts.APIPort != 8088 {
		SetConfig("server.api_port", opts.APIPort)
	}

	if opts.DashboardPort != 8000 {
		SetConfig("server.dashboard_port", opts.DashboardPort)
	}

	if opts.CampaignPort != 8001 {
		SetConfig("server.campaign_port", opts.CampaignPort)
	}
}
