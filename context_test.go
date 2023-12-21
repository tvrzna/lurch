package main

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestContext(t *testing.T) {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "lurch-test-workdir")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)

	conf := LoadConfig([]string{"-t", tmpdir})

	c := NewContext(conf)
	NewWebSocketService(c)

	p1 := c.OpenProject("project-1")
	if err = os.MkdirAll(p1.dir, 0755); err != nil {
		panic(err)
	}
	os.WriteFile(filepath.Join(p1.dir, "script.sh"), []byte("#!/bin/sh\n\necho Project-1 run\nsleep 1"), 0755)

	p2 := c.OpenProject("project-2")
	if err = os.MkdirAll(p2.dir, 0755); err != nil {
		panic(err)
	}
	os.WriteFile(filepath.Join(p2.dir, "script.sh"), []byte("#!/bin/sh\n\nmkdir target\ntouch target/data\necho Project-2 run\nsleep 60"), 0755)
	for i := 0; i < 12; i++ {
		p2.NewJob()
	}

	if projects, err := c.ListProjects(); err != nil {
		panic(err)
	} else if len(projects) != 2 {
		t.Fatalf("TestContext: 2 project were expected, but found %d", len(projects))
	}

	c.StartJob(p1, map[string]string{"key1": "val1", "_key2": "val2"})
	time.Sleep(3 * time.Second)

	if jobs, err := c.ListJobs(p1); err != nil {
		panic(err)
	} else if s := jobs[0].Status(); s != Finished {
		t.Fatalf("TestContext: job ends with unexpected status %s", s.String())
	}

	if job := c.OpenJob(p1, "1"); err != nil {
		panic(err)
	} else if s := job.Status(); s != Finished {
		t.Fatalf("TestContext: job ends with unexpected status %s", s.String())
	}

	c.StartJob(p2, nil)
	time.Sleep(2 * time.Second)
	j := c.OpenJob(p2, strconv.Itoa(p2.LastCount()))
	if !c.IsBeingBuilt(j) {
		t.Fatalf("TestContext: job #1 for project-2 should be still running")
	}
	c.Interrupt(j)
	time.Sleep(2 * time.Second)
	if c.IsBeingBuilt(j) {
		t.Fatalf("TestContext: job #1 for project-2 should not be running")
	}
	if s := j.Status(); s != Stopped {
		t.Fatalf("TestContext: job #1 for project-2 has status %s instead of stopped", s.String())
	}

}
