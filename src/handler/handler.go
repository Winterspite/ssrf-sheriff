package handler

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/winterspite/ssrf-sheriff/src/constants"
	"github.com/winterspite/ssrf-sheriff/src/generators"
	"github.com/winterspite/ssrf-sheriff/src/httpserver"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SerializableResponse is a generic type which both can be safely serialized to both XML and JSON
type SerializableResponse struct {
	SecretToken string `json:"token" xml:"token"`
}

// SSRFSheriffRouter is a wrapper around mux.Router to handle HTTP requests to the sheriff, with logging
type SSRFSheriffRouter struct {
	logger         *zap.Logger
	ssrfToken      string
	webhook        string
	healthcheckURL string
}

// NewHTTPServer provides a new HTTP server listener
func NewHTTPServer(mux *mux.Router) *http.Server {
	return &http.Server{
		Addr:    viper.GetString(constants.EnvAddr),
		Handler: mux,
	}
}

// NewSSRFSheriffRouter returns a new SSRFSheriffRouter which is used to route and handle all HTTP requests
func NewSSRFSheriffRouter(logger *zap.Logger) *SSRFSheriffRouter {
	return &SSRFSheriffRouter{
		logger:         logger,
		ssrfToken:      viper.GetString(constants.EnvSSRFToken),
		webhook:        viper.GetString(constants.EnvWebhookURL),
		healthcheckURL: viper.GetString(constants.EnvHealthcheckURL),
	}
}

// StartFilesGenerator starts the function which is dynamically generating JPG/PNG formats
// with the secret token rendered in the media
func StartFilesGenerator() {
	generators.InitMediaGenerators(viper.GetString(constants.EnvSSRFToken))
}

// StartServer starts the HTTP server
func StartServer(server *http.Server, lc fx.Lifecycle) {
	h := httpserver.NewHandle(server)

	lc.Append(fx.Hook{
		OnStart: h.Start,
		OnStop:  h.Shutdown,
	})
}

// PathHandler is the main handler for all inbound requests
func (s *SSRFSheriffRouter) PathHandler(w http.ResponseWriter, r *http.Request) {
	fileExtension := filepath.Ext(r.URL.Path)
	contentType := mime.TypeByExtension(fileExtension)

	var response string

	switch fileExtension {
	case ".json":
		res, _ := json.Marshal(SerializableResponse{SecretToken: s.ssrfToken})
		response = string(res)
	case ".xml":
		res, _ := xml.Marshal(SerializableResponse{SecretToken: s.ssrfToken})
		response = string(res)
	case ".html":
		tmpl := readTemplateFile("html.html")
		response = fmt.Sprintf(tmpl, s.ssrfToken, s.ssrfToken)
	case ".csv":
		tmpl := readTemplateFile("csv.csv")
		response = fmt.Sprintf(tmpl, s.ssrfToken)
	case ".txt":
		response = fmt.Sprintf("token=%s", s.ssrfToken)
	case ".png":
		response = readTemplateFile("png.png")
	case ".jpg", ".jpeg":
		response = readTemplateFile("jpeg.jpg")
	// TODO: dynamically generate these formats with the secret token rendered in the media
	case ".gif":
		response = readTemplateFile("gif.gif")
	case ".mp3":
		response = readTemplateFile("mp3.mp3")
	case ".mp4":
		response = readTemplateFile("mp4.mp4")
	default:
		response = s.ssrfToken
	}

	if contentType == "" {
		contentType = "text/plain"
	}

	s.logger.Info("New inbound HTTP request",
		zap.String("IP", r.RemoteAddr), zap.String("Path", r.URL.Path),
		zap.String("Response Content-Type", contentType),
		zap.Any("Request Headers", r.Header),
	)

	responseBytes := []byte(response)

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Secret-Token", s.ssrfToken)
	w.WriteHeader(http.StatusOK)

	if s.shouldPostNotification(r) {
		s.PostNotification(r)
	}

	_, _ = w.Write(responseBytes)
}

// shouldPostNotification contains our logic for whether to trigger a webhook notification of SSRF.
func (s *SSRFSheriffRouter) shouldPostNotification(r *http.Request) bool {
	elbHealthCheck := false

	if strings.Contains(r.Header.Get("User-Agent"), "ELB-HealthChecker/") {
		elbHealthCheck = true
	}

	// If we are not the healthcheck URL, it's not an ELB health check request, and we have a webhook configured.
	if !strings.Contains(r.URL.Path, s.healthcheckURL) && s.webhook != "" && !elbHealthCheck {
		return true
	}

	return false
}

func readTemplateFile(templateFileName string) string {
	data, err := os.ReadFile(path.Join("templates", path.Clean(templateFileName)))
	if err != nil {
		return ""
	}

	return string(data)
}

// NewServerRouter returns a new mux.Router for handling any HTTP request to /.*
func NewServerRouter(s *SSRFSheriffRouter) *mux.Router {
	router := mux.NewRouter()

	router.PathPrefix("/").HandlerFunc(s.PathHandler)

	return router
}

// NewLogger returns a new *zap.Logger
func NewLogger() (*zap.Logger, error) {
	zapConfig := zap.NewProductionConfig()
	zapConfig.Encoding = viper.GetString(constants.EnvLoggingFormat)
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zapConfig.DisableStacktrace = true

	if viper.GetString(constants.EnvLogFileName) != "" {
		zapConfig.OutputPaths = []string{"stdout", viper.GetString(constants.EnvLogFileName)}
	} else {
		zapConfig.OutputPaths = []string{"stdout"}
	}

	return zapConfig.Build()
}
