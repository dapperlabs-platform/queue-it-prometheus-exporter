package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	GAUGE = "gauge"
	// Number of metrics expected from StatisticsSummary
	SUMMARY_METRIC_COUNT = 14
	// Number of metrics iterated over in getStatisticsDetailsMetrics
	DETAILS_METRIC_COUNT = 17
	TOTAL_METRIC_COUNT   = SUMMARY_METRIC_COUNT + DETAILS_METRIC_COUNT
)

// newQueueitAPI creates a queueitAPI
func newQueueitAPI(logger *zap.Logger, baseURL string, apiKey string, omitTestWaitingRooms bool) *queueitAPI {
	return &queueitAPI{
		logger:               logger,
		apiKey:               apiKey,
		baseUrl:              baseURL,
		omitTestWaitingRooms: omitTestWaitingRooms,
	}
}

// doRequest executes an HTTP request and returns the body and error
func (q *queueitAPI) doRequest(method string, path string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", q.baseUrl, path), body)
	if err != nil {
		return nil, err
	}
	q.addHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

// addHeaders adds the required queue-it API headers to a request
func (q *queueitAPI) addHeaders(req *http.Request) {
	req.Header.Add("Api-Key", q.apiKey)
	req.Header.Add("Content-Type", "application/json;charset=utf-8")
}

// handleAPIError inspects the body of a successful queue-it API call for an error
// response. API responses from queue-it return 200 OK with an error JSON when failed
func (q *queueitAPI) handleAPIError(body []byte, incomingError error) error {
	var apiError APIError
	err := json.Unmarshal(body, &apiError)
	// if body fails to unmarshal to an apiError object log everything and
	// return unmarshal error
	if err != nil {
		q.logger.Info("unknown queue-it api error", zap.Error(err), zap.String("body", string(body)))
		return err
	}

	q.logger.Debug("queue-it api error", zap.Int("code", apiError.ErrorCode), zap.String("text", apiError.ErrorText))

	return fmt.Errorf("failed to connecto to queue-it api: %v", apiError)
}

// dropTestWaitingRooms removes test waiting rooms because queue-it API lacks a way to filter them out
func (q *queueitAPI) dropTestWaitingRooms(waitingRooms []WaitingRoom) []WaitingRoom {
	result := make([]WaitingRoom, 0)

	for _, wr := range waitingRooms {
		if !wr.IsTest {
			result = append(result, wr)
		}
	}

	return result
}

// getOpenWaitingRooms returns waiting ongoing waiting rooms
func (q *queueitAPI) getOpenWaitingRooms() ([]WaitingRoom, error) {
	input := []map[string]string{
		{
			"Name":     "Phase",
			"Operator": "in",
			"Value":    "prequeue, queue",
		},
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	body, err := q.doRequest("POST", "/2_0/event/search", strings.NewReader(string(inputJSON)))
	if err != nil {
		return nil, err
	}

	rooms := make([]WaitingRoom, 0)
	if err := json.Unmarshal(body, &rooms); err != nil {
		return nil, q.handleAPIError(body, err)
	}

	q.logger.Debug("queueitAPI.getOpenWaitingRooms(): fetched waiting rooms", zap.Int("count", len(rooms)))

	if q.omitTestWaitingRooms {
		rooms = q.dropTestWaitingRooms(rooms)
	}

	q.logger.Debug("queueitAPI.getOpenWaitingRooms(): filtered out testing waiting rooms", zap.Int("count", len(rooms)))

	return rooms, nil
}

// sendSummaryMetrics sends StatisticsSummary metrics to a channel
func (q *queueitAPI) sendSummaryMetrics(m *StatisticsSummary, waitingRoomID string, c chan *queueitMetric) {
	// SUMMARY_METRIC_COUNT must be set to the number of metrics sent from here
	c <- &queueitMetric{exportedMetricName: "queue_it_total_queue_count", value: m.TotalQueueCount, waitingRoomID: waitingRoomID}
	c <- &queueitMetric{exportedMetricName: "queue_it_total_queue_count_before_start", value: m.TotalQueueCountBeforeStart, waitingRoomID: waitingRoomID}
	c <- &queueitMetric{exportedMetricName: "queue_it_total_waiting_in_queue_count", value: m.TotalWaitingInQueueCount, waitingRoomID: waitingRoomID}
	c <- &queueitMetric{exportedMetricName: "queue_it_total_left_queue_count", value: m.TotalLeftQueueCount, waitingRoomID: waitingRoomID}
	c <- &queueitMetric{exportedMetricName: "queue_it_no_of_redirects_last_minute", value: m.NoOfRedirectsLastMinute, waitingRoomID: waitingRoomID}
	c <- &queueitMetric{exportedMetricName: "queue_it_no_of_unique_redirects_last_minute", value: m.NoOfUniqueRedirectsLastMinute, waitingRoomID: waitingRoomID}
	c <- &queueitMetric{exportedMetricName: "queue_it_safety_net_redirected_count", value: m.SafetyNetRedirectedCount, waitingRoomID: waitingRoomID}
	c <- &queueitMetric{exportedMetricName: "queue_it_redirector_redirected_count", value: m.RedirectorRedirectedCount, waitingRoomID: waitingRoomID}
	c <- &queueitMetric{exportedMetricName: "queue_it_total_redirected_count", value: m.TotalRedirectedCount, waitingRoomID: waitingRoomID}
	c <- &queueitMetric{exportedMetricName: "queue_it_total_email_count", value: m.TotalEmailCount, waitingRoomID: waitingRoomID}
	c <- &queueitMetric{exportedMetricName: "queue_it_total_email_notification_count", value: m.TotalEmailNotificationCount, waitingRoomID: waitingRoomID}
	c <- &queueitMetric{exportedMetricName: "queue_it_total_old_queue_numbers", value: m.TotalOldQueueNumbers, waitingRoomID: waitingRoomID}
	c <- &queueitMetric{exportedMetricName: "queue_it_total_exceeded_max_redirect_count", value: m.TotalExceededMaxRedirectCount, waitingRoomID: waitingRoomID}
	c <- &queueitMetric{exportedMetricName: "queue_it_returning_queue_items_in_less_than_30s_last_min", value: m.ReturningQueueItemsInLessThan30SLastMin, waitingRoomID: waitingRoomID}
}

// getWaitingRoomQueueStatisticsSummary sends metrics from the queue statistics summary api
// to the provided channel
// If the API call fails the challen will be fed a nil value
func (q *queueitAPI) getWaitingRoomQueueStatisticsSummary(id string, c chan *queueitMetric) {
	body, err := q.doRequest("GET", fmt.Sprintf("/2_0/event/%s/queue/statistics/summary", id), nil)
	if err != nil {
		// throw returned error away but log it
		q.handleAPIError(body, err)
		c <- nil
	}

	// turn JSON map into list of metrics
	var metrics StatisticsSummary
	err = json.Unmarshal(body, &metrics)
	if err != nil {
		q.logger.Info(
			"queueitAPI.getWaitingRoomQueueStatisticsSummary(): failed to unmarshal stats",
			zap.String("body", string(body)),
			zap.Error(err),
		)
		c <- nil
	}

	// send metrics to channel
	q.sendSummaryMetrics(&metrics, id, c)
}

func (q *queueitAPI) getStatisticsDetailsMetrics(id string, c chan *queueitMetric) {
	statisticsDetailsMetrics := []*queueitMetric{
		{queueitMetricName: "queuebeforeeventinflow", exportedMetricName: "queue_it_queue_before_event_inflow_count", description: "The amount of users who have joined the pre-queue"},
		{queueitMetricName: "queueinflow", exportedMetricName: "queue_it_queue_inflow_count", description: "Users who have joined either the pre-queue or the queue"},
		{queueitMetricName: "queueuniqueoutflow", exportedMetricName: "queue_it_queue_unique_outflow_count", description: "The number of initial queue redirects per minute (first redirect of the queue ID)"},
		{queueitMetricName: "queueoutflow", exportedMetricName: "queue_it_queue_outflow_count", description: "The amount of queue numbers which have been redirected from the queue"},
		{queueitMetricName: "safetynetoutflow", exportedMetricName: "queue_it_safety_net_outflow_count", description: "Redirected queue numbers which were redirected without having waited in the queue (requires Always Visible, so this value is irrelevant in your case)"},
		{queueitMetricName: "queueidsinqueue", exportedMetricName: "queue_it_queue_ids_in_queue_count", description: "The amount of Queue IDs currently waiting in line"},
		{queueitMetricName: "queueuniqueinflow", exportedMetricName: "queue_it_queue_unique_inflow_count", description: "The amount of new (unique) Queue IDs entering the queue per minute"},
		{queueitMetricName: "queueidscanceled", exportedMetricName: "queue_it_queue_ids_canceled_count", description: "The amount of Queue IDs which have been canceled by Cancel Action or API"},
		{queueitMetricName: "notificationfirst", exportedMetricName: "queue_it_notification_first_count", description: "The amount of users who received the first email notification upon signing up"},
		{queueitMetricName: "notificationyourturn", exportedMetricName: "queue_it_notification_your_turn_count", description: "The amount of users who received the It's Your Turn email notification"},
		{queueitMetricName: "exceededmaxredirectcount", exportedMetricName: "queue_it_exceeded_max_redirect_count", description: "The amount of visitors who pass through the waiting room more times than they are allowed (as configured in the Waiting Room Settings)"},
		{queueitMetricName: "maxoutflow", exportedMetricName: "queue_it_max_out_flow", description: "The highest amount of Queue IDs which are allowed to be redirected to your site per minute"},
		{queueitMetricName: "queueexpectedwaittime", exportedMetricName: "queue_it_queue_expected_wait_time", description: "For users arriving at a given time, this is the predicted wait time"},
		{queueitMetricName: "queueactualwaittime", exportedMetricName: "queue_it_queue_actual_wait_time", description: "The actual amount of minutes wait time in the queue"},
		{queueitMetricName: "returningqueueitemsinlessthan30s", exportedMetricName: "queue_it_returning_queue_items_in_less_than_30s", description: "If a Queue ID is returning to the queue less than 30 seconds after it was redirected to the target site, we count it as a fast re-entering user"},
		{queueitMetricName: "oldqueuenumbers", exportedMetricName: "queue_it_old_queue_numbers_count", description: "The amount of Queue IDs who have been first in line and did not choose to be redirected to the target site"},
		{queueitMetricName: "redirectedpercentage", exportedMetricName: "queue_it_redirected_percentage", description: "Percent of users who took their turn within a minute"},
	}

	now := time.Now()
	then := now.Add(-1 * time.Minute)

	for _, m := range statisticsDetailsMetrics {
		go q.getWaitingRoomQueueStatisticsDetail(id, m, then, now, c)
	}
}

// getWaitingRoomQueueStatisticsDetail sends a metric from the queue statistics details api
// to the provided channel
// If the API call fails the challen will be fed a nil value
func (q *queueitAPI) getWaitingRoomQueueStatisticsDetail(id string, m *queueitMetric, from time.Time, to time.Time, statsChan chan *queueitMetric) {
	fromQueryParam := url.QueryEscape(from.Format(time.RFC3339))
	toQueryParam := url.QueryEscape(to.Format(time.RFC3339))

	q.logger.Debug("queueitAPI.getWaitingRoomQueueStatisticsDetails(): getting statistics details", zap.String("waitingRoomId", id), zap.Time("from", from), zap.Time("to", to))

	body, err := q.doRequest("GET", fmt.Sprintf("/2_0/event/%s/queue/statistics/details/%s?from=%s&to=%s", id, m.queueitMetricName, fromQueryParam, toQueryParam), nil)
	if err != nil {
		// throw returned error away but log it
		q.handleAPIError(body, err)
		statsChan <- nil
	}

	var metric StatisticsDetail
	err = json.Unmarshal(body, &metric)
	if err != nil {
		q.logger.Info("queueitAPI.parseStatisticsDetailMetrics(): failed to unmarshal stats", zap.Error(err))
		statsChan <- nil
	}

	// deal with potentially empty Entries array
	var value float64
	if len(metric.Entries) == 0 {
		q.logger.Info("queueitAPI.parseStatisticsDetailMetrics(): stat detail metric has no value", zap.String("type", m.queueitMetricName))
	} else {
		value = metric.Entries[0].Sum
	}

	statsChan <- &queueitMetric{
		exportedMetricName: m.exportedMetricName,
		queueitMetricName:  m.queueitMetricName,
		waitingRoomID:      id,
		value:              value,
	}
}

// getMetrics queries the api for metrics from all active waiting rooms
func (q *queueitAPI) getMetrics() ([]*queueitMetric, error) {
	metrics := make([]*queueitMetric, 0)

	// Get active rooms we want to collect metrics for
	rooms, err := q.getOpenWaitingRooms()
	if err != nil {
		return nil, err
	}

	if len(rooms) == 0 {
		q.logger.Info("queueitAPI.getMetrics(): did not find any waiting room")
		return nil, nil
	}

	q.logger.Debug("queueitAPI.getMetrics(): found rooms", zap.Int("count", len(rooms)))

	q.logger.Debug("queueitAPI.getMetrics(): number of expected metrics", zap.Int("count", TOTAL_METRIC_COUNT))

	statsChan := make(chan *queueitMetric, TOTAL_METRIC_COUNT)
	defer func() {
		q.logger.Debug("queueitAPI.getMetrics(): cleaning up, closing channel")
		close(statsChan)
	}()

	// fan out fetching of summary and detail metrics
	for _, room := range rooms {
		// get summary metrics for waiting room
		go q.getWaitingRoomQueueStatisticsSummary(room.EventID, statsChan)
		// get waiting room detail metrics for the last minute
		go q.getStatisticsDetailsMetrics(room.EventID, statsChan)
	}

	// fan in metrics
	for n := 0; n < TOTAL_METRIC_COUNT; n++ {
		stat := <-statsChan

		if stat == nil {
			return nil, fmt.Errorf("queueitAPI.getMetrics(): failed to get statistics for waiting room")
		}

		metrics = append(metrics, stat)

		q.logger.Debug("queueitAPI.getMetrics(): done getting metrics",
			zap.String("waiting_room_id", stat.waitingRoomID),
			zap.Float64(stat.queueitMetricName, stat.value),
		)
	}

	return metrics, nil
}
