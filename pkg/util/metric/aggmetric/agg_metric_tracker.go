package aggmetric

import (
	"context"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/settings"
)

var AppNameLabelEnabled = settings.RegisterBoolSetting(
	settings.ApplicationLevel,
	"sql.application_name_metrics.enabled",
	"set to true to enable 'app' label in SQL metrics",
	false, /* default */
	settings.WithPublic)

var DBNameLabelEnabled = settings.RegisterBoolSetting(
	settings.ApplicationLevel,
	"sql.db_name_metrics.enabled",
	"set to true to enable 'db' label in SQL metrics",
	false, /* default */
	settings.WithPublic)

// metricTracker tracks all registered metrics with CacheStorage.
var metricTracker = struct {
	sync.Once
	mu      sync.Mutex
	metrics map[string]*childSet
}{}

func InitMetricTracker(sv *settings.Values) {
	metricTracker.Do(func() {
		metricTracker.metrics = make(map[string]*childSet)
	})

	AppNameLabelEnabled.SetOnChange(sv, func(_ context.Context) {
		reinitialiseMetrics(sv)
	})

	DBNameLabelEnabled.SetOnChange(sv, func(_ context.Context) {
		reinitialiseMetrics(sv)
	})

}

// registerMetric registers a metric with the tracker.
func registerMetric(name string, cs *childSet) {

	metricTracker.mu.Lock()
	defer metricTracker.mu.Unlock()
	metricTracker.metrics[name] = cs
}

func reinitialiseMetrics(sv *settings.Values) {
	metricTracker.mu.Lock()
	defer metricTracker.mu.Unlock()

	labelSet := []string{}

	if DBNameLabelEnabled.Get(sv) {
		labelSet = append(labelSet, "db_name")
	}

	if AppNameLabelEnabled.Get(sv) {
		labelSet = append(labelSet, "app_name")
	}

	for _, cs := range metricTracker.metrics {
		cs.reinitialise(labelSet...)
	}
}
