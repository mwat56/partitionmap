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
		- [Basic Usage](#basic-usage)
		- [Key Features](#key-features)
		- [Performance Considerations](#performance-considerations)
	- [Libraries](#libraries)
	- [Licence](#licence)

----

## Purpose

This is a `Go` package that implements a _partitioned data structure_ for storing key/value pairs. Key features include:

- Using a partitioning strategy with CRC32 hashing to distribute keys across multiple partitions,
- thread-safe implementation with mutex locks for concurrent data access,
- generic implementation supporting any value type,
- designed for efficient concurrent access by reducing lock contention,
- implements standard map operations (`Get`, `Put`, `Delete`, etc.).

The package is designed to provide better performance than a standard map when used in highly concurrent environments by sharding the data across multiple partitions.

## Installation

You can use `Go` to install this package for you:

	go get -u github.com/mwat56/partitionmap

## Usage

### Basic Usage

	package main

	import (
		"fmt"
		"github.com/mwat56/partitionmap"
	)

	func main() {
		// Create a new partition map for string values
		pm := partitionmap.New[string]()

		// Store key-value pairs
		pm.Put("user1", "John Doe")
		pm.Put("user2", "Jane Smith")

		// Retrieve values
		value, exists := pm.Get("user1")
		if exists {
			fmt.Printf("Found user1: %s\n", value)
		}

		// Delete a key
		pm.Delete("user2")

		// Check map size
		fmt.Printf("Map contains %d entries\n", pm.Len())

		// Iterate over all entries
		pm.ForEach(func(aKey string, aValue string) {
			fmt.Printf("%s: %s\n", aKey, aValue)
		})

		// Get all keys
		keys := pm.Keys()
		fmt.Printf("Keys: %v\n", keys)

		// Clear all entries
		pm.Clear()
	} // main()

### Key Features

1. Generic Implementation: Works with any value type using Go generics.

		intMap := partitionmap.New[int]()
		structMap := partitionmap.New[MyStruct]()

2. Thread-Safety: All operations are thread-safe, making it suitable for concurrent access.

		// Can be safely used from multiple goroutines
		go func() { pm.Put("key1", "value1") }()
		go func() { pm.Get("key2") }()

3. Method Chaining: Most methods return the map itself for chaining operations.

		pm.Put("key1", "value1").Put("key2", "value2").Delete("key3")

4. Efficient Partitioning: Uses CRC32 hashing to distribute keys across 64 partitions, reducing lock contention.

5. Lazy Partition Creation: Partitions are created lazily when needed, saving memory in a sparse map.

### Performance Considerations

- The map uses partitioning to reduce lock contention in concurrent scenarios.
- Partitions are created lazily when needed.
- For high-concurrency applications, this implementation can offer better performance than a standard map with a global lock.

The package is ideal for scenarios where you need a thread-safe map with good performance under concurrent access patterns.

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
