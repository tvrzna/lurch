package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Context struct {
	mutex     *sync.Mutex
	interrupt chan bool
	conf      *Config
	jobs      []*Job
	webServer *http.Server
}

// Init new context
func NewContext(c *Config) *Context {
	return &Context{conf: c, jobs: make([]*Job, 0), mutex: &sync.Mutex{}, interrupt: make(chan bool)}
}

func (c *Context) StartJob(p *Project) string {
	// Check if project is being built
	if c.isProjectBeingBuilt(p) {
		return ""
	}

	b, err := p.NewJob()
	if err != nil {
		return ""
	}

	c.mutex.Lock()
	c.jobs = append(c.jobs, b)
	c.mutex.Unlock()

	c.removeOldjobs(p)

	go c.start(b)
	return b.name
}

func (c *Context) Interrupt(b *Job) {
	c.mutex.Lock()
	for _, job := range c.jobs {
		if b.Equals(job) {
			log.Printf("-- interrupting job #%s of %s", b.name, b.p.name)
			job.interrupt <- true
		}
	}
	c.mutex.Unlock()
}

func (c *Context) InterruptAll() {
	c.mutex.Lock()
	for _, job := range c.jobs {
		log.Printf("-- interrupting job #%s of %s", job.name, job.p.name)
		job.interrupt <- true
	}
	c.mutex.Unlock()
}

func (c *Context) start(b *Job) {
	log.Printf(">> started job #%s for %s", b.name, b.p.name)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command(filepath.Join(b.p.dir, "script.cmd"))
	} else {
		cmd = exec.Command("sh", "-c", filepath.Join(b.p.dir, "script.sh"))
	}

	b.MkWorkspace()
	b.LogStart()
	cmd.Dir = b.WorkspacePath()
	output, err := os.OpenFile(b.OutputPath(), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		b.SetStatus(Failed)
		c.removeFromSlice(b)
		log.Printf("-- failed to open output for #%s of %s", b.name, b.p.name)
		return
	}
	defer output.Close()
	cmd.Stdin = output
	cmd.Stdout = output
	cmd.Stderr = output

	cmd.Start()
	go c.watchForInterrupt(b, cmd)

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if exiterr.ExitCode() == 0 {
				b.SetStatus(Finished)
			} else if b.Status() != Stopped {
				b.SetStatus(Failed)
			}
			output.WriteString(exiterr.String())
		} else {
			output.WriteString("\nFailed!")
			b.SetStatus(Failed)
		}
	} else {
		b.SetStatus(Finished)
	}
	close(b.interrupt)
	c.removeFromSlice(b)

	if err := c.compressFolder(b.WorkspacePath()+".tar.gz", b.WorkspacePath()); err != nil {
		log.Print("-- could not compress", b.WorkspacePath())
	}
	os.RemoveAll(b.WorkspacePath())

	log.Printf("<< finished job #%s for %s", b.name, b.p.name)
}

func (c *Context) removeFromSlice(b *Job) {
	c.mutex.Lock()
	if index := c.indexOf(b); index >= 0 {
		if len(c.jobs) == 1 {
			c.jobs = c.jobs[:0]
		} else {
			c.jobs = append(c.jobs[:index], c.jobs[index+1])
		}
	}
	c.mutex.Unlock()
}

func (c *Context) indexOf(b *Job) int {
	for i, job := range c.jobs {
		if b.Equals(job) {
			return i
		}
	}
	return -1
}

func (c *Context) watchForInterrupt(b *Job, cmd *exec.Cmd) {
	status := <-b.interrupt
	if status {
		b.SetStatus(Stopped)
		cmd.Process.Signal(os.Kill)
	}
}

func (c *Context) IsBeingBuilt(b *Job) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.indexOf(b) >= 0
}

func (c *Context) isProjectBeingBuilt(p *Project) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, b := range c.jobs {
		if b.p.name == p.name {
			return true
		}
	}
	return false
}

func (c *Context) removeOldjobs(p *Project) {
	jobs, _ := c.ListJobs(p)
	if len(jobs) > 10 {
		jobs = jobs[10:]
		log.Print("-- remove old jobs of ", p.name)
		for _, b := range jobs {
			os.RemoveAll(b.dir)
		}
	}
}

func (c *Context) ListProjects() ([]*Project, error) {
	result := make([]*Project, 0)
	entries, err := os.ReadDir(c.conf.path)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			result = append(result, &Project{name: e.Name(), dir: c.conf.path + "/" + e.Name()})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return strings.Compare(result[i].name, result[j].name) <= 0
	})

	return result, nil
}

func (c *Context) OpenProject(name string) *Project {
	return &Project{name: name, dir: c.conf.path + "/" + name}
}

func (c *Context) ListJobs(p *Project) ([]*Job, error) {
	result := make([]*Job, 0)
	entries, err := os.ReadDir(p.dir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			result = append(result, &Job{name: e.Name(), dir: filepath.Join(p.dir, e.Name()), p: p})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		bI, _ := strconv.Atoi(result[i].name)
		bJ, _ := strconv.Atoi(result[j].name)
		return bI > bJ
	})

	return result, nil
}

func (c *Context) OpenJob(p *Project, name string) *Job {
	return &Job{name: name, dir: filepath.Join(p.dir, name), p: p}
}

func (c *Context) compressFolder(outputPath, inputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	zw := gzip.NewWriter(file)
	tw := tar.NewWriter(zw)

	err = filepath.Walk(inputPath, func(file string, fi os.FileInfo, e error) error {
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		header.Name = strings.Replace(filepath.ToSlash(file), inputPath, "", 1)
		if header.Name == "" {
			return nil
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}

	if err := zw.Close(); err != nil {
		return err
	}
	return nil
}
