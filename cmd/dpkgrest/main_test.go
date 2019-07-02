package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func defaultInit() {
	databaseFile = ""
	includePattern = ""
	excludePattern = ""
	host = ""
	port = 80
	configFile = ""
	username = ""
	password = ""
	users = make([]User, 0)
}

func TestHttpServer(t *testing.T) {
	defaultInit()
	databaseFile = filepath.Join("..", "..", "testdata", "status")

	ts := httptest.NewServer(http.HandlerFunc(callback))
	defer ts.Close()

	res, err := http.Get(ts.URL + "/list")
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	require.Nil(t, err)

	expected := `[
		{
			"Name":"python-apt-common",
			"PackageSelectionState":"install",
			"PackageState":"installed",
			"Version":"0.9.3.12"
		},{
			"Name":"libregexp-common-perl",
			"PackageSelectionState":"install",
			"PackageState":"installed",
			"Version":"2013031301-1"
		},{
			"Name":"bind9-host",
			"PackageSelectionState":"install",
			"PackageState":"installed",
			"Version":"1:9.9.5.dfsg-9+deb8u15"
		},{
			"Name":"iputils-ping",
			"PackageSelectionState":"install",
			"PackageState":"installed",
			"Version":"3:20121221-5"
		},{
			"Name":"libxi6",
			"PackageSelectionState":"deinstall",
			"PackageState":"config-files",
			"Version":"2:1.7.4-1+deb8u1"
		},{
			"Name":"libedit2",
			"PackageSelectionState":"install",
			"PackageState":"installed",
			"Version":"3.1-20140620-2"
		},{
			"Name":"libpam-runtime",
			"PackageSelectionState":"install",
			"PackageState":"installed",
			"Version":"1.1.8-3.1+deb8u2+rpi3"
		}
	]`
	require.JSONEq(t, expected, string(body))
}

func TestHttpServerWithFilter(t *testing.T) {
	defaultInit()
	databaseFile = filepath.Join("..", "..", "testdata", "status")
	ts := httptest.NewServer(http.HandlerFunc(callback))
	defer ts.Close()

	res, err := http.Get(ts.URL + "/list?filter=lib")
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	require.Nil(t, err)

	expected := `[
		{
			"Name":"libregexp-common-perl",
			"PackageSelectionState":"install",
			"PackageState":"installed",
			"Version":"2013031301-1"
		},{
			"Name":"libxi6",
			"PackageSelectionState":"deinstall",
			"PackageState":"config-files",
			"Version":"2:1.7.4-1+deb8u1"
		},{
			"Name":"libedit2",
			"PackageSelectionState":"install",
			"PackageState":"installed",
			"Version":"3.1-20140620-2"
		},{
			"Name":"libpam-runtime",
			"PackageSelectionState":"install",
			"PackageState":"installed",
			"Version":"1.1.8-3.1+deb8u2+rpi3"
		}
	]`
	require.JSONEq(t, expected, string(body))
}

func TestHttpServerWithInvalidFilter(t *testing.T) {
	defaultInit()
	databaseFile = filepath.Join("..", "..", "testdata", "status")
	ts := httptest.NewServer(http.HandlerFunc(callback))
	defer ts.Close()

	res, err := http.Get(ts.URL + "/list?filter=lib)")
	require.Nil(t, err)
	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	res.Body.Close()
}

func TestHttpServerWithInvalidDatabase(t *testing.T) {
	defaultInit()
	databaseFile = "invalid"
	ts := httptest.NewServer(http.HandlerFunc(callback))
	defer ts.Close()

	res, err := http.Get(ts.URL + "/list")
	require.Nil(t, err)
	require.Equal(t, http.StatusInternalServerError, res.StatusCode)

	res.Body.Close()
}

func TestHttpServerWithInvalidPath(t *testing.T) {
	defaultInit()
	databaseFile = filepath.Join("..", "..", "testdata", "status")
	ts := httptest.NewServer(http.HandlerFunc(callback))
	defer ts.Close()

	res, err := http.Get(ts.URL + "/wrong")
	require.Nil(t, err)
	require.Equal(t, http.StatusNotFound, res.StatusCode)

	res.Body.Close()
}

func TestHttpServerWithBasicAuth(t *testing.T) {
	defaultInit()
	databaseFile = filepath.Join("..", "..", "testdata", "status")
	users = []User{
		User{
			Name:     "me",
			Password: "test",
		},
		User{
			Name:     "another",
			Password: "me",
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(callback))
	defer ts.Close()

	r, err := http.NewRequest(http.MethodGet, ts.URL+"/list?filter=notfound", nil)
	require.Nil(t, err)
	res, err := ts.Client().Do(r)
	require.Nil(t, err)
	require.Equal(t, http.StatusUnauthorized, res.StatusCode)
	res.Body.Close()

	r.SetBasicAuth("invalid", "credentials")
	res, err = ts.Client().Do(r)
	require.Nil(t, err)
	require.Equal(t, http.StatusUnauthorized, res.StatusCode)
	res.Body.Close()

	r.SetBasicAuth("another", "me")
	res, err = ts.Client().Do(r)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	res.Body.Close()
}

func TestConfig(t *testing.T) {
	defaultInit()
	configFile = filepath.Join("testdata", "test.config.yml")
	require.Nil(t, loadConfig(configFile))

	require.Equal(t, "/var/lib/dpkg/status", databaseFile)
	require.Equal(t, "php", includePattern)
	require.Equal(t, "lib", excludePattern)
	require.Equal(t, "0.0.0.0", host)
	require.Equal(t, 80, port)
	require.Equal(t, []User{
		User{
			Name:     "admin",
			Password: "admin",
		},
		User{
			Name:     "user",
			Password: "secret",
		},
	}, users)
}

func TestAddDefaultUser(t *testing.T) {
	defaultInit()
	addDefaultUser()
	require.Equal(t, []User{}, users)

	username = "admin"
	addDefaultUser()
	require.Equal(t, []User{}, users)

	username = ""
	password = "admin"
	addDefaultUser()
	require.Equal(t, []User{}, users)

	username = "me"
	password = "secret"
	addDefaultUser()
	require.Equal(t, []User{
		User{
			Name:     "me",
			Password: "secret",
		},
	}, users)
}
