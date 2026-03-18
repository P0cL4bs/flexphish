package servers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"flexphish/pkg/logger"
	"flexphish/pkg/utils"

	"go.uber.org/zap"
)

func StartDashboard(address string, port int, path string) error {

	filesDir := utils.GetBasePath(path)

	info, err := os.Stat(filesDir)
	if err != nil || !info.IsDir() {
		logger.Log.Error("Dashboard directory not found. Please build the frontend first",
			zap.String("path", filesDir),
		)
		return err
	}

	router := http.NewServeMux()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		clientIP := strings.Split(r.RemoteAddr, ":")[0]

		logger.Log.Debug("HTTP request",
			zap.String("client", clientIP),
			zap.String("method", r.Method),
			zap.String("host", r.Host),
			zap.String("url", r.URL.Path),
		)

		requestedPath := filepath.Join(filesDir, filepath.Clean(r.URL.Path))

		if info, err := os.Stat(requestedPath); err == nil && !info.IsDir() {
			http.ServeFile(w, r, requestedPath)
			return
		}

		http.ServeFile(w, r, filepath.Join(filesDir, "index.html"))
	})

	server := &http.Server{
		Addr:    address + ":" + strconv.Itoa(port),
		Handler: router,
	}

	logger.Log.Info(
		fmt.Sprintf("[+] dashboard running on %s:%d | ui: %s", address, port, filesDir),
	)

	return server.ListenAndServe()
}
