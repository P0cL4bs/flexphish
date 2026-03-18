package servers

import (
	"fmt"
	"net/http"
	"os"

	"flexphish/internal/config"
	"flexphish/pkg/logger"

	"go.uber.org/zap"
)

func StartCampaignServer(mux http.Handler) {

	host := config.GetString("server.host")
	port := config.GetInt("server.campaign_port")

	if port < 1024 && os.Geteuid() != 0 {
		logger.Log.Fatal("Port requires root privileges",
			zap.Int("port", port),
		)
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	logger.Log.Info(fmt.Sprintf("[+] Campaign server running on http://%s", addr))

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Failed to start campaign server",
				zap.Error(err),
			)
		}
	}()
}
