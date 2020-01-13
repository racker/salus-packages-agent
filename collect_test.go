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
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

type mockPackageLister struct {
	mock.Mock
}

func (m *mockPackageLister) IsSupported() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockPackageLister) PackagingSystem() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockPackageLister) ListPackages() ([]SoftwarePackage, error) {
	args := m.Called()
	pkgs := args.Get(0)
	// slices and especially possibly nil ones are weird to handle in testify's mocking
	if pkgs == nil {
		return nil, args.Error(1)
	} else {
		return pkgs.([]SoftwarePackage), args.Error(1)
	}
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
	lister1.On("IsSupported").Return(true)
	packages1 := []SoftwarePackage{
		{Name: "dbus", Version: "1:1.12.8-7.el8", Arch: "x86_64"},
	}
	lister1.On("ListPackages").Return(packages1, nil)

	lister2 := &mockPackageLister{}
	lister2.On("PackagingSystem").Return("mock2")
	lister2.On("IsSupported").Return(true)
	packages2 := []SoftwarePackage{
		{Name: "dpkg", Version: "1.19.7", Arch: "amd64"},
	}
	lister2.On("ListPackages").Return(packages2, nil)

	batch := &mockReporterBatch{}
	batch.On("ReportSuccess", mock.Anything, mock.Anything)
	batch.On("Close").Return(nil)

	err := CollectPackages([]SoftwarePackageLister{lister1, lister2}, batch, false)
	require.NoError(t, err)

	lister1.AssertExpectations(t)
	lister2.AssertExpectations(t)

	batch.AssertCalled(t, "ReportSuccess", "mock1", packages1)
	batch.AssertCalled(t, "ReportSuccess", "mock2", packages2)
}

func TestCollectPackages_handleErrorInOne(t *testing.T) {
	lister1 := &mockPackageLister{}
	lister1.On("PackagingSystem").Return("mock1")
	lister1.On("IsSupported").Return(true)
	packages1 := []SoftwarePackage{
		{Name: "dbus", Version: "1:1.12.8-7.el8", Arch: "x86_64"},
	}
	lister1.On("ListPackages").Return(packages1, nil)

	// lister2 is going to simulate an error during invoking the package manager
	lister2 := &mockPackageLister{}
	lister2.On("PackagingSystem").Return("mock2")
	lister2.On("IsSupported").Return(true)
	lister2.On("ListPackages").Return(nil,
		errors.New("something went wrong"))

	batch := &mockReporterBatch{}
	batch.On("ReportSuccess", mock.Anything, mock.Anything)
	batch.On("ReportFailure", mock.Anything, mock.Anything)
	batch.On("Close").Return(nil)

	err := CollectPackages([]SoftwarePackageLister{lister1, lister2}, batch, false)
	// this one propagates the error since the package manager was supposed to be supported here
	assert.Error(t, err)

	lister1.AssertExpectations(t)
	lister2.AssertExpectations(t)

	batch.AssertCalled(t, "ReportSuccess", "mock1", packages1)
	batch.AssertCalled(t, "ReportFailure", "mock2",
		mock.MatchedBy(func(err error) bool {
			return err.Error() == "something went wrong"
		}),
	)
}

func TestCollectPackages_reportNotSupported(t *testing.T) {
	lister1 := &mockPackageLister{}
	lister1.On("PackagingSystem").Return("mock1")
	lister1.On("IsSupported").Return(false)

	batch := &mockReporterBatch{}
	batch.On("ReportFailure", mock.Anything, mock.Anything)
	batch.On("Close").Return(nil)

	err := CollectPackages([]SoftwarePackageLister{lister1}, batch, true)
	// outer call itself purposely reports no error, but reporter batch, below, will get it
	require.NoError(t, err)

	lister1.AssertExpectations(t)

	batch.AssertCalled(t, "ReportFailure",
		"mock1",
		mock.MatchedBy(func(err error) bool {
			return err.Error() == "package system mock1 is not supported"
		}),
	)
}
