package runtime

import (
	"net"
	"net/http"

	"github.com/mileusna/useragent"
)

type VisitorInfo struct {
	IP        string
	UserAgent string
	Country   string
	City      string
	Device    string
	OS        string
	Browser   string
}

func ExtractVisitorInfo(r *http.Request) VisitorInfo {
	ua := r.UserAgent()

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	// Parse user agent
	parsed := useragent.Parse(ua)

	device := "desktop"
	if parsed.Mobile {
		device = "mobile"
	}
	if parsed.Tablet {
		device = "tablet"
	}

	return VisitorInfo{
		IP:        ip,
		UserAgent: ua,
		Device:    device,
		OS:        parsed.OS,
		Browser:   parsed.Name,
	}
}
