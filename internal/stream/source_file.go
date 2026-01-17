package stream

import (
	"errors"
	"net/http"
	"os"
	"os/exec"
	"renderctl/logger"
)

type fileSource struct{ path string }

func (f fileSource) Open() (StreamReadCloser, error) {
	return os.Open(f.path)
}

type urlSource struct {
	url string
}

func (u urlSource) Open() (StreamReadCloser, error) {
	resp, err := http.Get(u.url)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

type rollingFileSource struct {
	path    string
	cmd     *exec.Cmd
	started bool
	opened  bool
}

func (r *rollingFileSource) Open() (StreamReadCloser, error) {
	logger.Notify("Opening rolling file source")

	if r.opened {
		return nil, errors.New("screen stream already active")
	}

	if r.started && r.cmd.ProcessState != nil && r.cmd.ProcessState.Exited() {
		logger.Notify("ffmpeg exited, restarting")
		r.started = false
	}

	if !r.started {
		if err := r.cmd.Start(); err != nil {
			return nil, err
		}
		r.started = true
		logger.Notify("ffmpeg started")
	}

	rc, err := os.Open(r.path)
	if err != nil {
		return nil, err
	}

	r.opened = true
	return rc, nil
}

func (r *rollingFileSource) Close() error {
	r.opened = false

	if r.cmd != nil && r.cmd.Process != nil {
		_ = r.cmd.Process.Kill()
		_, _ = r.cmd.Process.Wait()
	}

	return nil
}
