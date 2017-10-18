package mpfireworq

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

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

	m := make(map[string]float64)
	var totalCompletes int64
	var totalElapsed int64
	for _, s := range stats {
		totalCompletes += s.TotalCompletes
		totalElapsed += s.TotalElapsed
		m["queue_running_workers"] += float64(s.TotalWorkers - s.IdleWorkers)
		m["queue_idle_workers"] += float64(s.IdleWorkers)
		m["queue_outstanding_jobs"] += float64(s.OutstandingJobs)
		m["jobs_failure"] += float64(s.TotalCompletes - s.TotalSuccesses)
		m["jobs_success"] += float64(s.TotalSuccesses)
		m["jobs_outstanding"] += float64(s.TotalPops - s.TotalCompletes)
		m["jobs_waiting"] += float64(s.TotalPushes - s.TotalPops)
		m["jobs_events_pushed"] += float64(s.TotalPushes)
		m["jobs_events_popped"] += float64(s.TotalPops)
		m["jobs_events_failed"] += float64(s.TotalFailures)
		m["jobs_events_succeeded"] += float64(s.TotalSuccesses)
		m["jobs_events_completed"] += float64(s.TotalCompletes)
	}
	if totalCompletes > 0 {
		m["jobs_average_elapsed_time"] = float64(totalElapsed) / float64(totalCompletes)
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
				{Name: "queue_running_workers", Label: "Running", Stacked: true},
				{Name: "queue_idle_workers", Label: "Idle", Stacked: true},
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
		*optLabelPrefix = *optPrefix
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
