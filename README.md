# Multi-elevator network project
[![Go Report Card](https://goreportcard.com/badge/github.com/sigtot/sanntid)](https://goreportcard.com/report/github.com/sigtot/sanntid)
[![GoDoc](https://godoc.org/github.com/sigtot/sanntid?status.svg)](https://godoc.org/github.com/sigtot/sanntid)

> Elevator Project for TTK4145 Real-time Programming

![simulator](/elev.gif)

## Design
The design for this system is based on the common publish/subscribe pattern which is heavily used distributed systems.
In hte figure below, the modules of the system are shown, with dashed, labeled arrows correspond to a module being
either a subscriber or a publisher to a certain topic.

Solid arrows however, signify more direct forms of communication between modules (channels, function calls, etc.).

![module_overview](https://i.imgur.com/q8aMH2N.png)

## Imported packages
### elevio package
The elevator driver used in the project was provided by the course instructors.
As we were not pleased with some of the implementation done in the driver, the repository was forked, 
and made some changes to. The original elevator driver can be found [here](https://github.com/TTK4145/driver-go).

### bbolt
Bolt is the database used by the order watchers to store not yet delivered calls.
The package used in this project is the bbolt package by etcd, which can be found [here](https://github.com/etcd-io/bbolt)

### logrus
Logging events in the system is done with the logrus package.
The package can be found [here](https://github.com/Sirupsen/logrus)

### Go standard library
A fair few of the packages from the Go standard library are used in this project.
For specific packages see the imports in the files. These packages are used for a lot of different tasks,
from io and os related, to encoding and compressing. For the network related parts of the project the net package 
and net/http package are used.
