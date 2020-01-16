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
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// mockCommandBuilder works in conjunction with TestMockCommandHelper to mock the package manager
// executables in the same way that Go's exec package does their unit testing.
func mockCommandBuilder(commandName string, arg ...string) *exec.Cmd {
	cs := []string{"-test.run=TestMockCommandHelper", "--"}
	// add the original bits
	cs = append(cs, commandName)
	cs = append(cs, arg...)

	// re-exec the current executable whichi was built by go test
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")

	return cmd
}

// TestMockCommandHelper isn't a real test, but instead follows the pattern of the exec package's
// unit testing with TestHelperProcess.
// This article provides more background about the technique
// https://npf.io/2015/06/testing-exec-command/
func TestMockCommandHelper(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		// skip during top level unit test execution
		return
	}

	// locate the args of mocked package manager command
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}

	file, err := os.Open(filepath.Join("testdata", fmt.Sprintf("%s.out", args[0])))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	os.Exit(0)
}

func TestThreeColumnPackageLister_ListPackages(t *testing.T) {
	lister := threeColumnPackageLister{
		commandBuilder: mockCommandBuilder,
		commandName:    "rpm",
		logger:         zap.NewNop(),
	}

	packages, err := lister.ListPackages()
	require.NoError(t, err)
	assert.Len(t, packages, 174)
	// and spot check one
	assert.Equal(t, "tzdata", packages[0].Name)
	assert.Equal(t, "2019a-1.el8", packages[0].Version)
	assert.Equal(t, "noarch", packages[0].Arch)
}

func TestThreeColumnPackageLister_ListPackages_errorFromCommand(t *testing.T) {
	lister := threeColumnPackageLister{
		commandBuilder: mockCommandBuilder,
		commandName:    "does_not_exist",
		logger:         zap.NewNop(),
	}

	_, err := lister.ListPackages()
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to run package manager: exit status 1")
}

func TestThreeColumnPackageLister_ListPackages_malformed(t *testing.T) {
	lister := threeColumnPackageLister{
		commandBuilder: mockCommandBuilder,
		commandName:    "malformed",
		logger:         zap.NewNop(),
	}

	_, err := lister.ListPackages()
	assert.Error(t, err)
	assert.EqualError(t, err, "package manager output line was malformed: tzdata 2019a-1.el8")
}

func TestDebianLister(t *testing.T) {
	logger := zap.NewNop()
	lister := DebianLister(logger)
	require.IsType(t, &threeColumnPackageLister{}, lister)
	casted := lister.(*threeColumnPackageLister)

	assert.Equal(t, "dpkg-query", casted.commandName)
	assert.NotEmpty(t, casted.commandArgs)
}

func TestRpmLister(t *testing.T) {
	logger := zap.NewNop()
	lister := RpmLister(logger)
	require.IsType(t, &threeColumnPackageLister{}, lister)
	casted := lister.(*threeColumnPackageLister)

	assert.Equal(t, "rpm", casted.commandName)
	assert.NotEmpty(t, casted.commandArgs)
}
