// Copyright 2016 The kingshard Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package config

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

//用于通过api保存配置
var configFileName string

// Config 整个config文件对应的结构
type Config struct {
	Addr     string `yaml:"addr"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`

	WebAddr     string `yaml:"web_addr"`
	WebUser     string `yaml:"web_user"`
	WebPassword string `yaml:"web_password"`

	LogFilePrefix  string       `yaml:"log_file_prefix"`
	LogPath        string       `yaml:"log_path"`
	LogLevel       string       `yaml:"log_level"`
	LogSql         bool         `yaml:"log_sql"` // LogSql      string       `yaml:"log_sql"`
	SlowLogTime    int          `yaml:"slow_log_time"`
	AllowIps       string       `yaml:"allow_ips"`
	BlsFile        string       `yaml:"blacklist_sql_file"`
	Charset        string       `yaml:"proxy_charset"`
	MonitorAddress string       `yaml:"monitor_addr"`
	MonitorPort    string       `yaml:"monitor_port"`
	Nodes          []NodeConfig `yaml:"nodes"`

	Schema SchemaConfig `yaml:"schema"`
}

// NodeConfig node节点对应的配置
type NodeConfig struct {
	Name             string `yaml:"name"`
	DownAfterNoAlive int    `yaml:"down_after_noalive"`
	MaxConnNum       int    `yaml:"max_conns_limit"`

	User     string `yaml:"user"`
	Password string `yaml:"password"`

	Master string `yaml:"master"`
	Slave  string `yaml:"slave"`
}

// SchemaConfig schema对应的结构体
type SchemaConfig struct {
	Nodes     []string      `yaml:"nodes"`
	Default   string        `yaml:"default"` //default node
	ShardRule []ShardConfig `yaml:"shard"`   //route rule
}

// ShardConfig range,hash or date
type ShardConfig struct {
	DB            string   `yaml:"db"`
	Table         string   `yaml:"table"`
	Key           string   `yaml:"key"`
	Nodes         []string `yaml:"nodes"`
	Locations     []int    `yaml:"locations"`
	Type          string   `yaml:"type"`
	TableRowLimit int      `yaml:"table_row_limit"`
	DateRange     []string `yaml:"date_range"`
}

// ParseConfigData ...
func ParseConfigData(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if err := checkValueValid(&cfg); err != nil {
		return nil, err
	}

	// fmt.Println(cfg)
	return &cfg, nil
}

// ParseConfigFile ....
func ParseConfigFile(fileName string) (*Config, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	configFileName = fileName

	return ParseConfigData(data)
}

func checkValueValid(cfg *Config) error {
	var ports []string
	if cfg.MonitorPort == "" {
		cfg.MonitorPort = "19827"
	}
	ports = append(ports, cfg.MonitorPort)
	for _, v := range cfg.Nodes {
		addrandport := strings.Split(v.Master, ":")
		if len(addrandport) != 2 {
			return fmt.Errorf("master[%s] format is err", v.Name)
		}
		ports = append(ports, addrandport[1])

		slaves := strings.Split(v.Slave, ",")
		for _, slave := range slaves {
			if slave == "" {
				continue
			}
			sl := strings.Split(slave, "@")
			addrandport := strings.Split(sl[0], ":")
			if len(addrandport) != 2 {
				return fmt.Errorf("Node [%s] slave [%s] format is err", v.Name, slave)
			}
			ports = append(ports, addrandport[1])
		}
	}
	ret := isPort(ports)
	if ret != nil {
		return ret
	}

	return nil
}

func isPort(ports []string) error {
	for _, port := range ports {
		p, ok := strconv.Atoi(port)
		if ok != nil || p > 65535 || p <= 0 {
			return fmt.Errorf("port[%s] is not availd", port)
		}
	}
	return nil
}

// WriteConfigFile Dump cfg to File(configFileName)
func WriteConfigFile(cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(configFileName, data, 0755)
	if err != nil {
		return err
	}

	return nil
}
