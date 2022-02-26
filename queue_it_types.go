package main

import (
	"strings"
	"time"

	"go.uber.org/zap"
)

// queueitMetric represents a queue it metric for a waiting room
type queueitMetric struct {
	// Metric name as provided by the Queue-it api
	queueitMetricName string
	// snake_cased, optinally suffixed version of queueitMetricName
	exportedMetricName string
	description        string
	waitingRoomID      string
	metricType         string
	value              float64
}

// queueitAPI represents a Queue-it API client
type queueitAPI struct {
	logger               *zap.Logger
	apiKey               string
	baseUrl              string
	summaryNameToMetric  map[string]queueitMetric
	detailNameToMetric   map[string]queueitMetric
	omitTestWaitingRooms bool
}

// queueitMetricsByType groups metrics by their types
type queueitMetricsByType struct {
	gauges []*queueitMetric
}

// Custom unmarshallers
type stringToBool bool
type stringToTime struct {
	time.Time
}

func (t *stringToBool) UnmarshalJSON(data []byte) error {
	str := strings.ToLower(strings.Replace(string(data), "\"", "", 2))

	if str == "true" {
		*t = true
	} else {
		*t = false
	}

	return nil
}

func (t *stringToTime) UnmarshalJSON(data []byte) error {
	str := strings.Replace(string(data), "\"", "", 2)

	ts, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return err
	}

	t.Time = ts

	return nil
}

// WaitingRoom represents a waiting room data structure as returned by
// the Queue-it API
type WaitingRoom struct {
	EventID                    string
	DisplayName                string
	PreQueueStartsMinuesBefore int          `json:"PreQueueStartsMinuesBefore,string"`
	EventStartTime             stringToTime `json:"EventStartTime"`
	EventEndTime               stringToTime `json:"EventEndTime"`
	QueueStatusText            string
	IsTest                     stringToBool `json:"IsTest"`
}

// APIError represents the error object returned by the Queue-it API as
// part of a failed request's 200 OK body
type APIError struct {
	ErrorCode      int
	ErrorText      string
	HttpStatusCode int
}

// StatisticsDetailEntry represents the per-minute observation count for
// a given detail statistics endpoint metric
type StatisticsDetailEntry struct {
	Sum       float64 `json:",string"`
	MinMinute float64 `json:",string"`
	MaxMinute float64 `json:",string"`
}

// StatisticsDetail represents a statistics detail API response
type StatisticsDetail struct {
	VersionTimestamp string
	From             string
	To               string
	Interval         int `json:",string"`
	Entries          []StatisticsDetailEntry
	SumOffset        int `json:",string"`
}
