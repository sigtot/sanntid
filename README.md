# Multi-elevator network project
[![Go Report Card](https://goreportcard.com/badge/github.com/sigtot/sanntid)](https://goreportcard.com/report/github.com/sigtot/sanntid)

> Elevator Project for TTK4145 Real-time Programming

## Design
### Major Challanges
The elevator nodes might suffer from power loss at any time. Furthermore, the nodes communicate over unreliable network, such that some nodes may be unavailable to the rest of the network.

#### Distributed Consensus
We want the elevators to agree on which one of them should serve a given order. This boils down to obtaining distributed consensus between the nodes.

A major method for obtaining distributed consensus is [PAXOS](https://en.wikipedia.org/wiki/Paxos_(computer_science)). 

However, a modern, simpler algorithm, which is employed in (etcd)[https://github.com/etcd-io/etcd] among other places will probably be easier to use. This algorithm is called (raft)[https://raft.github.io/]. It works by holding _elections_ and electing a _leader_, which then decides what to write to the other nodes. The leader must send out consecutive _heartbeat_ signals to notify the nodes that it's alive. If one of the nodes doesn't hear from the leader in a long time, it will trigger a reelection and likely take the leader's place. 

More about [state machine replication](https://en.wikipedia.org/wiki/State_machine_replication)

#### Storage and power loss
Storing data to disk can be problematic in the case of a power outage. In such a case, data might only be partially written and corrupted. 

The home directories on the lab computers use [ext3](https://en.wikipedia.org/wiki/Ext3). This is a [journaled](https://en.wikipedia.org/wiki/Journaling_file_system) filesystem, which means storage operations can be considered atomic. A journaled filesystem accomplishes atomicity by writing any changes to a journal before applying them. This way, if a power loss occurs during an operation, the recovery process can check the journal for the changes and replay any failed ones. This process is known as [write-ahead logging (WAL)](https://en.wikipedia.org/wiki/Write-ahead_logging).

* Etcd has a WAL implementation: [github.com/etcd-io/etcd/tree/master/wal](https://github.com/etcd-io/etcd/tree/master/wal)
* [Stackoverflow: Losing Power While Writing to a File](https://stackoverflow.com/questions/16835529/losing-power-while-writing-to-a-file)
* [The ARIES wikipedia page](https://en.wikipedia.org/wiki/Algorithms_for_Recovery_and_Isolation_Exploiting_Semantics) has a section on checkpoints in journal logs

### Raft
#### Network partitions
During a network partition, the network gets separated into two parts. In this event, a new leader is elected in the majority network. In the minority partition, no state changes are accepted, since a majority is not in place. The minority partition must therefore either a) be put out of service or b) continue operating in a safe mode, which does not rely on the network. We can for instance have a mode in which the node that receives the order always delivers it too. When a safe mode state is finished, we must wait until all it's operations have been completed, as these will completely disappear from the actual distributed state log.

Visual go-through of the raft algorithm: [thesecretlivesofdata.com/raft](http://thesecretlivesofdata.com/raft/)
