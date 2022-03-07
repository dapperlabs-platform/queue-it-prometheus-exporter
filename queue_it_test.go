package main

import (
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
	type want struct {
		len     int
		eventID string
	}
	type test struct {
		input []WaitingRoom
		want  want
	}
	q := &queueitAPI{}
	tests := []test{
		{input: []WaitingRoom{{EventID: "1", IsTest: true}, {EventID: "2", IsTest: false}}, want: want{len: 1, eventID: "2"}},
	}

	for _, tc := range tests {
		got := q.dropTestWaitingRooms(tc.input)

		if tc.want.len != len(got) || tc.want.eventID != got[0].EventID {
			t.Fail()
		}
	}
}

// If this test times out then sendSummaryMetrics is sending
// fewer than SUMMARY_METRIC_COUNT metrics to the channel
func TestSendSummaryMetrics(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	q := &queueitAPI{
		logger: logger,
	}

	input := &StatisticsSummary{}

	c := make(chan *queueitMetric, SUMMARY_METRIC_COUNT*2)
	defer func() {
		logger.Sync()
		close(c)
	}()

	go func() {
		q.sendSummaryMetrics(input, "id", c)
	}()

	for n := 0; n < SUMMARY_METRIC_COUNT; n++ {
		<-c
	}

	if len(c) > 0 {
		t.Error("sendSummaryMetrics sent more than SUMMARY_METRIC_COUNT to channel")
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

	m := &queueitMetric{
		queueitMetricName:  "queueinflow",
		exportedMetricName: "queue_it_queue_inflow",
	}

	c := make(chan *queueitMetric, 1)
	now := time.Now()
	then := now.Add(-1 * time.Minute)
	q.getWaitingRoomQueueStatisticsDetail("0224allstarstdgqr3", m, then, now, c)
	go func() {
		time.Sleep(5 * time.Second)
		close(c)
	}()

	for m := range c {
		fmt.Println("->>", m)
	}
}
