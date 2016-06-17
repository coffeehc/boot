package serviceboot

type DebugConfig struct {
	EnableAccessInfo bool `yaml:"enableAccessInfo"`
}

func (this DebugConfig) IsEnableAccessInfo() bool {
	return this.EnableAccessInfo
}
