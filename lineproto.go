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
	"bytes"
	protocol "github.com/influxdata/line-protocol"
	lpsender "github.com/itzg/line-protocol-sender"
	"go.uber.org/zap"
	"io"
	"os"
	"time"
)

const (
	LpMeasurementName        = "packages"
	LpMeasurementFailureName = "packages_failed"
	LpSystemTag              = "system"
	LpPackageTag             = "package"
	LpArchTag                = "arch"
	LpVersionField           = "version"
	LpErrorField             = "error"

	// follow the pattern of telegraf's --test option and use their same prefix
	LpOutPrefix = "> "
)

func NewLineProtocolConsoleReporter(logger *zap.Logger) PackagesReporter {
	return &lineProtocolConsoleReporter{out: os.Stdout, logger: logger}
}

type lineProtocolConsoleReporter struct {
	out    io.Writer
	logger *zap.Logger
}

func (l *lineProtocolConsoleReporter) StartBatch(timestamp time.Time) PackagesReporterBatch {
	return &lineProtocolConsoleBatch{timestamp: timestamp, out: l.out, logger: l.logger}
}

type lineProtocolConsoleBatch struct {
	timestamp time.Time
	logger    *zap.Logger
	out       io.Writer
}

func (l *lineProtocolConsoleBatch) Close() error {
	// nothing needed
	return nil
}

func (l *lineProtocolConsoleBatch) ReportSuccess(system string, packages []SoftwarePackage) {
	metrics := buildLineProtocolMetrics(l.timestamp, system, packages)

	var buf bytes.Buffer

	for _, metric := range metrics {
		buf.Reset()
		l.writeMetric(&buf, metric)
	}
}

func (l *lineProtocolConsoleBatch) writeMetric(buf *bytes.Buffer, metric *lpsender.SimpleMetric) {
	buf.WriteString(LpOutPrefix)
	encoder := protocol.NewEncoder(buf)
	_, err := encoder.Encode(metric)
	if err != nil {
		l.logger.Error("failed to encode metric", zap.Error(err), zap.Any("metric", metric))
		return
	}
	// encode appends newline, so write as-is
	_, err = l.out.Write(buf.Bytes())
	if err != nil {
		l.logger.Error("failed to write out metric", zap.Error(err))
	}
}

func (l *lineProtocolConsoleBatch) ReportFailure(system string, err error) {
	metric := lpsender.NewSimpleMetric(LpMeasurementFailureName)
	metric.SetTime(l.timestamp)
	metric.AddTag(LpSystemTag, system)
	metric.AddField(LpErrorField, err.Error())

	var buf bytes.Buffer
	l.writeMetric(&buf, metric)
}

func buildLineProtocolMetrics(timestamp time.Time, system string, packages []SoftwarePackage) []*lpsender.SimpleMetric {
	metrics := make([]*lpsender.SimpleMetric, 0, len(packages))

	for _, pkg := range packages {
		metric := lpsender.NewSimpleMetric(LpMeasurementName)
		metric.SetTime(timestamp)
		metric.AddTag(LpSystemTag, system)
		metric.AddTag(LpPackageTag, pkg.Name)
		metric.AddTag(LpArchTag, pkg.Arch)
		metric.AddField(LpVersionField, pkg.Version)

		metrics = append(metrics, metric)
	}

	return metrics
}
