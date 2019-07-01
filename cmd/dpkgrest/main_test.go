package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHttpServer(t *testing.T) {
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
	databaseFile = filepath.Join("..", "..", "testdata", "status")
	ts := httptest.NewServer(http.HandlerFunc(callback))
	defer ts.Close()

	res, err := http.Get(ts.URL + "/list?filter=lib)")
	require.Nil(t, err)
	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	res.Body.Close()
}

func TestHttpServerWithInvalidDatabase(t *testing.T) {
	databaseFile = "invalid"
	ts := httptest.NewServer(http.HandlerFunc(callback))
	defer ts.Close()

	res, err := http.Get(ts.URL + "/list")
	require.Nil(t, err)
	require.Equal(t, http.StatusInternalServerError, res.StatusCode)

	res.Body.Close()
}

func TestHttpServerWithInvalidPath(t *testing.T) {
	databaseFile = filepath.Join("..", "..", "testdata", "status")
	ts := httptest.NewServer(http.HandlerFunc(callback))
	defer ts.Close()

	res, err := http.Get(ts.URL + "/wrong")
	require.Nil(t, err)
	require.Equal(t, http.StatusNotFound, res.StatusCode)

	res.Body.Close()
}
