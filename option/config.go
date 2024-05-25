package option

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

type _Options struct {
	Tunnels   []Tunnel   `yaml:"tunnels,omitempty"`
	Inbounds  []Inbound  `yaml:"inbounds,omitempty"`
	Outbounds []Outbound `yaml:"outbounds,omitempty"`
}

type Options _Options

func (o *Options) UnmarshalYAML(content []byte) error {
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)
	err := decoder.Decode((*_Options)(o))
	if err != nil {
		return err
	}
	return nil
}
