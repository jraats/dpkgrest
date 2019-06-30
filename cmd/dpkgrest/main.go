package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/jraats/dpkgrest"
)

var (
	databaseFile   string
	includePattern string
	excludePattern string
	host           string
	port           int

	includeSearch *regexp.Regexp
	excludeSearch *regexp.Regexp
)

func init() {
	flag.StringVar(&databaseFile, "source", "/var/lib/dpkg/status", "the dpkg status database")
	flag.StringVar(&includePattern, "include", "", "a regex to match only those packages")
	flag.StringVar(&excludePattern, "exclude", "", "a regex to exclude all these packages")
	flag.StringVar(&host, "host", "0.0.0.0", "the host to bind to")
	flag.IntVar(&port, "port", 80, "the port to bind to")
	flag.Parse()
}

func main() {
	var err error
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

	http.HandleFunc("/", callback)
	err = http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot run webserver %v\n", err)
		os.Exit(1)
	}
}

func callback(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	var pattern *regexp.Regexp
	queryValues := r.URL.Query()
	if queryPattern, ok := queryValues["q"]; ok && len(queryPattern) > 0 {
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
