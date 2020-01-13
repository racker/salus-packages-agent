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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

type mockPackageLister struct {
	mock.Mock
}

func (m *mockPackageLister) PackagingSystem() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockPackageLister) ListPackages() ([]SoftwarePackage, error) {
	args := m.Called()
	return args.Get(0).([]SoftwarePackage), args.Error(1)
}

type mockReporterBatch struct {
	mock.Mock
}

func (m *mockReporterBatch) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockReporterBatch) ReportSuccess(system string, packages []SoftwarePackage) {
	m.Called(system, packages)
}

func (m *mockReporterBatch) ReportFailure(system string, err error) {
	m.Called(system, err)
}

func TestCollectPackages_success(t *testing.T) {
	lister1 := &mockPackageLister{}
	lister1.On("PackagingSystem").Return("mock1")
	packages1 := []SoftwarePackage{
		{Name: "dbus", Version: "1:1.12.8-7.el8", Arch: "x86_64"},
	}
	lister1.On("ListPackages").Return(packages1, nil)
	lister2 := &mockPackageLister{}
	lister2.On("PackagingSystem").Return("mock2")
	packages2 := []SoftwarePackage{
		{Name: "dpkg", Version: "1.19.7", Arch: "amd64"},
	}
	lister2.On("ListPackages").Return(packages2, nil)

	batch := &mockReporterBatch{}
	batch.On("ReportSuccess", mock.Anything, mock.Anything)
	batch.On("Close").Return(nil)

	err := CollectPackages([]SoftwarePackageLister{lister1, lister2}, batch)
	require.NoError(t, err)

	lister1.AssertExpectations(t)
	lister2.AssertExpectations(t)

	batch.AssertCalled(t, "ReportSuccess", "mock1", packages1)
	batch.AssertCalled(t, "ReportSuccess", "mock2", packages2)
}
