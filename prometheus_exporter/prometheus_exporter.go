package prometheus_exporter

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"gitlab.seznam.net/sklik-devops/slo-exporter/pkg/slo_event_producer"
)

var (
	component string
	log       *logrus.Entry
)

func init() {
	const component = "prometheus_exporter"
	log = logrus.WithField("component", component)
}

type PrometheusSloEventExporter struct {
	counterVec  *prometheus.CounterVec
	knownLabels []string
}

func New(labels []string) *PrometheusSloEventExporter {
	return &PrometheusSloEventExporter{
		prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "slo_events_total",
			Help:        "Total number of SLO events exported with it's result and metadata.",
			ConstLabels: nil,
		}, labels),
		labels,
	}
}

func (e *PrometheusSloEventExporter) Run(ctx context.Context, input <-chan *slo_event_producer.SloEvent) {
	prometheus.MustRegister(e.counterVec)

	go func() {
		defer log.Info("stopping...")
		for {
			select {
			case event, ok := <-input:
				if !ok {
					log.Info("input channel closed, finishing")
				}
				e.processEvent(event)
			}
		}
	}()
}

// make sure that eventMetadata contains exactly the expected set, so that it passed Prometheus library sanity checks
func normalizeEventMetadata(knownMetadata []string, eventMetadata map[string]string) map[string]string {
	normalized := make(map[string]string)
	for _, k := range knownMetadata {
		v, _ := eventMetadata[k]
		normalized[k] = v
	}
	return normalized
}

func (e *PrometheusSloEventExporter) processEvent(event *slo_event_producer.SloEvent) {
	e.counterVec.With(prometheus.Labels(normalizeEventMetadata(e.knownLabels, event.SloMetadata))).Inc()
}