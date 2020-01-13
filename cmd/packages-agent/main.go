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
	"fmt"
	"github.com/itzg/go-flagsfiller"
	"github.com/itzg/zapconfigs"
	packagesagent "github.com/racker/salus-packages-agent"
	"go.uber.org/zap"
	"os"
	"time"
)

var config struct {
	Interval time.Duration `default:"1h"`
	OneShot  bool
	Debug    bool
	Include  struct {
		Debian bool `default:"true"`
		Rpm    bool `default:"true"`
	}
}

func main() {
	err := flagsfiller.Parse(&config, flagsfiller.WithEnv(""))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	var logger *zap.Logger
	if config.Debug {
		logger = zapconfigs.NewDebugLogger()
	} else {
		logger = zapconfigs.NewDefaultLogger()
	}

	var listers []packagesagent.SoftwarePackageLister

	if config.Include.Debian {
		listers = append(listers, packagesagent.DebianLister(logger))
	}
	if config.Include.Rpm {
		listers = append(listers, packagesagent.RpmLister(logger))
	}

	reporter := packagesagent.NewConsoleReporter()

	if config.OneShot {
		err := packagesagent.CollectPackages(listers, reporter.StartBatch(time.Now()), false)
		if err != nil {
			logger.Error("failed to collect packages", zap.Error(err))
			os.Exit(1)
		}
	} else {
		ticker := time.Tick(config.Interval)
		for timestamp := range ticker {
			err := packagesagent.CollectPackages(listers, reporter.StartBatch(timestamp), false)
			if err != nil {
				logger.Error("failed to collect packages", zap.Error(err))
			}
		}
	}
}
