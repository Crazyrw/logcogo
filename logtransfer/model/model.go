package model

type Config struct {
	KafkaConfig `ini:"kafka"`
	ESConfig    `ini:"es"`
}

type KafkaConfig struct {
	Address string `ini:"address"`
	Topic   string `ini:"topic"`
}

type ESConfig struct {
	Address       string `ini:"address"`
	Index         string `ini:"index"`
	ChanSize      int64  `ini:"chan_size"`
	GoroutineSize int64  `ini:"goroutine_size"`
}
