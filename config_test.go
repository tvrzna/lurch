package main

import (
	"os"
	"strconv"
	"testing"
)

func TestLoadConfigWithData(t *testing.T) {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "lurch-test-workdir")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)

	port := 12345
	path := tmpdir
	url := "https://lurch.tst/"

	c := LoadConfig([]string{"-p", strconv.Itoa(port), "-t=" + path, "-a=" + url})

	if c.port != port {
		t.Fatalf("TestLoadConfigWithData: unexpected port: %d instead of %d", c.port, port)
	}

	if c.path != path {
		t.Fatalf("TestLoadConfigWithData: unexpected path: '%s' instead of '%s'", c.path, path)
	}

	if c.getAppUrl() != url {
		t.Fatalf("TestLoadConfigWithData: unexpected app-url: '%s' instead of '%s'", c.getAppUrl(), url)
	}

	if c.getServerUri() != "localhost:"+strconv.Itoa(port) {
		t.Fatalf("TestLoadConfigWithData: unexpected server uri: '%s' instead of '%s'", c.getServerUri(), "localhost:"+strconv.Itoa(port))
	}

	if c.GetVersion() != "develop" {
		t.Fatalf("TestLoadConfigWithData: unexpected version: '%s' instead of '%s'", c.GetVersion(), "develop")
	}
}

func TestGetAppUrlEmpty(t *testing.T) {
	c := LoadConfig([]string{})
	if c.getAppUrl() != "http://localhost:5000" {
		t.Fatalf("TestGetAppUrlEmpty: unexpected default app url '%s'", "http://localhost:5000")
	}
}

func TestGetVersion(t *testing.T) {
	c := LoadConfig([]string{})
	buildVersion = "test"
	if c.GetVersion() != "test" {
		t.Fatalf("TestGetVersion: unexpected version: '%s' instead of '%s'", c.GetVersion(), "test")
	}
}

func TestWrongPath(t *testing.T) {
	path := "$dev/////nul////l" ////

	c := LoadConfig([]string{"-t"})
	c.setPath(path)

	if c.path == path {
		t.Fatalf("TestWrongPath: path should not be '%s'", path)
	}
}
