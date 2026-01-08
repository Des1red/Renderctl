package avtransport

type Capabilities struct {
	Actions map[string]bool
	Media   map[string][]string // mime -> profiles
}
