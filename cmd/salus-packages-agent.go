/*
 * Copyright 2020 Rackspace US, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"github.com/itzg/go-flagsfiller"
	"github.com/itzg/zapconfigs"
	packagesagent "github.com/racker/salus-packages-agent"
	"go.uber.org/zap"
	"os"
	"time"
)

// set by goreleaser build
var (
	version string
	commit  string
	date    string
)

var args struct {
	Debug   bool   `usage:"enables debug logging"`
	Version bool   `usage:"show version and exit" env:""`
	Configs string `usage:"directory containing config files that define continuous monitoring"`
	Include struct {
		Debian bool `default:"true" usage:"enables debian package listing, when not using configs"`
		Rpm    bool `default:"true" usage:"enables rpm package listing, when not using configs"`
	}
	LineProtocol struct {
		ToConsole bool   `usage:"indicates that line-protocol lines should be output to stdout"`
		ToSocket  string `usage:"the [host:port] of a telegraf TCP socket_listener"`
	}
}

func main() {
	err := flagsfiller.Parse(&args, flagsfiller.WithEnv("Agent"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if args.Version {
		fmt.Printf("%s (%s, %s)", version, commit, date)
		os.Exit(0)
	}

	var logger *zap.Logger
	if args.Debug {
		logger = zapconfigs.NewDebugLogger()
	} else {
		logger = zapconfigs.NewDefaultLogger()
	}
	defer logger.Sync()

	ctx := context.Background()

	var reporter packagesagent.PackagesReporter

	if args.LineProtocol.ToConsole {
		reporter = packagesagent.NewLineProtocolConsoleReporter(logger)
	} else if args.LineProtocol.ToSocket != "" {
		reporter, err = packagesagent.NewLineProtocolSocketReporter(ctx, args.LineProtocol.ToSocket, logger)
		if err != nil {
			logger.Fatal("failed to setup line-protocol socket reporter", zap.Error(err))
		}
	} else {
		// fallback to console reporter for humans
		reporter = packagesagent.NewConsoleReporter()
	}

	if args.Configs != "" {
		configs, err := packagesagent.LoadConfigs(args.Configs)
		if err != nil {
			logger.Fatal("failed to load configs", zap.Error(err))
		}

		packagesagent.CollectWithConfigs(ctx, configs, reporter, logger)

		// block and allow collector routines to run
		select {}
	} else {
		var listers []packagesagent.SoftwarePackageLister
		if args.Include.Debian {
			listers = append(listers, packagesagent.DebianLister(logger))
		}
		if args.Include.Rpm {
			listers = append(listers, packagesagent.RpmLister(logger))
		}

		err := packagesagent.CollectPackages(listers, reporter.StartBatch(time.Now()), false)
		if err != nil {
			logger.Fatal("failed to collect packages", zap.Error(err))
		}
	}
}
