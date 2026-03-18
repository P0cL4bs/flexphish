package cli

import "flag"

type CLIOptions struct {
	ConfigFile    string
	DBPath        string
	Host          string
	APIPort       int
	DashboardPort int
	CampaignPort  int
	DevMode       bool
	CreateUser    bool
	DeleteUser    bool
	RunDashboard  bool

	Email    string
	Password string
	Role     string
}

func ParseFlags() *CLIOptions {

	opts := &CLIOptions{}

	flag.StringVar(&opts.ConfigFile, "config", "configs/app.yaml", "Config file path")
	flag.StringVar(&opts.DBPath, "db", "flexphish.db", "Database path")

	flag.StringVar(&opts.Host, "host", "", "Server host")

	flag.IntVar(&opts.APIPort, "api-port", 8088, "API port")
	flag.IntVar(&opts.DashboardPort, "dashboard-port", 8000, "Dashboard port")
	flag.IntVar(&opts.CampaignPort, "campaign-port", 8001, "Campaign port")
	flag.BoolVar(&opts.CreateUser, "create-user", false, "Create a new user")
	flag.BoolVar(&opts.DeleteUser, "delete-user", false, "Delete a user")

	flag.StringVar(&opts.Email, "email", "", "User email")
	flag.StringVar(&opts.Password, "password", "", "User password")
	flag.StringVar(&opts.Role, "role", "user", "User role (admin/user)")

	flag.BoolVar(&opts.RunDashboard, "dashboard", true, "Start the dashboard server")
	flag.BoolVar(&opts.DevMode, "dev", true, "Enable development mode")

	flag.Parse()

	return opts
}
