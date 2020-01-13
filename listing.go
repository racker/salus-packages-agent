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
	"bufio"
	"bytes"
	"fmt"
	"go.uber.org/zap"
	"os/exec"
	"strings"
)

type SoftwarePackage struct {
	Name    string
	Version string
	Arch    string
}

type SoftwarePackageLister interface {
	PackagingSystem() string
	IsSupported() bool
	ListPackages() ([]SoftwarePackage, error)
}

// commandBuilder matches the exec.Command signature and provides a mocking point
type commandBuilder func(commandName string, arg ...string) *exec.Cmd

// threeColumnPackageLister is able to generically use any package manager query command where
// the arguments declare an output format consisting of the space separated fields
// package name, version, and CPU architecture
type threeColumnPackageLister struct {
	packagingSystem string
	commandBuilder  commandBuilder
	commandName     string
	commandArgs     []string
	logger          *zap.Logger
}

func (t *threeColumnPackageLister) IsSupported() bool {
	_, err := exec.LookPath(t.commandName)
	if err != nil {
		return false
	} else {
		return true
	}
}

func (t *threeColumnPackageLister) PackagingSystem() string {
	return t.packagingSystem
}

func (t *threeColumnPackageLister) ListPackages() ([]SoftwarePackage, error) {
	cmd := t.commandBuilder(t.commandName, t.commandArgs...)
	t.logger.Debug("calling packaging tool",
		zap.String("name", t.commandName), zap.Strings("args", t.commandArgs))
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run package manager: %w", err)
	}

	buffer := bytes.NewBuffer(output)
	scanner := bufio.NewScanner(buffer)
	var pkgs []SoftwarePackage
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 3 {
			return nil, fmt.Errorf("package manager output line was malformed: %s", line)
		}
		pkgs = append(pkgs, SoftwarePackage{
			Name:    parts[0],
			Version: parts[1],
			Arch:    parts[2],
		})
	}

	return pkgs, nil
}

func RpmLister(logger *zap.Logger) SoftwarePackageLister {
	return &threeColumnPackageLister{
		packagingSystem: "rpm",
		commandBuilder:  exec.Command,
		commandName:     "rpm",
		commandArgs:     []string{"--query", "--all", "--queryformat", "'%{name} %{evr} %{arch}\\n'"},
		logger:          logger,
	}
}

func DebianLister(logger *zap.Logger) SoftwarePackageLister {
	return &threeColumnPackageLister{
		packagingSystem: "debian",
		commandBuilder:  exec.Command,
		commandName:     "dpkg-query",
		commandArgs:     []string{"--show", "--showformat", "'${Package} ${Version} ${Architecture}\\n'"},
		logger:          logger,
	}
}
