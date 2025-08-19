// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

package harness

// CollectorService defines the interface for a collector service
type CollectorService interface {
	Start() error
	Stop() error
	ProcessData(input interface{}) (interface{}, error)
}
