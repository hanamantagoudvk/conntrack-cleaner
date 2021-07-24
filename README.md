# conntrack-cleaner
The conntrack-cleaner cleans the `UNREPLIED` TCP conntrack entries. There are scenarios where Loadbalancer service is configured before endpoints are deployed. In such as case, service external-IP is allocated by a controller and this IP is known in advance even before the service is deployed. The service external IP is configured on the client machines located outside the Kubernetes cluster and they are trying to reach the service all the time. The adjacent routers are configured to forward the traffic matching the service external-IP towards the worker node statically.

In the beginning the clients are trying to reach to the service external-IP and at this time the Kubernetes cluster do not have the service or the pods up and running. At this time the iptables do not have entry for the service external-IP and this means that the packets will not be handled by the iptables populated by kube-proxy. These `TCP SYN` packets creates a conntrack entry which will get removed in 120 seconds if no further packets are matching to the tuple.

Say the client is having limited (may be 3) TCP source ports due to the way firewalling and clients are configured. This means that clients keep trying to establish TCP session with same source port and causes the `conntrack` entry to be alive and not cleaned up at all.

# Launching the agent as a DaemonSet

This repo includes an example yaml file that can be used to launch the conntrack-cleaner agent as a DaemonSet in a Kubernetes cluster.

    kubectl create -f conntrack-cleaner-agent.yaml

# Configuring the agent

There are two environment variables to be configured in deployment yaml file.

a) `CONNTRACK_TABLE_DUMP_FREQUENCY` signifies how frequently conntack-cleaner agent is reading/dumping the conntrack table from worker nodes.
This environment variable is of type `Duration`. Default value is `1second`. 
`NOTE: Dont set the frequency very high interms of micro/nano/milli seconds.`

b) `CONNECTION_RENEWAL_THRESHOLD` signifies the threshold value after which the agent deletes the conntrack table entry whose expiry timer was getting renewed continuously.
 Default value is 3.