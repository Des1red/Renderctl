package avtransport

type Capabilities struct {
	Actions map[string]bool
	Media   map[string][]string // mime -> profiles
}

func EnrichCapabilities(
	avScpdURL string,
	connMgrControlURL string,
	target Target,
) (*Capabilities, error) {

	actions, err := FetchActions(avScpdURL)
	if err != nil {
		return nil, err
	}

	validated := ValidateActions(target)
	for k, v := range validated {
		actions[k] = v
	}

	media, err := FetchMediaProtocols(connMgrControlURL)
	if err != nil {
		media = map[string][]string{}
	}

	return &Capabilities{
		Actions: actions,
		Media:   media,
	}, nil
}
