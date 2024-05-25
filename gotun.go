package gotun

import (
	"context"
	"fmt"
	"time"

	"github.com/taodev/gotun/option"
	"github.com/taodev/gotun/tunnel"
)

type GoTun struct {
	createAt time.Time
	tunnels  []*tunnel.TunnelTCP
}

type Options struct {
	option.Options
	Context context.Context
}

func New(options Options) (*GoTun, error) {
	createAt := time.Now()
	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}

	tunnels := make([]*tunnel.TunnelTCP, 0, len(options.Tunnels))
	for i, tunnelOptions := range options.Tunnels {
		in, err := tunnel.New(ctx, tunnelOptions)
		if err != nil {
			return nil, fmt.Errorf("parse tunnel[%d]: %w", i, err)
		}
		tunnels = append(tunnels, in)
	}

	return &GoTun{
		createAt: createAt,
		tunnels:  tunnels,
	}, nil
}

func (t *GoTun) Close() error {
	return nil
}

func (t *GoTun) Start() (err error) {
	for _, tunnel := range t.tunnels {
		err = tunnel.Start()
		if err != nil {
			return
		}
	}
	return
}
