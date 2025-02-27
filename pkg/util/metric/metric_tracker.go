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

func (t *MetricTracker) Transform(in *prometheusgo.Metric, name string) {
	delta, valid := t.convert(in, name)
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

func (t *MetricTracker) convert(in *prometheusgo.Metric, name string) (out DeltaValue, valid bool) {
	metricName := generateMetricName(name, in.Label)
	metricValue, isValid := getMetricValue(*in)
	if !isValid {
		return DeltaValue{}, false
	}
	metricTimestamp := time.Now().Unix()

	s, ok := t.states.LoadOrStore(metricName, &State{
		PrevPoint: ValuePoint{FloatValue: metricValue, ObservedTimestamp: metricTimestamp},
	})

	if !ok {
		out.FloatValue = metricValue
		out.StartTimestamp = metricTimestamp
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

	state.PrevPoint = ValuePoint{FloatValue: metricValue, ObservedTimestamp: metricTimestamp}
	valid = true
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

func generateMetricName(name string, labels []*prometheusgo.LabelPair) string {
	var labelStr string
	for _, label := range labels {
		labelStr += label.String() + ","
	}
	if len(labelStr) > 0 {
		labelStr = labelStr[:len(labelStr)-1]
	}
	return name + "{" + labelStr + "}"
}
