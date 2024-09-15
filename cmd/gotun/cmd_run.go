package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime"
	runtimeDebug "runtime/debug"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/taodev/gotun"
	"github.com/taodev/gotun/option"
)

// 添加service

var commandRun = &cobra.Command{
	Use:   "run",
	Short: "Run service",
	Run: func(cmd *cobra.Command, args []string) {
		err := run()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	mainCommand.AddCommand(commandRun)
}

func readConfig(path string) (*option.Options, error) {
	configContent, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var options option.Options
	err = options.UnmarshalYAML(configContent)
	if err != nil {
		return nil, err
	}

	return &options, nil
}

func create() (*gotun.GoTun, context.CancelFunc, error) {
	// 设置进程数
	runtime.GOMAXPROCS(runtime.NumCPU())

	options, err := readConfig(configPath)
	if err != nil {
		return nil, nil, err
	}
	ctx, cancel := context.WithCancel(globalCtx)
	instance, err := gotun.New(gotun.Options{
		Options: *options,
		Context: ctx,
	})
	if err != nil {
		cancel()
		return nil, nil, err
	}

	err = instance.Start()
	if err != nil {
		cancel()
		return nil, nil, err
	}

	return instance, cancel, nil
}

func run() error {
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(osSignals)

	for {
		instance, cancel, err := create()
		if err != nil {
			return err
		}
		runtimeDebug.FreeOSMemory()
		for {
			osSignal := <-osSignals
			if osSignal == syscall.SIGHUP {
				break
			}
			cancel()
			instance.Close()
			if osSignal != syscall.SIGHUP {
				return nil
			}
			break
		}
	}
}
