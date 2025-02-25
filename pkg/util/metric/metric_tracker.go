package metric

import (
	prometheusgo "github.com/prometheus/client_model/go"
	"sync"
	"time"
)

type ValuePoint struct {
	FloatValue        float64
	ObservedTimestamp int64
}

type State struct {
	sync.Mutex
	PrevPoint ValuePoint
}

type DeltaValue struct {
	StartTimestamp int64
	FloatValue     float64
}

type MetricTracker struct {
	states    sync.Map
	startTime int64
}

func NewMetricTracker() *MetricTracker {
	t := &MetricTracker{
		startTime: time.Now().Unix(),
	}
	return t
}

func (t *MetricTracker) Transform(in *prometheusgo.Metric) {
	delta, valid := t.convert(in)
	if !valid {
		return
	}
	if in.GetGauge() != nil {
		*in.Gauge.Value = delta.FloatValue
	} else if in.GetCounter() != nil {
		*in.Counter.Value = delta.FloatValue
	}
	return
}

func (t *MetricTracker) convert(in *prometheusgo.Metric) (out DeltaValue, valid bool) {
	metricName := in.String()
	metricValue, isValid := getMetricValue(*in)
	if !isValid {
		return DeltaValue{}, false
	}
	metricTimestamp := in.TimestampMs

	s, ok := t.states.LoadOrStore(metricName, &State{
		PrevPoint: ValuePoint{FloatValue: metricValue, ObservedTimestamp: *metricTimestamp},
	})
	if !ok {
		out.FloatValue = metricValue
		out.StartTimestamp = *metricTimestamp
		return
	}

	state := s.(*State)
	state.Lock()
	defer state.Unlock()

	out.StartTimestamp = state.PrevPoint.ObservedTimestamp

	//Delta calculation
	value := metricValue
	prevValue := state.PrevPoint.FloatValue
	delta := value - prevValue
	out.FloatValue = delta

	state.PrevPoint = ValuePoint{FloatValue: metricValue, ObservedTimestamp: *metricTimestamp}
	return
}
func getMetricValue(metric prometheusgo.Metric) (float64, bool) {
	if metric.Gauge != nil {
		return metric.Gauge.GetValue(), true
	}
	if metric.Counter != nil {
		return metric.Counter.GetValue(), true
	}
	return 0, false
}
