package property


type Config struct {
	URL        string `json:"url"`
	Token      string `json:"token"`
	Intent     string `json:"intent"`
	Reserve    int    `json:"reserve"`
	AllowTools bool   `json:"allowTools"`
	TimeoutSec int    `json:"timeout"`
	ShowSeq    bool   `json:"showSeq"`
}


func GetDefaultConfig() *Config {
	return &Config{
		URL:        "ws://127.0.0.1:8787/ws",
		Token:      "abc",
		Intent:     "qa",
		Reserve:    512,
		AllowTools: true,
		TimeoutSec: 120,
		ShowSeq:    false,
	}
}
