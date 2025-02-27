package metric

import (
	"fmt"
	"testing"
)

func TestTransform(t *testing.T) {
	metricTracker := NewMetricTracker()

	MetaSelectDeltaExecuted := Metadata{
		Name:        "sql.select.delta.count",
		Help:        "delta of number of SQL SELECT statements successfully executed from previous window",
		Measurement: "delta of SQL Statements",
		Unit:        Unit_COUNT,
	}

	c := NewExportedCounterVec(MetaSelectDeltaExecuted, []string{"label1"}, AggregationTemporalityDelta)
	c.Inc(map[string]string{"label1": "value1"}, 2)

	metricArray := c.ToPrometheusMetrics()
	metric := metricArray[0]
	metricTracker.Transform(metric, "sql.select.delta.count")
	fmt.Println(metric)

	c.Inc(map[string]string{"label1": "value1"}, 3)
	metricArray = c.ToPrometheusMetrics()
	metric = metricArray[0]
	metricTracker.Transform(metric, "sql.select.delta.count")
	fmt.Println(metric)

	c.Inc(map[string]string{"label1": "value1"}, 4)
	metricArray = c.ToPrometheusMetrics()
	metric = metricArray[0]
	metricTracker.Transform(metric, "sql.select.delta.count")
	fmt.Println(metric)
}
