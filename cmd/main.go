/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"os"
	"strconv"
	"time"

	"k8s.io/klog"
)

func main() {
	cleaner := newConntrackCleaner(getConntrackDumpFrequency(), getThreshold())
	go cleaner.runConntrackTableDump()
	cleaner.runConnCleaner()
}

func getConntrackDumpFrequency() time.Duration {
	defaultDumpFrequency := time.Duration(1) * time.Second
	frequency, ok := os.LookupEnv("CONNTRACK_TABLE_DUMP_FREQUENCY")
	if !ok {
		klog.Warning("CONNTRACK_TABLE_DUMP_FREQUENCY env variable not set in podspec. Taking default value as 1sec")
		return defaultDumpFrequency
	}
	configuredDumpFrequency, err := time.ParseDuration(frequency)
	if err != nil {
		klog.Warning("invalid value given for CONNTRACK_TABLE_DUMP_FREQUENCY in podspec. Taking default value as 1sec")
		return defaultDumpFrequency
	}
	return configuredDumpFrequency
}

func getThreshold() int {
	defaultThreshold := 3
	threshold, ok := os.LookupEnv("CONNECTION_RENEWAL_THRESHOLD")
	if !ok {
		klog.Warning("CONNECTION_RENEWAL_THRESHOLD env variable not set in podspec. Taking default value as 3")
		return defaultThreshold
	}
	configuredThreshold, err := strconv.Atoi(threshold)
	if err != nil {
		klog.Warning("invalid value given for CONNECTION_RENEWAL_THRESHOLD in podspec. Taking default value as 3")
		return defaultThreshold
	}
	return configuredThreshold
}
