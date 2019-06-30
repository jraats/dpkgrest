package dpkgrest

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadPackages(t *testing.T) {
	reader, err := os.Open(filepath.Join("testdata", "status"))
	require.Nil(t, err)

	packages, err := ReadPackages(reader, nil, nil)
	require.Nil(t, err)

	expected := []Package{
		{
			Name:                  "python-apt-common",
			Version:               "0.9.3.12",
			PackageState:          "installed",
			PackageSelectionState: "install",
		}, {
			Name:                  "libregexp-common-perl",
			Version:               "2013031301-1",
			PackageState:          "installed",
			PackageSelectionState: "install",
		}, {
			Name:                  "bind9-host",
			Version:               "1:9.9.5.dfsg-9+deb8u15",
			PackageState:          "installed",
			PackageSelectionState: "install",
		}, {
			Name:                  "iputils-ping",
			Version:               "3:20121221-5",
			PackageState:          "installed",
			PackageSelectionState: "install",
		}, {
			Name:                  "libxi6",
			Version:               "2:1.7.4-1+deb8u1",
			PackageState:          "config-files",
			PackageSelectionState: "deinstall",
		}, {
			Name:                  "libedit2",
			Version:               "3.1-20140620-2",
			PackageState:          "installed",
			PackageSelectionState: "install",
		}, {
			Name:                  "libpam-runtime",
			Version:               "1.1.8-3.1+deb8u2+rpi3",
			PackageState:          "installed",
			PackageSelectionState: "install",
		},
	}
	require.Equal(t, expected, packages)
}

func TestReadPackagesWithIncludeFilter(t *testing.T) {
	reader, err := os.Open(filepath.Join("testdata", "status"))
	require.Nil(t, err)

	includeRegex, err := regexp.Compile(`\-common\-?`)
	require.Nil(t, err)

	packages, err := ReadPackages(reader, includeRegex, nil)
	require.Nil(t, err)

	expected := []Package{
		{
			Name:                  "python-apt-common",
			Version:               "0.9.3.12",
			PackageState:          "installed",
			PackageSelectionState: "install",
		}, {
			Name:                  "libregexp-common-perl",
			Version:               "2013031301-1",
			PackageState:          "installed",
			PackageSelectionState: "install",
		},
	}
	require.Equal(t, expected, packages)
}

func TestReadPackagesWithExcludeFilter(t *testing.T) {
	reader, err := os.Open(filepath.Join("testdata", "status"))
	require.Nil(t, err)

	excludeRegex, err := regexp.Compile(`^lib`)
	require.Nil(t, err)

	packages, err := ReadPackages(reader, nil, excludeRegex)
	require.Nil(t, err)

	expected := []Package{
		{
			Name:                  "python-apt-common",
			Version:               "0.9.3.12",
			PackageState:          "installed",
			PackageSelectionState: "install",
		}, {
			Name:                  "bind9-host",
			Version:               "1:9.9.5.dfsg-9+deb8u15",
			PackageState:          "installed",
			PackageSelectionState: "install",
		}, {
			Name:                  "iputils-ping",
			Version:               "3:20121221-5",
			PackageState:          "installed",
			PackageSelectionState: "install",
		},
	}
	require.Equal(t, expected, packages)
}

func TestReadPackagesWithIncludeExcludeFilter(t *testing.T) {
	reader, err := os.Open(filepath.Join("testdata", "status"))
	require.Nil(t, err)

	includeRegex, err := regexp.Compile(`^lib`)
	require.Nil(t, err)

	excludeRegex, err := regexp.Compile(`\-common|\-runtime`)
	require.Nil(t, err)

	packages, err := ReadPackages(reader, includeRegex, excludeRegex)
	require.Nil(t, err)

	expected := []Package{
		{
			Name:                  "libxi6",
			Version:               "2:1.7.4-1+deb8u1",
			PackageState:          "config-files",
			PackageSelectionState: "deinstall",
		}, {
			Name:                  "libedit2",
			Version:               "3.1-20140620-2",
			PackageState:          "installed",
			PackageSelectionState: "install",
		},
	}
	require.Equal(t, expected, packages)
}
