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

package packagesagent

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"io"
	"time"
)

var (
	initialCollectionDelay = 1 * time.Second
)

type PackagesReporter interface {
	StartBatch(timestamp time.Time) PackagesReporterBatch
}

type PackagesReporterBatch interface {
	io.Closer
	ReportSuccess(system string, packages []SoftwarePackage)
	ReportFailure(system string, err error)
}

func CollectPackages(listers []SoftwarePackageLister, reporterBatch PackagesReporterBatch, reportWhenNotSupported bool) error {
	defer reporterBatch.Close()

	for _, lister := range listers {
		system := lister.PackagingSystem()

		if !lister.IsSupported() {
			if reportWhenNotSupported {
				reporterBatch.ReportFailure(system, fmt.Errorf("package system %s is not supported", system))
			}
			continue
		}

		packages, err := lister.ListPackages()
		if err != nil {
			reporterBatch.ReportFailure(system, err)
			return fmt.Errorf("failed to collect %s packages: %w", system, err)
		} else {
			reporterBatch.ReportSuccess(system, packages)
		}
	}

	return nil
}

// CollectWithConfigs will start a go routine each to periodically collect packages according
// to each given configuration.
func CollectWithConfigs(ctx context.Context, configs []*Config, reporter PackagesReporter, logger *zap.Logger) {
	for _, config := range configs {
		go collectWithConfig(ctx, config, reporter, logger)
	}
}

// listersFromConfig is a var to allow for unit test replacement with mocks
var listersFromConfig = func(config *Config, logger *zap.Logger) []SoftwarePackageLister {
	var listers []SoftwarePackageLister
	if config.IncludeDebian {
		listers = append(listers, DebianLister(logger))
	}
	if config.IncludeRpm {
		listers = append(listers, RpmLister(logger))
	}
	return listers
}

func collectWithConfig(ctx context.Context, config *Config, reporter PackagesReporter, logger *zap.Logger) {
	initialDelayChan := time.After(initialCollectionDelay)
	ticker := time.NewTicker(time.Duration(config.Interval))

	listers := listersFromConfig(config, logger)

	handleTick := func(timestamp time.Time) {
		batch := reporter.StartBatch(timestamp)
		err := CollectPackages(listers, batch, config.FailWhenNotSupported)
		if err != nil {
			logger.Error("failed to collect packages", zap.Error(err))
		}
		err = batch.Close()
		if err != nil {
			logger.Error("failed to close reporter batch", zap.Error(err))
		}
	}

	for {
		select {
		case timestamp := <-initialDelayChan:
			handleTick(timestamp)
		case timestamp := <-ticker.C:
			handleTick(timestamp)
		case <-ctx.Done():
			return
		}
	}
}

type consoleReporter struct {
}

func NewConsoleReporter() PackagesReporter {
	return &consoleReporter{}
}

type consoleReporterBatch struct{}

func (c *consoleReporter) StartBatch(timestamp time.Time) PackagesReporterBatch {
	fmt.Printf("============================================================\n")
	fmt.Printf("%s\n", timestamp.Format(time.RFC822))

	return &consoleReporterBatch{}
}

func (c *consoleReporterBatch) Close() error {
	fmt.Printf("============================================================\n")
	return nil
}

func (c *consoleReporterBatch) ReportSuccess(system string, packages []SoftwarePackage) {
	fmt.Printf("-- %s --------------------------------------------------\n", system)
	for _, p := range packages {
		fmt.Printf("%-20s %-25s %s\n", p.Name, p.Version, p.Arch)
	}
}

func (c *consoleReporterBatch) ReportFailure(system string, err error) {
	// outer caller will log this
}
