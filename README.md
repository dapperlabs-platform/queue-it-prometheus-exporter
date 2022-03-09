# QUEUE-IT-PROMETHEUS-EXPORTER

Prometheus exporter for [Queue-it](https://queue-it.com/) metrics.

## Installation and usage

### Configuration

The exporter expects the following configuration flags:

| name                           | description                                           | default value |
| ------------------------------ | ----------------------------------------------------- | ------------- |
| config.queue-it-base-url       | Base URL to your Queue-it api                         |               |
| config.queue-it-api-key-path   | Absolute path to Queue-it API Key file.               |               |
| config.omit-test-waiting-rooms | Whether to filter out test waiting rooms metrics      | true          |
| web.listen-address             | Address on which to expose metrics and web interface. | :8000         |
| web.telemetry-path             | Path under which to expose metrics.                   | /metrics      |
| web.healthcheck-path           | Path under which to run healthchecks                  | /healthz      |

> If provided, a `QUEUE_IT_API_KEY` environment variable supersedes the `config.queue-it-api-key-path` config

### Build from source

Build a binary with `make build-local` and run it with the appropriate flags and environment variables and run it with

```sh
$ QUEUE_IT_API_KEY=foo ./queue-it-prometheus-exporter -config.queue-it-base-url=https://<account>.api2.queue-it.net
```

### From a docker container

Run a container from the [repo registry](https://github.com/dapperlabs-platform/queue-it-prometheus-exporter/pkgs/container/queue-it-prometheus-exporter)

```sh
$ docker run \
  -v /path/to/queue-it-api-key:/queue-it-api-key \
  ghcr.io/dapperlabs-platform/queue-it-prometheus-exporter:latest \
  -config.queue-it-base-url=https://<account>.api2.queue-it.net \
  -config.queue-it-api-key-path=/queue-it-api-key
```

Have a [Prometheus scrape config](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config) discover the process or container on the provided path/port (:8000/metrics default) and you're good to go.

## Exported metrics

Metrics are pulled from 2 statistics endpoints from [Queue-it API](https://api2.queue-it.net/swagger/index.html):

- `/2_0/event/{waitingRoomId}/queue/statistics/summary` provides a timestamped snapshot of metric values
- `/2_0/event/{waitingRoomId}/queue/statistics/details/{statisticType}` provides per-minute values for metrics as well as the overall sum for the metric up to the minute in question.

Queue-it metrics don't follow [Prometheus naming conventions](https://prometheus.io/docs/practices/naming/) so we rename them before exporting:

> All metrics are exporter with a `waiting_room_id` label

| Queue-it name                    | exported name                                   |
| -------------------------------- | ----------------------------------------------- |
| TotalQueueCount                  | queue_it_total_queue_count                      |
| TotalQueueCountBeforeStart       | queue_it_total_queue_count_before_start         |
| TotalWaitingInQueueCount         | queue_it_total_waiting_in_queue_count           |
| TotalLeftQueueCount              | queue_it_total_left_queue_count                 |
| NoOfRedirectsLastMinute          | queue_it_no_of_redirects_last_minute            |
| NoOfUniqueRedirectsLastMinute    | queue_it_no_of_unique_redirects_last_minute     |
| SafetyNetRedirectedCount         | queue_it_safety_net_redirected_count            |
| RedirectorRedirectedCount        | queue_it_redirector_redirected_count            |
| TotalRedirectedCount             | queue_it_total_redirected_count                 |
| TotalEmailCount                  | queue_it_total_email_count                      |
| TotalEmailNotificationCount      | queue_it_total_email_notification_count         |
| TotalOldQueueNumbers             | queue_it_total_old_queue_numbers                |
| TotalExceededMaxRedirectCount    | queue_it_total_exceeded_max_redirect_count      |
| queuebeforeeventinflow           | queue_it_queue_before_event_inflow_count        |
| queueinflow                      | queue_it_queue_inflow_count                     |
| queueuniqueoutflow               | queue_it_queue_unique_outflow_count             |
| queueoutflow                     | queue_it_queue_outflow_count                    |
| queueoutflow (Accumulated)       | queue_it_queue_outflow_accumulated              |
| safetynetoutflow                 | queue_it_safety_net_outflow_count               |
| queueidsinqueue                  | queue_it_queue_ids_in_queue_count               |
| queueuniqueinflow                | queue_it_queue_unique_inflow_count              |
| queueidscanceled                 | queue_it_queue_ids_canceled_count               |
| notificationfirst                | queue_it_notification_first_count               |
| notificationyourturn             | queue_it_notification_your_turn_count           |
| exceededmaxredirectcount         | queue_it_exceeded_max_redirect_count            |
| maxoutflow                       | queue_it_max_out_flow                           |
| queueexpectedwaittime            | queue_it_queue_expected_wait_time               |
| queueactualwaittime              | queue_it_queue_actual_wait_time                 |
| returningqueueitemsinlessthan30s | queue_it_returning_queue_items_in_less_than_30s |
| oldqueuenumbers                  | queue_it_old_queue_numbers_count                |
| redirectedpercentage             | queue_it_redirected_percentage                  |
