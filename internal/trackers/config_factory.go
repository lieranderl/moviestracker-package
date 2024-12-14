package trackers

func NewRutorConfig(baseURL string) ParserConfig {
	return ParserConfig{
		Type:    TrackerRutor,
		Name:    "Rutor",
		BaseURL: baseURL,

		RequiresAuth: false,

		ParseConfig: &ParseConfig{
			MoviePattern:  "",
			SeriesPattern: "[",
			ExcludePatterns: []string{
				"PROPER", "REPACK",
			},
			DateFormat:    "02.01.2006",
			TitleSelector: "tr.gai,tr.tum",
			SizeSelector:  "tr td:nth-child(4)",
			SeedsSelector: "tr td:nth-child(5)",
			LeechSelector: "tr td:nth-child(6)",
		},

		RequestConfig: DefaultRequestConfig,
	}
}

func NewKinozalConfig(baseURL string) ParserConfig {
	return ParserConfig{
		Type:    TrackerKinozal,
		Name:    "Kinozal",
		BaseURL: baseURL,

		RequiresAuth: true,
		AuthConfig: &AuthConfig{
			LoginURL:       baseURL + "/takelogin.php",
			LoginEnvVar:    "KZ_LOGIN",
			PasswordEnvVar: "KZ_PASSWORD",
		},

		ParseConfig: &ParseConfig{
			MoviePattern:  "",
			SeriesPattern: "сезон",
			ExcludePatterns: []string{
				"Трейлер", "Анонс",
			},
			DateFormat:    "2006-01-02T15:04:05.000Z",
			TitleSelector: "tr.bg",
			SizeSelector:  "td:nth-child(4)",
			SeedsSelector: "td:nth-child(5)",
			LeechSelector: "td:nth-child(6)",
		},

		RequestConfig: DefaultRequestConfig,
	}
}
