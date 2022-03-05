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
	value              float64
}

// queueitAPI represents a Queue-it API client
type queueitAPI struct {
	logger               *zap.Logger
	apiKey               string
	baseUrl              string
	detailNameToMetric   map[string]queueitMetric
	omitTestWaitingRooms bool
}

// Custom unmarshallers
type stringToBool bool
type stringToTime struct {
	time.Time
}

func (t *stringToBool) UnmarshalJSON(data []byte) error {
	str := strings.ToLower(strings.Replace(string(data), "\"", "", 2))

	*t = str == "true"

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

type StatisticsSummary struct {
	VersionTimestamp                        stringToTime `json:"VersionTimestamp"`
	TotalQueueCount                         float64      `json:",string"`
	TotalQueueCountBeforeStart              float64      `json:",string"`
	TotalWaitingInQueueCount                float64      `json:",string"`
	TotalLeftQueueCount                     float64      `json:",string"`
	NoOfRedirectsLastMinute                 float64      `json:",string"`
	NoOfUniqueRedirectsLastMinute           float64      `json:",string"`
	SafetyNetRedirectedCount                float64      `json:",string"`
	RedirectorRedirectedCount               float64      `json:",string"`
	TotalRedirectedCount                    float64      `json:",string"`
	TotalEmailCount                         float64      `json:",string"`
	TotalEmailNotificationCount             float64      `json:",string"`
	TotalOldQueueNumbers                    float64      `json:",string"`
	TotalExceededMaxRedirectCount           float64      `json:",string"`
	ReturningQueueItemsInLessThan30SLastMin float64      `json:",string"`
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
