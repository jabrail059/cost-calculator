package handlers

import "time"

type Config struct {
	OneCURL         string
	OneCTimeout     time.Duration
	UploadMaxMemory int64
}

var handlerConfig = Config{
	OneCTimeout:     30 * time.Second,
	UploadMaxMemory: 10 << 20,
}

func Configure(cfg Config) {
	handlerConfig = cfg
}
