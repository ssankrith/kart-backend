package observability

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/ssankrith/kart-backend/internal/domain"
	"github.com/ssankrith/kart-backend/internal/promo"
)

// InstrumentPromo wraps a PromoChecker with Prometheus metrics for validation outcomes and latency.
func InstrumentPromo(inner domain.PromoChecker) domain.PromoChecker {
	return &instrumentedPromo{inner: inner}
}

type instrumentedPromo struct {
	inner domain.PromoChecker
}

var (
	promoValidations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "kart",
			Subsystem: "promo",
			Name:      "validations_total",
			Help:      "Promo Valid calls by coarse outcome",
		},
		[]string{"result"},
	)
	promoValidationSeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "kart",
			Subsystem: "promo",
			Name:      "validation_duration_seconds",
			Help:      "Wall time for a single Valid call (includes prelude check)",
			Buckets:   prometheus.DefBuckets,
		},
	)
)

func (p *instrumentedPromo) Valid(code string) bool {
	start := time.Now()
	defer func() {
		promoValidationSeconds.Observe(time.Since(start).Seconds())
	}()

	if !promo.CouponCodePreludeOK(code) {
		promoValidations.WithLabelValues("prelude_reject").Inc()
		return false
	}
	ok := p.inner.Valid(code)
	if ok {
		promoValidations.WithLabelValues("hit").Inc()
	} else {
		promoValidations.WithLabelValues("miss").Inc()
	}
	return ok
}

func (p *instrumentedPromo) Close() error {
	return p.inner.Close()
}
