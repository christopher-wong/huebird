package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	apiPollSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "nfl_api_poll_success_total",
		Help: "Total number of successful API polls",
	})

	apiPollFailure = promauto.NewCounter(prometheus.CounterOpts{
		Name: "nfl_api_poll_failure_total",
		Help: "Total number of failed API polls",
	})

	apiDecodeSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "nfl_api_decode_success_total",
		Help: "Total number of successful JSON decodes",
	})

	apiDecodeFailure = promauto.NewCounter(prometheus.CounterOpts{
		Name: "nfl_api_decode_failure_total",
		Help: "Total number of failed JSON decodes",
	})

	natsKVPutSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "nfl_nats_kv_put_success_total",
		Help: "Total number of successful NATS KV puts",
	})

	natsKVPutFailure = promauto.NewCounter(prometheus.CounterOpts{
		Name: "nfl_nats_kv_put_failure_total",
		Help: "Total number of failed NATS KV puts",
	})

	scoreChangeTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "nfl_score_changes_total",
		Help: "Total number of score changes detected",
	})
)
