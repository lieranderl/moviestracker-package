package trackers

type ConfigBuilder struct {
	config ParserConfig
}

func NewConfigBuilder(trackerType TrackerType, baseURL string) *ConfigBuilder {
	return &ConfigBuilder{
		config: ParserConfig{
			Type:          trackerType,
			Name:          string(trackerType),
			BaseURL:       baseURL,
			RequestConfig: DefaultRequestConfig,
		},
	}
}

func (b *ConfigBuilder) WithAuth(authConfig *AuthConfig) *ConfigBuilder {
	b.config.RequiresAuth = true
	b.config.AuthConfig = authConfig
	return b
}

func (b *ConfigBuilder) WithParseConfig(parseConfig *ParseConfig) *ConfigBuilder {
	b.config.ParseConfig = parseConfig
	return b
}

func (b *ConfigBuilder) WithRequestConfig(requestConfig *RequestConfig) *ConfigBuilder {
	b.config.RequestConfig = requestConfig
	return b
}

func (b *ConfigBuilder) Build() ParserConfig {
	return b.config
}
