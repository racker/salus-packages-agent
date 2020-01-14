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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfigs_success(t *testing.T) {
	configs, err := LoadConfigs(filepath.Join("testdata", "configs"))
	require.NoError(t, err)

	assert.Len(t, configs, 2)

	// since the file ordering is not deterministic the assertion logic is a little messy
	for i := 0; i < len(configs); i++ {
		if configs[i].IncludeRpm {
			assert.False(t, configs[i].IncludeDebian)
			// rpm file exercises the default interval scenario since one wasn't specified
			assert.Equal(t, DefaultInterval, configs[i].Interval)
		} else if configs[i].IncludeDebian {
			assert.False(t, configs[i].IncludeRpm)
			assert.Equal(t, Interval(6*time.Hour), configs[i].Interval)
			assert.True(t, configs[i].FailWhenNotSupported)
		} else {
			t.Fail()
		}
	}
}

func TestLoadConfigs_missingDir(t *testing.T) {
	_, err := LoadConfigs(filepath.Join("testdata", "not-configs"))
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to walk configs directory: open testdata/not-configs: no such file or directory")
}

func TestLoadConfigs_notADir(t *testing.T) {
	_, err := LoadConfigs(filepath.Join("testdata", "configs", "deb-monitor.json"))
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to walk configs directory: not a directory")
}

func TestLoadConfigs_badDecode(t *testing.T) {
	_, err := LoadConfigs(filepath.Join("testdata", "configs-malformed"))
	assert.Error(t, err)
	assert.EqualError(t, err, "failed to decode config file bad.json: invalid character 'N' looking for beginning of value")
}
