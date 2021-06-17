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
	"bytes"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"k8s.io/klog"
)

var (
	sourceIPStr        = "src="
	destinationIPStr   = "dst="
	sourcePortStr      = "sport="
	destinationPortStr = "dport="
)

type conntrackCleaner struct {
	tableDumpFrequency   time.Duration
	connRenewalThreshold int
	ciChannel            chan connectionInfo
	connectionMap        map[string]connectionInfoStore
}

type connectionInfo struct {
	expiryTime      int
	sourceIP        string
	destinationIP   string
	sourcePort      string
	destinationPort string
}

func newConntrackCleaner(frequency time.Duration, threshold int) *conntrackCleaner {
	return &conntrackCleaner{
		tableDumpFrequency:   frequency,
		connRenewalThreshold: threshold,
		ciChannel:            make(chan connectionInfo),
		connectionMap:        make(map[string]connectionInfoStore),
	}
}

func extractConnInfo(parsedEntry []string) (*connectionInfo, error) {
	expTime, err := strconv.Atoi(parsedEntry[7])
	if err != nil {
		return nil, err
	}
	return &connectionInfo{
		expiryTime:      expTime,
		sourceIP:        strings.Split(parsedEntry[9], sourceIPStr)[1],
		destinationIP:   strings.Split(parsedEntry[10], destinationIPStr)[1],
		sourcePort:      strings.Split(parsedEntry[11], sourcePortStr)[1],
		destinationPort: strings.Split(parsedEntry[12], destinationPortStr)[1],
	}, nil
}

func parseConntrackEntry(entry string) []string {
	return strings.Split(entry, " ")
}

func parseConntrackTable(table string) []string {
	return strings.Split(table, "\n")
}

func (c *conntrackCleaner) processConntrackTable(table *bytes.Buffer) {
	entryList := parseConntrackTable(table.String())
	for _, entry := range entryList {
		if len(entry) != 0 {
			parsedEntry := parseConntrackEntry(entry)
			connInfo, err := extractConnInfo(parsedEntry)
			if err != nil {
				klog.Errorf("error extracting connection info : %v", err)
				continue
			}
			c.ciChannel <- *connInfo
		}
	}
}

func executeCmd(output *bytes.Buffer) error {
	var err error
	tcpConnList := exec.Command("conntrack", "-L", "-p", "tcp")
	grep := exec.Command("grep", "UNREPLIED")
	grep.Stdin, err = tcpConnList.StdoutPipe()
	if err != nil {
		return err
	}

	grep.Stdout = output
	// Start the grep command first. (The order will be last command first)
	grep.Start()
	tcpConnList.Run()
	grep.Wait()
	return nil
}

func (c *conntrackCleaner) runConntrackTableDump() {
	//Periodically take dump of conntrack table.
	for {
		func() {
			defer time.Sleep(c.tableDumpFrequency)
			var output bytes.Buffer
			err := executeCmd(&output)
			if err != nil {
				klog.Errorf("error executing conntrack cmd : %v", err)
				return
			}
			if output.Len() != 0 {
				c.processConntrackTable(&output)
			}
		}()
	}
}
