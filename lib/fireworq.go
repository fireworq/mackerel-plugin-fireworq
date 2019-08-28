package mpfireworq

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	mp "github.com/mackerelio/go-mackerel-plugin"
)

var (
	// We need escape queue name to confirm to metric name specification.
	// See also: https://mackerel.io/ja/api-docs/entry/host-metrics#post-graphdef
	invalidNameReg = regexp.MustCompile("[^-a-zA-Z0-9_]")
)

// FireworqStats represents the statistics of Fireworq
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
	ActiveNodes     int64 `json:"active_nodes"`
}

// InspectedJob describes a job in a queue.
type InspectedJob struct {
	ID         uint64          `json:"id"`
	Category   string          `json:"category"`
	URL        string          `json:"url"`
	Payload    json.RawMessage `json:"payload,omitempty"`
	Status     string          `json:"status"`
	CreatedAt  time.Time       `json:"created_at"`
	NextTry    time.Time       `json:"next_try"`
	Timeout    uint            `json:"timeout"`
	FailCount  uint            `json:"fail_count"`
	MaxRetries uint            `json:"max_retries"`
	RetryDelay uint            `json:"retry_delay"`
}

// InspectedJobs describes a (page of) job list in a queue.
type InspectedJobs struct {
	Jobs       []InspectedJob `json:"jobs"`
	NextCursor string         `json:"next_cursor"`
}

// FireworqPlugin is a mackerel plugin
type FireworqPlugin struct {
	URI         string
	Prefix      string
	LabelPrefix string
}

func (p FireworqPlugin) fetchJob(queue string, list string) (*InspectedJob, error) {
	resp, err := http.Get(p.URI + fmt.Sprintf("/queue/%s/%s?order=asc&limit=1", queue, list))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jobs *InspectedJobs
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&jobs)
	if err != nil {
		return nil, err
	}
	if len(jobs.Jobs) > 0 {
		return &jobs.Jobs[0], nil
	}

	return nil, nil
}

func (p FireworqPlugin) fetchMostDelayedJob(queue string) (*InspectedJob, error) {
	job, err := p.fetchJob(queue, "waiting")
	if err != nil {
		return nil, err
	}

	if job != nil {
		return job, nil
	}

	return p.fetchJob(queue, "grabbed")
}

// FetchMetrics fetchs the metrics
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
		sum.ActiveNodes += s.ActiveNodes
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
	m["active_nodes"] = float64(sum.ActiveNodes)
	m["active_nodes_percentage"] = float64(sum.ActiveNodes*100) / float64(len(stats))

	for q, s := range stats {
		if s.ActiveNodes >= 1 {
			q = invalidNameReg.ReplaceAllString(q, "-")

			if job, err := p.fetchMostDelayedJob(q); err == nil {
				var delay float64
				if job != nil {
					delay = float64(time.Since(job.NextTry).Seconds())
				}
				m[fmt.Sprintf("queue.delay.%s", q)] = delay
			}
		}
	}

	return m, nil
}

// GraphDefinition of FireworqPlugin
func (p FireworqPlugin) GraphDefinition() map[string]mp.Graphs {
	graphdef := map[string]mp.Graphs{
		"node": {
			Label: p.LabelPrefix + " Node",
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "active_nodes", Label: "Active"},
				{Name: "active_nodes_percentage", Label: "Active (%)"},
			},
		},
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
		"queue.delay": {
			Label: p.LabelPrefix + " Delayed Time in sec",
			Unit:  "float",
			Metrics: []mp.Metrics{
				{Name: "*", Label: "%1"},
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

// MetricKeyPrefix of FireworqPlugin
func (p FireworqPlugin) MetricKeyPrefix() string {
	return p.Prefix
}

// Do the plugin
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
