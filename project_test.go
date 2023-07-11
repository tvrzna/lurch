package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProject(t *testing.T) {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "lurch-test-workdir")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)

	projectPath := filepath.Join(tmpdir, "test-project")
	if err := os.Mkdir(projectPath, 0755); err != nil {
		panic(err)
	}

	p := &Project{name: "test-project", dir: projectPath}

	if p.LastCount() != 0 {
		t.Fatalf("TestProject: unexpected last count %d", p.LastCount())
	}

	i, err := p.RotateCount()
	if err != nil || i != 1 {
		t.Fatal("TestProject: rotate count failed")
	}

	if p.LastCount() != 1 {
		t.Fatalf("TestProject: unexpected last count %d", p.LastCount())
	}

	_, err = p.NewJob()
	if err != nil {
		panic(err)
	}

	if p.LastCount() != 2 {
		t.Fatalf("TestProject: unexpected last count %d", p.LastCount())
	}
}

func TestSetParamsOnProject(t *testing.T) {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "lurch-test-workdir")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)

	projectPath := filepath.Join(tmpdir, "test-project")
	if err := os.Mkdir(projectPath, 0755); err != nil {
		panic(err)
	}

	p := &Project{name: "test-project", dir: projectPath}

	p.SetParams(map[string]string{
		"1":   "value1",
		"k2":  "value2",
		"k-3": "value3",
		"k_4": "value4",
	})

	if v := p.params["1"]; v != "" {
		t.Fatalf("TestSetParamsOnProject: unexpected value for key 1: '%s'", v)
	}

	if v := p.params["K2"]; v != "value2" {
		t.Fatalf("TestSetParamsOnProject: unexpected value for key k2: '%s'", v)
	}

	if v := p.params["K-3"]; v != "" {
		t.Fatalf("TestSetParamsOnProject: unexpected value for key k-3: '%s'", v)
	}

	if v := p.params["K_4"]; v != "value4" {
		t.Fatalf("TestSetParamsOnProject: unexpected value for key k_4: '%s'", v)
	}
}

func TestSaveLoadParamsOnProject(t *testing.T) {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "lurch-test-workdir")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)

	projectPath := filepath.Join(tmpdir, "test-project")
	if err := os.Mkdir(projectPath, 0755); err != nil {
		panic(err)
	}

	p := &Project{name: "test-project", dir: projectPath}

	p.SetParams(map[string]string{"k1": "v1", "k2": "v2"})

	if err := p.SaveParams(); err != nil {
		t.Fatal("TestSaveParams: unexpected error", err)
	}

	p2 := &Project{dir: projectPath}
	p2.LoadParams()

	if len(p2.params) != 2 {
		t.Fatal("TestSaveParams: unexpected length of loaded params")
	}
}
