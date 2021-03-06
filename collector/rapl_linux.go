//go:build !norapl
// +build !norapl

/*
 * @Author: your name
 * @Date: 2021-09-30 10:25:46
 * @LastEditTime: 2021-09-30 10:40:47
 * @LastEditors: your name
 * @Description: In User Settings Edit
 * @FilePath: /node_exporter/collector/rapl_linux.go
 */
// Copyright 2019 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs/sysfs"
)

type raplCollector struct {
	fs     sysfs.FS
	logger log.Logger
}

func init() {
	registerCollector("rapl", defaultEnabled, NewRaplCollector)
}

// NewRaplCollector returns a new Collector exposing RAPL metrics.
func NewRaplCollector(logger log.Logger) (Collector, error) {
	fs, err := sysfs.NewFS(*sysPath)

	if err != nil {
		return nil, err
	}

	collector := raplCollector{
		fs:     fs,
		logger: logger,
	}
	return &collector, nil
}

// Update implements Collector and exposes RAPL related metrics.
func (c *raplCollector) Update(ch chan<- prometheus.Metric) error {
	// nil zones are fine when platform doesn't have powercap files present.
	zones, err := sysfs.GetRaplZones(c.fs)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrNoData
		}
		if errors.Is(err, os.ErrPermission) {
			return ErrNoData
		}
		return fmt.Errorf("failed to retrieve rapl stats: %w", err)
	}

	for _, rz := range zones {
		newMicrojoules, err := rz.GetEnergyMicrojoules()
		if err != nil {
			return err
		}
		index := strconv.Itoa(rz.Index)

		descriptor := prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "rapl", rz.Name+"_joules_total"),
			"Current RAPL "+rz.Name+" value in joules",
			[]string{"index"}, nil,
		)

		ch <- prometheus.MustNewConstMetric(
			descriptor,
			prometheus.CounterValue,
			float64(newMicrojoules)/1000000.0,
			index,
		)
	}
	return nil
}
