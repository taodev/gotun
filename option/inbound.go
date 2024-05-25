package option

type _Inbound struct {
	Type   string `yaml:"type"`
	Listen string `yaml:"listen"`
}

type Inbound _Inbound
