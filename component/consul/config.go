package consul

import "time"

type Config struct {
	Address    string        `mapstructure:"address,omitempty" json:"address,omitempty"`
	Token      string        `mapstructure:"token,omitempty" json:"token,omitempty"`
	Datacenter string        `mapstructure:"datacenter,omitempty" json:"datacenter,omitempty"`
	Namespace  string        `mapstructure:"namespace,omitempty" json:"namespace,omitempty"`
	TokenFile  string        `mapstructure:"token_file,omitempty" json:"token_file,omitempty"`
	WaitTime   time.Duration `mapstructure:"wait_time,omitempty" json:"wait_time,omitempty"`
}
