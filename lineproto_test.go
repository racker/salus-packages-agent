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
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"net"
	"testing"
	"time"
)

func TestLineProtocolConsoleBatch_ReportSuccess(t *testing.T) {
	timestamp, err := time.ParseInLocation(time.RFC3339, "2006-01-02T15:04:05Z", time.UTC)
	require.NoError(t, err)

	var out bytes.Buffer
	reporter := &lineProtocolConsoleReporter{out: &out, logger: zap.NewNop()}
	batch := reporter.StartBatch(timestamp)
	require.NotNil(t, batch)

	batch.ReportSuccess("rpm", []SoftwarePackage{
		{Name: "tzdata", Version: "2019a-1.el8", Arch: "noarch"},
		{Name: "libselinux", Version: "2.8-6.el8", Arch: "x86_64"},
	})

	assert.Equal(t, `> packages,system=rpm,package=tzdata,arch=noarch version="2019a-1.el8" 1136214245000000000
> packages,system=rpm,package=libselinux,arch=x86_64 version="2.8-6.el8" 1136214245000000000
`, out.String())
}

func TestLineProtocolConsoleBatch_ReportFailure(t *testing.T) {
	timestamp, err := time.ParseInLocation(time.RFC3339, "2006-01-02T15:04:05Z", time.UTC)
	require.NoError(t, err)

	var out bytes.Buffer
	reporter := &lineProtocolConsoleReporter{out: &out, logger: zap.NewNop()}
	batch := reporter.StartBatch(timestamp)
	require.NotNil(t, batch)

	batch.ReportFailure("rpm", errors.New("this is just a test"))

	assert.Equal(t, `> packages_failed,system=rpm error="this is just a test" 1136214245000000000
`, out.String())
}

func TestLineProtocolSocketBatch_ReportSuccess(t *testing.T) {
	timestamp, err := time.ParseInLocation(time.RFC3339, "2006-01-02T15:04:05Z", time.UTC)
	require.NoError(t, err)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	// our mock socket_listener will post the received lines here
	consumedLines := make(chan string, 1)

	listener, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	defer listener.Close()

	go mockLineProtocolListener(t, listener, consumedLines)

	reporter, err := NewLineProtocolSocketReporter(ctx, listener.Addr().String(), zap.NewNop())
	require.NoError(t, err)

	batch := reporter.StartBatch(timestamp)

	batch.ReportSuccess("rpm", []SoftwarePackage{
		{Name: "tzdata", Version: "2019a-1.el8", Arch: "noarch"},
		{Name: "libselinux", Version: "2.8-6.el8", Arch: "x86_64"},
	})

	err = batch.Close()
	require.NoError(t, err)

	assertLineOrTimeout(t, consumedLines, `packages,system=rpm,package=tzdata,arch=noarch version="2019a-1.el8" 1136214245000000000`)
	assertLineOrTimeout(t, consumedLines, `packages,system=rpm,package=libselinux,arch=x86_64 version="2.8-6.el8" 1136214245000000000`)
}

func TestLineProtocolSocketBatch_ReportFailure(t *testing.T) {
	timestamp, err := time.ParseInLocation(time.RFC3339, "2006-01-02T15:04:05Z", time.UTC)
	require.NoError(t, err)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	// our mock socket_listener will post the received lines here
	consumedLines := make(chan string, 1)

	listener, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	defer listener.Close()

	go mockLineProtocolListener(t, listener, consumedLines)

	reporter, err := NewLineProtocolSocketReporter(ctx, listener.Addr().String(), zap.NewNop())
	require.NoError(t, err)

	batch := reporter.StartBatch(timestamp)

	batch.ReportFailure("rpm", errors.New("this is just a test"))

	err = batch.Close()
	require.NoError(t, err)

	assertLineOrTimeout(t, consumedLines, `packages_failed,system=rpm error="this is just a test" 1136214245000000000`)
}

func assertLineOrTimeout(t *testing.T, lines <-chan string, expected string) {
	select {
	case line := <-lines:
		assert.Equal(t, expected, line)
	case <-time.After(50 * time.Millisecond):
		t.Errorf("timeout waiting for line: %s", expected)
	}
}

func mockLineProtocolListener(t *testing.T, listener net.Listener, lines chan<- string) {
	for {
		conn, err := listener.Accept()
		t.Log("accepted connection")
		if err != nil {
			// unit test closed our socket, so we're done
			return
		}

		go mockLineProtocolConsumer(t, conn, lines)
	}
}

func mockLineProtocolConsumer(t *testing.T, conn net.Conn, lines chan<- string) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		t.Log("read line", line)
		lines <- line
	}
}
