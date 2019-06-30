package dpkgrest

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

const (
	// PackageStateNotInstalled The package is not installed on your system.
	PackageStateNotInstalled = "not-installed"
	// PackageStateConfigFiles Only the configuration files of the package exist on the system.
	PackageStateConfigFiles = "config-files"
	// PackageStateHalfInstalled The installation of the package has  been  started,  but  not  completed  for  some reason.
	PackageStateHalfInstalled = "half-installed"
	// PackageStateUnpacked The package is unpacked, but not configured.
	PackageStateUnpacked = "unpacked"
	// PackageStateHalfConfigured The package is unpacked and configuration has been started, but not yet completed for some reason.
	PackageStateHalfConfigured = "half-configured"
	// PackageStateTriggersAwaited The package awaits trigger processing by another package.
	PackageStateTriggersAwaited = "triggers-awaited"
	// PackageStateTriggersPending The package has been triggered.
	PackageStateTriggersPending = "triggers-pending"
	// PackageStateInstalled The package is unpacked and configured OK.
	PackageStateInstalled = "installed"

	// PackageSelectionStateInstall The package is selected for installation.
	PackageSelectionStateInstall = "install"
	// PackageSelectionStateHold A package marked to be on hold is not handled by dpkg, unless  forced  to  do  that with option --force-hold.
	PackageSelectionStateHold = "hold"
	// PackageSelectionStateDeinstall The  package  is  selected  for  deinstallation  (i.e. we want to remove all files, except configuration files).
	PackageSelectionStateDeinstall = "deinstall"
	// PackageSelectionStatePurge The package is selected to be purged (i.e. we want to remove everything from system directories, even configuration files).
	PackageSelectionStatePurge = "purge"
)

// Package contains all the information about the debian package
type Package struct {
	Name                  string
	Version               string
	PackageState          string
	PackageSelectionState string
}

// ReadPackages reads the dpkg database and returns all the debian packages. Using includeSearch only packages with that name
// are returned. Using excludeSearch all those packages are excluded (which has more prio then includeSearch).
func ReadPackages(f io.Reader, includeSearch *regexp.Regexp, excludeSearch *regexp.Regexp) ([]Package, error) {
	packages := make([]Package, 0)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		// eat all non package information
		if !strings.HasPrefix(text, "Package: ") {
			continue
		}
		packageName := text[9:]
		if excludeSearch != nil {
			// check if we want to exclude this package
			if excludeSearch.MatchString(packageName) {
				continue
			}
		}
		if includeSearch != nil {
			// check if we want to include this package
			if !includeSearch.MatchString(packageName) {
				continue
			}
		}
		p := Package{
			Name: packageName,
		}
		readPackage(scanner, &p)
		packages = append(packages, p)
	}
	return packages, scanner.Err()
}

func readPackage(scanner *bufio.Scanner, p *Package) {
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			break
		}
		if strings.HasPrefix(text, "Status: ") {
			statusus := strings.Split(text[8:], " ")
			if len(statusus) != 3 {
				continue
			}
			p.PackageSelectionState = statusus[0]
			p.PackageState = statusus[2]
		}
		if strings.HasPrefix(text, "Version: ") {
			p.Version = text[9:]
			continue
		}

	}
}
