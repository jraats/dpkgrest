package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"gopkg.in/yaml.v2"

	"github.com/jraats/dpkgrest"
)

// Config contains all the fields that can be set via a config file
type Config struct {
	Source  string `yaml:"source"`
	Include string `yaml:"include"`
	Exclude string `yaml:"exclude"`
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	Users   []User `yaml:"users"`
}

// User contains basic auth credentials
type User struct {
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
}

var (
	databaseFile   string
	includePattern string
	excludePattern string
	host           string
	port           int
	configFile     string
	username       string
	password       string
	users          []User

	includeSearch *regexp.Regexp
	excludeSearch *regexp.Regexp
)

func init() {
	flag.StringVar(&databaseFile, "source", "/var/lib/dpkg/status", "the dpkg status database")
	flag.StringVar(&includePattern, "include", "", "a regex to match only those packages")
	flag.StringVar(&excludePattern, "exclude", "", "a regex to exclude all these packages")
	flag.StringVar(&host, "host", "0.0.0.0", "the host to bind to")
	flag.IntVar(&port, "port", 80, "the port to bind to")
	flag.StringVar(&configFile, "config", "", "location to the configuration file")
	flag.StringVar(&username, "username", "", "the username for basic auth")
	flag.StringVar(&password, "password", "", "the password for basic auth")
	flag.Parse()
}

func main() {
	var err error
	addDefaultUser()
	if configFile != "" {
		if err = loadConfig(configFile); err != nil {
			fmt.Fprintf(os.Stderr, "invalid config file %v\n", err)
			os.Exit(1)
		}
	}

	if includePattern != "" {
		includeSearch, err = regexp.Compile(includePattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid include regex %v\n", err)
			os.Exit(1)
		}
	}

	if excludePattern != "" {
		excludeSearch, err = regexp.Compile(excludePattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid exclude regex %v\n", err)
			os.Exit(1)
		}
	}

	http.HandleFunc("/list", callback)
	err = http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot run webserver %v\n", err)
		os.Exit(1)
	}
}

func addDefaultUser() {
	users = make([]User, 0)
	if username != "" && password != "" {
		users = append(users, User{Name: username, Password: password})
	}
}

func loadConfig(file string) error {
	var cfg Config
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	decoder := yaml.NewDecoder(f)
	if err = decoder.Decode(&cfg); err != nil {
		return err
	}
	databaseFile = cfg.Source
	includePattern = cfg.Include
	excludePattern = cfg.Exclude
	host = cfg.Host
	port = cfg.Port
	users = append(users, cfg.Users...)
	return nil
}

func callback(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/list" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	// Check basic auth
	if len(users) != 0 {
		username, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Add(`WWW-Authenticate`, `Basic realm="login"`)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		var valid bool
		for _, user := range users {
			if user.Name == username && user.Password == password {
				valid = true
				break
			}
		}
		if !valid {
			w.Header().Add(`WWW-Authenticate`, `Basic realm="login"`)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
	}

	var pattern *regexp.Regexp
	queryValues := r.URL.Query()
	if queryPattern, ok := queryValues["filter"]; ok && len(queryPattern) > 0 {
		var err error
		pattern, err = regexp.Compile(queryPattern[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	packages, err := ListPackages(pattern)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	output, err := json.Marshal(packages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	if _, err := w.Write(output); err != nil {
		fmt.Printf("error writing to client %v\n", err)
	}
}

// ListPackages querys the dpkg status database and filters with the given pattern
func ListPackages(filter *regexp.Regexp) ([]dpkgrest.Package, error) {
	f, err := os.Open(databaseFile)
	if err != nil {
		fmt.Printf("couldn't open dpkg status database %v\n", err)
		return nil, err
	}
	defer f.Close()

	packages, err := dpkgrest.ReadPackages(f, includeSearch, excludeSearch)
	if err != nil {
		return nil, err
	}
	if filter == nil {
		return packages, err
	}

	filterPackages := make([]dpkgrest.Package, 0, len(packages))
	for _, pkg := range packages {
		if filter.MatchString(pkg.Name) {
			filterPackages = append(filterPackages, pkg)
		}
	}
	return filterPackages, err
}
