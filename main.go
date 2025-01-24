package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var version = "development"

// this struct used to parse config.yaml
type Config struct {
	Calendars []CalendarConfig `yaml:"calendars"`
}

// This function loads the configuration file and does some basic validation
// Returns false if the config is not valid or an error occurs
func (config *Config) LoadConfig(file string) bool {
	data, err := os.ReadFile(file)
	if err != nil {
		slog.Error("Unable to open config file! You can use -config to specify a different file", "file", file)
		return false
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		slog.Error("Error while unmarshalling yaml! Check config file is valid", "file", file)
		return false
	}

	// ensure calendars exist
	if len(config.Calendars) == 0 {
		slog.Error("No calendars found! Configuration should define at least one calendar")
		return false
	}

	// validate calendar configs and load secrets
	for i := range config.Calendars {

		// grab pointer so we can mutate values when loading from file
		calendarConfig := &config.Calendars[i]

		// check if url should be loaded from file
		if calendarConfig.FeedURLFile != "" {
			calendarConfig.FeedURL, err = readSecretFile(calendarConfig.FeedURLFile)
			if err != nil {
				slog.Error("Unable to read feed_url_file", "calendar", calendarConfig.Name, "feed_url_file", calendarConfig.FeedURLFile)
				return false
			}
		}

		// check if url seems valid
		if !strings.HasPrefix(calendarConfig.FeedURL, "http://") && !strings.HasPrefix(calendarConfig.FeedURL, "https://") {
			slog.Debug("Calendar URL must begin with http:// or https://", "calendar", calendarConfig.Name, "feed_url", len(calendarConfig.Filters))
			return false
		}

		// check if token should be loaded from file
		if calendarConfig.TokenFile != "" {
			calendarConfig.Token, err = readSecretFile(calendarConfig.TokenFile)
			if err != nil {
				slog.Error("Unable to read token_file", "calendar", calendarConfig.Name, "token_file", calendarConfig.TokenFile)
			}
		}

		// Check to see if auth is disabled (token not set)
		// If so print a warning message and make sure public is enabled in config
		if calendarConfig.Token == "" {
			if !calendarConfig.Public {
				slog.Error("Calendar cannot have authentication disabled without public option enabled in the configuration", "calendar", calendarConfig.Name)
				return false
			}
			slog.Warn("Calendar has no token set. Authentication will be disabled", "calendar", calendarConfig.Name)
		}

		// Print a warning if the calendar has no filters
		if len(calendarConfig.Filters) == 0 {
			slog.Warn("Calendar has no filters and will be proxy-only", "calendar", calendarConfig.Name)
			break
		}

	}

	return true // config is parsed successfully

}

func readSecretFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func main() {

	// command-line args
	var (
		configFile     string
		debugLogging   bool
		jsonLogging    bool
		listenPort     int
		validateConfig bool
		printVersion   bool
	)
	flag.StringVar(&configFile, "config", "config.yaml", "config file")
	flag.BoolVar(&debugLogging, "debug", false, "enable debug logging")
	flag.BoolVar(&printVersion, "version", false, "print version and exit")
	flag.BoolVar(&jsonLogging, "json", false, "output logging in JSON format")
	flag.IntVar(&listenPort, "port", 8080, "listening port for api")
	flag.BoolVar(&validateConfig, "validate", false, "validate config and exit")
	flag.Parse()

	// print version and exit
	if printVersion {
		fmt.Println("version:", version)
		os.Exit(0)
	}

	// setup logging options
	loggingLevel := slog.LevelInfo // default loglevel
	if debugLogging {
		loggingLevel = slog.LevelDebug // debug logging enabled
	}
	opts := &slog.HandlerOptions{
		Level: loggingLevel,
	}

	// create json or text logger based on args
	var logger *slog.Logger
	if jsonLogging {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	slog.SetDefault(logger)

	// load configuration
	slog.Debug("reading config", "configFile", configFile)
	var config Config
	if !config.LoadConfig(configFile) {
		os.Exit(1) // fail if config is not valid
	}
	slog.Debug("loaded config")

	// print a message and exit if validate arg was specified
	if validateConfig {
		slog.Info("configuration was validated successfully")
		os.Exit(0)
	}

	// iterate through calendars in the config and setup a handler for each
	// todo: consider refactor to route requests dynamically?
	for _, calendarConfig := range config.Calendars {

		// configure HTTP endpoint
		httpPath := "/calendars/" + calendarConfig.Name + "/feed"
		slog.Debug("Configuring endpoint", "calendar", calendarConfig.Name, "http_path", httpPath)
		http.HandleFunc(httpPath, func(w http.ResponseWriter, r *http.Request) {

			slog.Debug("Received request for calendar", "http_path", httpPath, "calendar", calendarConfig.Name, "client_ip", r.RemoteAddr)

			// validate token
			token := r.URL.Query().Get("token")
			if token != calendarConfig.Token {
				slog.Warn("Unauthorized access attempt", "client_ip", r.RemoteAddr)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// fetch and filter upstream calendar
			feed, err := calendarConfig.fetch()
			if err != nil {
				slog.Error("Error fetching and filtering feed", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// return calendar
			w.Header().Set("Content-Type", "text/calendar")
			_, err = w.Write(feed)
			if err != nil {
				slog.Error("Error writing response", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			slog.Info("Calendar request processed", "http_path", httpPath, "calendar", calendarConfig.Name, "client_ip", r.RemoteAddr)
		})

	}

	// add a readiness and liveness check endpoint (return blank 200 OK response)
	http.HandleFunc("/liveness", func(w http.ResponseWriter, r *http.Request) {})
	http.HandleFunc("/readiness", func(w http.ResponseWriter, r *http.Request) {})

	// start the webserver
	slog.Info("Starting web server", "port", listenPort)
	if err := http.ListenAndServe(":"+strconv.Itoa(listenPort), nil); err != nil {
		slog.Error("Error starting web server", "error", err)
	}

}
