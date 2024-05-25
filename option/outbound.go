package option

type _Outbound struct {
	Type     string `yaml:"type"`
	Addr     string `yaml:"address"`
	Password string `yaml:"password,omitempty"`
}

type Outbound _Outbound
