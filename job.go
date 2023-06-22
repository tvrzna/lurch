package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type JobStatus byte

const (
	Unknown JobStatus = iota
	Finished
	Stopped
	Failed
	InProgress
)

// Stringify job status
func (b JobStatus) String() string {
	return []string{"unknown", "finished", "stopped", "failed", "inprogress"}[int(b)]
}

// Marshal job status type into string
func (b JobStatus) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("\"")
	buffer.WriteString(b.String())
	buffer.WriteString("\"")

	return buffer.Bytes(), nil
}

// Unmarshal string job status into type
func (b *JobStatus) UnmarshalJSON(data []byte) error {
	var v string
	json.Unmarshal(data, &v)

	*b = map[string]JobStatus{
		"unknown":    Unknown,
		"finished":   Finished,
		"stopped":    Stopped,
		"failed":     Failed,
		"inprogress": InProgress,
	}[v]

	return nil
}

type Job struct {
	name      string
	dir       string
	p         *Project
	interrupt chan bool
}

// Get status of job
func (b *Job) Status() JobStatus {
	data, err := os.ReadFile(filepath.Join(b.dir, "status"))
	if err != nil {
		return Unknown
	}
	i, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	return JobStatus(byte(i))
}

// Set status of job
func (b *Job) SetStatus(s JobStatus) error {
	return os.WriteFile(filepath.Join(b.dir, "status"), []byte(strconv.Itoa(int(s))), 0600)
}

// Get path to console ouput of job
func (b *Job) OutputPath() string {
	return filepath.Join(b.dir, "console.log")
}

// Read content of console output of job
func (b *Job) ReadOutput() (string, error) {
	data, err := os.ReadFile(b.OutputPath())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Get start date of job
func (b *Job) StartDate() time.Time {
	s, err := os.Stat(filepath.Join(b.dir, "start"))
	if err != nil {
		return time.UnixMicro(0)
	}
	return s.ModTime()
}

// Get end date of job
func (b *Job) EndDate() time.Time {
	s, err := os.Stat(filepath.Join(b.dir, "status"))
	if err != nil {
		return time.UnixMicro(0)
	}
	return s.ModTime()
}

// Path to workspace
func (b *Job) WorkspacePath() string {
	return filepath.Join(b.dir, "workspace")
}

// Make workspace directory
func (b *Job) MkWorkspace() error {
	return os.MkdirAll(b.WorkspacePath(), 0755)
}

// Creates file for tracking the time of start
func (b *Job) LogStart() error {
	file, err := os.OpenFile(filepath.Join(b.dir, "start"), os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	return file.Close()
}

// Path to artifact archive
func (b *Job) ArtifactPath() string {
	return filepath.Join(b.dir, "workspace.tar.gz")
}

// Checks if jobs are equal
func (b *Job) Equals(other *Job) bool {
	return b.p != nil && other.p != nil && b.p.name == other.p.name && b.name == other.name
}
