package servers

type StreamSource interface {
	Open() (StreamReadCloser, error)
}

type StreamContainer interface {
	Key() string
	MimeCandidates() []string
}

type StreamReadCloser interface {
	Read(p []byte) (int, error)
	Close() error
}
