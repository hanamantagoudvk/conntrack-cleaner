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
	"os/exec"

	"k8s.io/klog"
)

type connectionInfoStore struct {
	staleConnectionMarkCount int
	connEntry                connectionInfo
}

func deleteStaleConnEntry(sourceIP, destinationIP string) {
	_, err := exec.Command("conntrack", "-D", "-p", "tcp", "-s", sourceIP, "-d", destinationIP).CombinedOutput()
	if err != nil {
		klog.Errorf("error deleting conntrack entry : %s", err)
	}
	klog.V(4).Infof("conntrack entry deleted successfully for sourceIP: %s, destinationIP: %s", sourceIP, destinationIP)
}

func getKeyForConnInfo(connInfo connectionInfo) string {
	return connInfo.sourceIP + ":" + connInfo.sourcePort + ";" + connInfo.destinationIP + ":" + connInfo.destinationPort
}

func (c *conntrackCleaner) cleanStaleConntrackEntries(connInfo connectionInfo) {
	key := getKeyForConnInfo(connInfo)
	value, ok := c.connectionMap[key]
	if !ok {
		c.connectionMap[key] = connectionInfoStore{staleConnectionMarkCount: 0, connEntry: connInfo}
		return
	}
	//staleConnectionMarkCount is incremented if expiry time is equal or greater
	//than previous. Once staleConnectionMarkCount exceeds threshold, it
	//needs to be deleted.
	if connInfo.expiryTime >= value.connEntry.expiryTime {
		value.staleConnectionMarkCount++
		if value.staleConnectionMarkCount > c.connRenewalThreshold {
			deleteStaleConnEntry(connInfo.sourceIP, connInfo.destinationIP)
			delete(c.connectionMap, key)
		} else {
			c.connectionMap[key] = value
		}
	}
}

func (c *conntrackCleaner) runConnCleaner() {
	for {
		select {
		case connInfo := <-c.ciChannel:
			c.cleanStaleConntrackEntries(connInfo)
		}
	}
}
