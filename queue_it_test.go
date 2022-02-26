package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestHandlerAddHeaders(t *testing.T) {
	q := &queueitAPI{
		apiKey:  "a-b-c",
		baseUrl: "https://local.server",
	}
	req, _ := http.NewRequest("GET", "https://nowhere.local", nil)
	q.addHeaders(req)
	if req.Header["Api-Key"][0] != q.apiKey || req.Header["Content-Type"][0] != "application/json;charset=utf-8" {
		t.Fail()
	}
}

func TestHandlerDropTestWaitingRooms(t *testing.T) {
	q := &queueitAPI{}
	input := []WaitingRoom{
		{
			EventID: "1",
			IsTest:  true,
		},
		{
			EventID: "2",
			IsTest:  false,
		},
	}

	got := q.dropTestWaitingRooms(input)
	fmt.Println(got)

	if len(got) != 1 || got[0].EventID != "2" {
		t.Fail()
	}
}

func TestParseStatisticsSummaryMetrics(t *testing.T) {
	q := &queueitAPI{}
	dict := map[string]queueitMetric{
		"metric-1": {
			queueitMetricName:  "metric-1",
			exportedMetricName: "queue_it_metric_1",
		},
		"metric-2": {
			queueitMetricName:  "metric-2",
			exportedMetricName: "queue_it_metric_2",
		},
	}

	valid := map[string]string{
		"metric-1": "1.5",
		"metric-2": "2",
	}

	s, _ := json.Marshal(valid)
	got, err := q.parseStatisticsSummaryMetrics("1", s, dict)

	if err != nil {
		t.Fail()
	}

	if len(got) != 2 {
		t.Fail()
	}

	for n := 0; n < len(got); n++ {
		if got[n].exportedMetricName == "metric_1" && got[n].value != 1.5 {
			t.Fail()
		}
		if got[n].exportedMetricName == "metric_2" && got[n].value != 2 {
			t.Fail()
		}
	}

	// invalid number value should fail
	invalid := map[string]string{
		"metric-1": "1.5",
		"metric-2": "abc",
	}

	s, _ = json.Marshal(invalid)
	_, err = q.parseStatisticsSummaryMetrics("1", s, dict)

	if err == nil {
		t.Fail()
	}

	// metric is missing from response should fail
	s, _ = json.Marshal(invalid)
	dict = map[string]queueitMetric{
		"metric-3": {
			queueitMetricName:  "metric-3",
			exportedMetricName: "queue_it_metric_3",
		},
	}
	_, err = q.parseStatisticsSummaryMetrics("1", s, dict)

	if err == nil {
		t.Fail()
	}
}

func SkipTestGetMetrics(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	q := &queueitAPI{
		logger:  logger,
		apiKey:  os.Getenv("QUEUE_IT_API_KEY"),
		baseUrl: os.Getenv("QUEUE_IT_BASE_URL"),
	}
	fmt.Println(q.getMetrics())
}

func SkipBenchmarkGetMetrics(b *testing.B) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	q := &queueitAPI{
		logger:  logger,
		apiKey:  os.Getenv("QUEUE_IT_API_KEY"),
		baseUrl: os.Getenv("QUEUE_IT_BASE_URL"),
	}
	fmt.Println(q.getMetrics())
	for n := 0; n < b.N; n++ {
		q.getMetrics()
	}
}

func SkipTestGetWaitingRoomQueueStatisticsDetail(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	q := &queueitAPI{
		logger:  logger,
		apiKey:  os.Getenv("QUEUE_IT_API_KEY"),
		baseUrl: os.Getenv("QUEUE_IT_BASE_URL"),
	}

	c := make(chan *queueitMetric, 1)
	now := time.Now()
	then := now.Add(-1 * time.Minute)
	q.getWaitingRoomQueueStatisticsDetail("0224allstarstdgqr3", "queueinflow", then, now, c)
	go func() {
		time.Sleep(5 * time.Second)
		close(c)
	}()

	for m := range c {
		fmt.Println("->>", m)
	}
}
