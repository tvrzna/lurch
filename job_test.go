package main

import (
	"encoding/json"
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
