package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestJobStatusMarshalJSON(t *testing.T) {
	data, err := json.Marshal(Failed)
	if err != nil {
		panic(err)
	}
	if string(data) != "\"failed\"" {
		t.Fatalf("TestJobStatusMarshalJSON: unexpected result, got %s instead of %s", string(data), "\"failed\"")
	}
}

func TestJobStatusUnmarshalJSON(t *testing.T) {
	var status JobStatus

	err := json.Unmarshal([]byte("\"inprogress\""), &status)
	if err != nil {
		panic(err)
	}
	if status != InProgress {
		t.Fatalf("TestJobStatusUnmarshalJSON: unexpected result, got %s instead of %s", status.String(), InProgress.String())
	}
}

func TestEmptyProject(t *testing.T) {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "lurch-test-workdir")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)

	jobPath := filepath.Join(tmpdir, "build-1")

	b := &Job{name: "1", dir: jobPath}

	if s := b.Status(); s != Unknown {
		t.Fatalf("TestEmptyProject: unexpected status of empty job, got %s instead of %s", s.String(), Unknown.String())
	}

	if output, err := b.ReadOutput(); err == nil || output != "" {
		t.Fatalf("TestEmptyProject: unexpected output, should be empty")
	}

	if sd := b.StartDate(); sd != time.UnixMicro(0) {
		t.Fatalf("TestEmptyProject: unexpected start time of empty job, got %s instead of %s", sd.String(), time.UnixMicro(0).String())
	}

	if ed := b.EndDate(); ed != time.UnixMicro(0) {
		t.Fatalf("TestEmptyProject: unexpected end time of empty job, got %s instead of %s", ed.String(), time.UnixMicro(0).String())
	}

	if err = b.LogStart(); err == nil {
		t.Fatalf("TestEmptyProject: expected error during log start, when no directory exist")
	}
}

func TestStatus(t *testing.T) {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "lurch-test-workdir")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)

	jobPath := filepath.Join(tmpdir, "build-1")
	if err := os.Mkdir(jobPath, 0755); err != nil {
		panic(err)
	}

	b := &Job{name: "1", dir: jobPath}
	if err := b.SetStatus(Stopped); err != nil {
		panic(err)
	}
	if s := b.Status(); s != Stopped {
		t.Fatalf("TestStatus: unexpected status, got %s instead of %s", s.String(), Stopped.String())
	}
}

func TestStartDate(t *testing.T) {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "lurch-test-workdir")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)

	jobPath := filepath.Join(tmpdir, "build-1")
	if err := os.Mkdir(jobPath, 0755); err != nil {
		panic(err)
	}

	b := &Job{name: "1", dir: jobPath}

	if err := b.LogStart(); err != nil {
		panic(err)
	}
	if sd := b.StartDate(); sd == time.UnixMicro(0) {
		t.Fatalf("TestStartDate: unexpected start date, got 0 instead of current time")
	}
}

func TestEndDate(t *testing.T) {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "lurch-test-workdir")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)

	jobPath := filepath.Join(tmpdir, "build-1")
	if err := os.Mkdir(jobPath, 0755); err != nil {
		panic(err)
	}

	b := &Job{name: "1", dir: jobPath}

	if err := b.SetStatus(Finished); err != nil {
		panic(err)
	}
	if ed := b.EndDate(); ed == time.UnixMicro(0) {
		t.Fatalf("TestEndDate: unexpected end date, got 0 instead of current time")
	}
}

func TestWorkspace(t *testing.T) {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "lurch-test-workdir")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)

	jobPath := filepath.Join(tmpdir, "build-1")
	if err := os.Mkdir(jobPath, 0755); err != nil {
		panic(err)
	}

	b := &Job{name: "1", dir: jobPath}

	if !strings.HasPrefix(b.WorkspacePath(), jobPath) {
		t.Fatalf("TestWorkspace: unexpected workspace path '%s', should start with '%s'", b.WorkspacePath(), jobPath)
	}

	if err := b.MkWorkspace(); err != nil {
		panic(err)
	}
}

func TestOutput(t *testing.T) {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "lurch-test-workdir")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)

	jobPath := filepath.Join(tmpdir, "build-1")
	if err := os.Mkdir(jobPath, 0755); err != nil {
		panic(err)
	}

	b := &Job{name: "1", dir: jobPath}

	expectedData := "some data to be read"

	if err = os.WriteFile(b.OutputPath(), []byte(expectedData), 0644); err != nil {
		panic(err)
	}

	if data, err := b.ReadOutput(); err == nil {
		if data != expectedData {
			t.Fatalf("TestOutput: unexpected output, got '%s' instead of '%s'", data, expectedData)
		}
	} else {
		panic(err)
	}
}

func TestEquals(t *testing.T) {
	j1 := &Job{name: "1", p: &Project{name: "project1"}, interrupt: make(chan bool)}
	j2 := &Job{name: "1", p: &Project{name: "project1"}}
	j3 := &Job{name: "1", p: &Project{name: "project2"}}
	j4 := &Job{name: "2", p: &Project{name: "project1"}}
	j5 := &Job{name: "1"}

	if !j1.Equals(j2) {
		t.Fatalf("TestEquals: j1 and j2 should equal")
	}

	if j1.Equals(j3) {
		t.Fatalf("TestEquals: j1 and j3 shouldn't equal")
	}

	if j1.Equals(j4) {
		t.Fatalf("TestEquals: j1 and j4 shouldn't equal")
	}

	if j1.Equals(j5) {
		t.Fatalf("TestEquals: j1 and j3 shouldn't equal")
	}

	if j5.Equals(j2) {
		t.Fatalf("TestEquals: j2 and j5 shouldn't equal")
	}
}

func TestSetParams(t *testing.T) {
	b := &Job{}

	b.SetParams(map[string]string{
		"1":   "value1",
		"k2":  "value2",
		"k-3": "value3",
		"k_4": "value4",
	})

	if p := b.params["1"]; p != "" {
		t.Fatalf("TestSetParams: unexpected value for key 1: '%s'", p)
	}

	if p := b.params["K2"]; p != "value2" {
		t.Fatalf("TestSetParams: unexpected value for key k2: '%s'", p)
	}

	if p := b.params["K-3"]; p != "" {
		t.Fatalf("TestSetParams: unexpected value for key k-3: '%s'", p)
	}

	if p := b.params["K_4"]; p != "value4" {
		t.Fatalf("TestSetParams: unexpected value for key k_4: '%s'", p)
	}
}

func TestSaveLoadParams(t *testing.T) {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "lurch-test-workdir")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)

	jobPath := filepath.Join(tmpdir, "build-1")
	if err := os.Mkdir(jobPath, 0755); err != nil {
		panic(err)
	}

	b := &Job{dir: jobPath}
	b.SetParams(map[string]string{"k1": "v1", "k2": "v2"})

	if err := b.SaveParams(); err != nil {
		t.Fatal("TestSaveParams: unexpected error", err)
	}

	b2 := &Job{dir: jobPath}
	b2.LoadParams()

	if len(b2.params) != 2 {
		t.Fatal("TestSaveParams: unexpected length of loaded params")
	}

	fmt.Println(b2.params)
}
