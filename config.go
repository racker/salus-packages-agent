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
	"encoding/json"
	"fmt"
	"github.com/karrick/godirwalk"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultInterval = Interval(1 * time.Hour)
)

// Go doesn't have builtin JSON unmarshalling of time.Duration, so declare our own type and unmarshaller
type Interval time.Duration

// UnmarshalJSON parses string formatted duration/interval values
func (i *Interval) UnmarshalJSON(b []byte) error {
	var value string
	err := json.Unmarshal(b, &value)
	if err != nil {
		return err
	}
	dur, err := time.ParseDuration(value)
	if err != nil {
		return err
	}
	*i = Interval(dur)
	return nil
}

type Config struct {
	Interval             Interval `json:"interval"`
	IncludeRpm           bool     `json:"include-rpm"`
	IncludeDebian        bool     `json:"include-debian"`
	FailWhenNotSupported bool     `json:"fail-when-not-supported"`
}

func LoadConfigs(configsDir string) ([]*Config, error) {
	configs := make([]*Config, 0)

	scanner, err := godirwalk.NewScanner(configsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to walk configs directory: %w", err)
	}

	for scanner.Scan() {
		dirent, err := scanner.Dirent()
		if err != nil {
			return nil, fmt.Errorf("failed to read config directory entry: %w", err)
		}

		name := dirent.Name()
		if dirent.IsRegular() && strings.HasSuffix(name, ".json") {
			config, err := loadConfigFile(configsDir, name)
			if err != nil {
				return nil, err
			}
			configs = append(configs, config)
		}
	}
	if scanner.Err() != nil {
		return nil, fmt.Errorf("failed to walk configs directory: %w", scanner.Err())
	}

	return configs, nil
}

func loadConfigFile(configsDir string, name string) (*Config, error) {
	file, err := os.Open(filepath.Join(configsDir, name))
	if err != nil {
		return nil, fmt.Errorf("failed to open config file %s: %w", name, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var config Config
	err = decoder.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config file %s: %w", name, err)
	}

	if config.Interval == 0 {
		config.Interval = DefaultInterval
	}

	return &config, nil
}
