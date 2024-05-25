package tunnel

import (
	"context"
	"fmt"

	"github.com/taodev/gotun/option"
)

func New(ctx context.Context, options option.Tunnel) (*TunnelTCP, error) {
	switch options.Type {
	case "tcp":
		return NewTunnelTCP(options), nil
	}

	return nil, fmt.Errorf("unknown tunnel type: %s", options.Type)
}
