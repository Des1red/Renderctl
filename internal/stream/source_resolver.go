package stream

import (
	"io"
	"os/exec"
	"renderctl/internal/servers"
	"renderctl/logger"
)

type resolverSource struct {
	url string
	cmd *exec.Cmd
}

type resolverReadCloser struct {
	io.ReadCloser
	cmd *exec.Cmd
}

func (r *resolverReadCloser) Close() error {
	if r.cmd != nil && r.cmd.Process != nil {
		_ = r.cmd.Process.Kill()
		_, _ = r.cmd.Process.Wait()
	}
	return r.ReadCloser.Close()
}

func newResolverSource(url string) *resolverSource {
	return &resolverSource{url: url}
}

func (r *resolverSource) Open() (servers.StreamReadCloser, error) {
	logger.Status("Starting media resolver (yt-dlp + ffmpeg)")

	// yt-dlp command:
	// - output to stdout
	// - merge audio+video
	// - force TS container (stream-safe)
	cmd := exec.Command(
		"sh", "-c",
		`yt-dlp -f "bv*[vcodec^=avc1]+ba/best" -o - "`+r.url+`" | \
ffmpeg -loglevel error -i pipe:0 -f mpegts -codec copy pipe:1`,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	// cmd.Stderr = os.Stderr
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, err
	} else {
		logger.Done("Media resolver started")
	}
	go func() {
		if err := cmd.Wait(); err != nil {
			logger.Notify("Media resolver exited: %v", err)
		}
	}()

	r.cmd = cmd

	return &resolverReadCloser{
		ReadCloser: stdout,
		cmd:        cmd,
	}, nil

}
