package mpfireworq

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"

	mp "github.com/mackerelio/go-mackerel-plugin"
)

type FireworqStats struct {
	TotalPushes     int64 `json:"total_pushes"`
	TotalPops       int64 `json:"total_pops"`
	TotalSuccesses  int64 `json:"total_successes"`
	TotalFailures   int64 `json:"total_failures"`
	TotalCompletes  int64 `json:"total_completes"`
	TotalElapsed    int64 `json:"total_elapsed"`
	OutstandingJobs int64 `json:"outstanding_jobs"`
	TotalWorkers    int64 `json:"total_workers"`
	IdleWorkers     int64 `json:"idle_workers"`
}

type FireworqPlugin struct {
	URI         string
	Prefix      string
	LabelPrefix string
}

func (p FireworqPlugin) FetchMetrics() (map[string]float64, error) {
	resp, err := http.Get(p.URI + "/queues/stats")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	stats := make(map[string]*FireworqStats)
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&stats)
	if err != nil {
		return nil, err
	}

	var sum FireworqStats
	for _, s := range stats {
		sum.TotalPushes += s.TotalPushes
		sum.TotalPops += s.TotalPops
		sum.TotalSuccesses += s.TotalSuccesses
		sum.TotalFailures += s.TotalFailures
		sum.TotalCompletes += s.TotalCompletes
		sum.TotalElapsed += s.TotalElapsed
		sum.OutstandingJobs += s.OutstandingJobs
		sum.TotalWorkers += s.TotalWorkers
		sum.IdleWorkers += s.IdleWorkers
	}

	m := make(map[string]float64)
	m["queue_running_workers"] = float64(sum.TotalWorkers - sum.IdleWorkers)
	m["queue_idle_workers"] = float64(sum.IdleWorkers)
	m["queue_outstanding_jobs"] = float64(sum.OutstandingJobs)
	m["jobs_failure"] = float64(sum.TotalCompletes - sum.TotalSuccesses)
	m["jobs_success"] = float64(sum.TotalSuccesses)
	m["jobs_outstanding"] = float64(sum.TotalPops - sum.TotalCompletes)
	if sum.TotalPushes > sum.TotalPops {
		m["jobs_waiting"] = float64(sum.TotalPushes - sum.TotalPops)
	} else {
		m["jobs_waiting"] = 0
	}
	m["jobs_events_pushed"] = float64(sum.TotalPushes)
	m["jobs_events_popped"] = float64(sum.TotalPops)
	m["jobs_events_failed"] = float64(sum.TotalFailures)
	m["jobs_events_succeeded"] = float64(sum.TotalSuccesses)
	m["jobs_events_completed"] = float64(sum.TotalCompletes)
	if sum.TotalCompletes > 0 {
		m["jobs_average_elapsed_time"] = float64(sum.TotalElapsed) / float64(sum.TotalCompletes)
	} else {
		m["jobs_average_elapsed_time"] = 0
	}

	return m, nil
}

func (p FireworqPlugin) GraphDefinition() map[string]mp.Graphs {
	graphdef := map[string]mp.Graphs{
		"queue.workers": {
			Label: p.LabelPrefix + " Queue Workers",
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "queue_idle_workers", Label: "Idle", Stacked: true},
				{Name: "queue_running_workers", Label: "Running", Stacked: true},
			},
		},
		"queue.buffer": {
			Label: p.LabelPrefix + " Queue Buffer",
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "queue_outstanding_jobs", Label: "Outstanding Jobs"},
			},
		},
		"jobs": {
			Label: p.LabelPrefix + " Jobs",
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "jobs_failure", Label: "Failure", Diff: true, Stacked: true},
				{Name: "jobs_success", Label: "Success", Diff: true, Stacked: true},
				{Name: "jobs_outstanding", Label: "Outstanding", Diff: true, Stacked: true},
				{Name: "jobs_waiting", Label: "Waiting", Diff: true, Stacked: true},
			},
		},
		"jobs.events": {
			Label: p.LabelPrefix + " Job Events",
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "jobs_events_pushed", Label: "Pushed", Diff: true, Stacked: true},
				{Name: "jobs_events_popped", Label: "Popped", Diff: true, Stacked: true},
				{Name: "jobs_events_failed", Label: "Failed", Diff: true, Stacked: true},
				{Name: "jobs_events_succeeded", Label: "Succeeded", Diff: true, Stacked: true},
				{Name: "jobs_events_completed", Label: "Completed", Diff: true, Stacked: true},
			},
		},
		"jobs.elapsed": {
			Label: p.LabelPrefix + " Elapsed Time Per Completed Job in ms",
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "jobs_average_elapsed_time", Label: "Average"},
			},
		},
	}
	return graphdef
}

func (p FireworqPlugin) MetricKeyPrefix() string {
	return p.Prefix
}

func Do() {
	optScheme := flag.String("scheme", "http", "Scheme")
	optHost := flag.String("host", "localhost", "Host")
	optPort := flag.String("port", "8080", "Port")
	optPrefix := flag.String("metric-key-prefix", "fireworq", "Metric key prefix")
	optLabelPrefix := flag.String("metric-label-prefix", "", "Metric Label prefix")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	var fireworq FireworqPlugin
	fireworq.URI = fmt.Sprintf("%s://%s:%s", *optScheme, *optHost, *optPort)
	fireworq.Prefix = *optPrefix
	if *optLabelPrefix == "" {
		*optLabelPrefix = strings.Title(*optPrefix)
	}
	fireworq.LabelPrefix = *optLabelPrefix

	helper := mp.NewMackerelPlugin(fireworq)
	if *optTempfile != "" {
		helper.Tempfile = *optTempfile
	} else {
		helper.SetTempfileByBasename(fmt.Sprintf("mackerel-plugin-fireworq-%s-%s", *optHost, *optPort))
	}

	helper.Run()
}
