package mpfireworq

type QueueStat int64

const (
	Pushes            QueueStat = iota
	Pops              QueueStat = iota
	Successes         QueueStat = iota
	Failures          QueueStat = iota
	PermanentFailures QueueStat = iota
	Completes         QueueStat = iota
)

func (s QueueStat) String() string {
	switch s {
	case Pushes:
		return "pushes"
	case Pops:
		return "pops"
	case Successes:
		return "successes"
	case Failures:
		return "failures"
	case PermanentFailures:
		return "permanent_failures"
	case Completes:
		return "completed"
	}
	return "unknown"
}

func (s QueueStat) MetricName() string {
	return "queue." + s.String()
}

func (s QueueStat) Metric(f *FireworqStats) int64 {
	switch s {
	case Pushes:
		return f.TotalPushes
	case Pops:
		return f.TotalPops
	case Successes:
		return f.TotalSuccesses
	case Failures:
		return f.TotalFailures
	case PermanentFailures:
		return f.TotalPermanentFailures
	case Completes:
		return f.TotalCompletes
	}
	return 0
}

func (s QueueStat) Label() string {
	switch s {
	case Pushes:
		return "Pushed Jobs"
	case Pops:
		return "Popped Jobs"
	case Successes:
		return "Succeeded Jobs"
	case Failures:
		return "Failed Jobs"
	case PermanentFailures:
		return "Permanently Failed Jobs"
	case Completes:
		return "Completed Jobs"
	}
	return "Unknown Stat"
}

func NewQueueStat(s string) QueueStat {
    switch s {
        case "pushes":
			return Pushes
		case "pops":
			return Pops
		case "successes":
			return Successes
		case "failures":
			return Failures
		case "permanent_failures":
			return PermanentFailures
		case "completes":
			return Completes
	}
	panic("Unknown queue stat string: " + s)
}
