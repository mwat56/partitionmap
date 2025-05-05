# PartitionMap

[![golang](https://img.shields.io/badge/Language-Go-green.svg)](https://golang.org/)
[![GoDoc](https://godoc.org/github.com/mwat56/partitionmap?status.svg)](https://godoc.org/github.com/mwat56/partitionmap)
[![Go Report](https://goreportcard.com/badge/github.com/mwat56/partitionmap)](https://goreportcard.com/report/github.com/mwat56/partitionmap)
[![Issues](https://img.shields.io/github/issues/mwat56/partitionmap.svg)](https://github.com/mwat56/partitionmap/issues?q=is%3Aopen+is%3Aissue)
[![Size](https://img.shields.io/github/repo-size/mwat56/partitionmap.svg)](https://github.com/mwat56/partitionmap/)
[![Tag](https://img.shields.io/github/tag/mwat56/partitionmap.svg)](https://github.com/mwat56/partitionmap/tags)
[![View examples](https://img.shields.io/badge/learn%20by-examples-0077b3.svg)](https://github.com/mwat56/partitionmap/blob/main/_demo/demo.go)
[![License](https://img.shields.io/github/mwat56/partitionmap.svg)](https://github.com/mwat56/partitionmap/blob/main/LICENSE)

- [PartitionMap](#partitionmap)
	- [Purpose](#purpose)
	- [Installation](#installation)
	- [Usage](#usage)
	- [Libraries](#libraries)
	- [Licence](#licence)

----

## Purpose

This is a `Go` package that implements a _partitioned map data structure_ for storing key-value pairs. Key features include:

- Using a partitioning strategy with CRC32 hashing to distribute keys across multiple partitions,
- thread-safe implementation with mutex locks for concurrent access,
- generic implementation supporting any value type,
- designed for efficient concurrent access by reducing lock contention,
- implements standard map operations (`Get`, `Put`, `Delete`, etc.).

The package is designed to provide better performance than a standard map when used in highly concurrent environments by sharding the data across multiple partitions.

## Installation

You can use `Go` to install this package for you:

    go get -u github.com/mwat56/partitionmap

## Usage

    //TODO

## Libraries

The following external libraries were used building `partitionmap`:

* _No external libraries are used by this module._

## Licence

        Copyright © 2024, 2025  M.Watermann, 10247 Berlin, Germany
                        All rights reserved
                    EMail : <support@mwat.de>

> This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.
>
> This software is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
>
> You should have received a copy of the GNU General Public License along with this program. If not, see the [GNU General Public License](http://www.gnu.org/licenses/gpl.html) for details.

----
[![GFDL](https://www.gnu.org/graphics/gfdl-logo-tiny.png)](http://www.gnu.org/copyleft/fdl.html)
