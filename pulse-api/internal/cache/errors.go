package cache

import (
	"errors"
)

var (
	// ErrEmptyNodeID is returned when node ID is empty
	ErrEmptyNodeID = errors.New("node ID cannot be empty")
	// ErrNilMetricPoint is returned when metric point is nil
	ErrNilMetricPoint = errors.New("metric point cannot be nil")
	// ErrNilMetricRecord is returned when metric record is nil
	ErrNilMetricRecord = errors.New("metric record cannot be nil")
	// ErrBufferFull is returned when batch writer buffer is full
	ErrBufferFull = errors.New("batch writer buffer is full")
)
