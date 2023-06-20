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
