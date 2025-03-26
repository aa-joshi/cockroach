package aggmetric

import "testing"

func TestMetricTracker(t *testing.T) {
	registerMetric("hi", &childSet{})
}
