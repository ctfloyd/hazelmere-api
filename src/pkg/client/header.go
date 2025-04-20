package client

type Option func(map[string]string)

func makeHeadersFromConfig(config HazelmereConfig) map[string]string {
	return makeHeaders(
		withToken(config.Token),
		withCallingApplication(config.CallingApplication),
	)
}

func makeHeaders(opts ...Option) map[string]string {
	headers := make(map[string]string)
	for _, opt := range opts {
		opt(headers)
	}
	return headers
}

func withToken(token string) Option {
	return func(headers map[string]string) {
		headers["Authorization"] = "Bearer " + token
	}
}

func withCallingApplication(application string) Option {
	return func(headers map[string]string) {
		headers["x-hz-caller"] = application
	}
}
