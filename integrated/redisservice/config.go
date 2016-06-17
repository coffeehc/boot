package redisservice

import "time"

type RedisConfig struct {
	Cluster            bool          `yaml:"cluster"`
	Addrs              []string      `yaml:"addrs"`
	MaxRedirects       int           `yaml:"maxRedirects"`
	Password           string        `yaml:"password"`
	DialTimeout        time.Duration `yaml:"dialTimeout"`
	ReadTimeout        time.Duration `yaml:"readTimeout"`
	WriteTimeout       time.Duration `yaml:"writeTimeout"`
	PoolSize           int           `yaml:"poolSize"`
	PoolTimeout        time.Duration `yaml:"poolTimeout"`
	IdleTimeout        time.Duration `yaml:"idleTimeout"`
	IdleCheckFrequency time.Duration `yaml:"idleCheckFrequency"`
}
