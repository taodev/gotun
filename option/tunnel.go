package option

type _Tunnel struct {
	Type           string `yaml:"type"`
	Tag            string `yaml:"tag,omitempty"`
	Addr           string `yaml:"addr"`
	Password       string `yaml:"password,omitempty"`
	TargetAddr     string `yaml:"target_addr,omitempty"`
	TargetPassword string `yaml:"target_password,omitempty"`
	Compression    bool   `yaml:"compression,omitempty"`
}

type Tunnel _Tunnel
