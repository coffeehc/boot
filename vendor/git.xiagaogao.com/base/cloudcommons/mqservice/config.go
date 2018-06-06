package mqservice

import (
	"fmt"
	"net/url"
	"os"
)

type VhostConfig struct {
	Vhost    string `yaml:"vhost"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	MQAddr   string `yaml:"mq_addr"`
}

func (config *VhostConfig) toUrl() string {
	return fmt.Sprintf("amqp://%s:%s@%s/%s", url.QueryEscape(config.User), url.QueryEscape(config.Password), config.MQAddr, config.Vhost)
}

func WarpEnvVhostConfig(config *VhostConfig) {
	mqAddr, ok := os.LookupEnv("ENV_MQ_ADDR")
	if ok {
		config.MQAddr = mqAddr
	}
}
