package metrics

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/redis/go-redis/extra/redisprometheus/v9"
	"github.com/redis/go-redis/v9"
)

type Metrics interface {
	PrometheusRegisterer() prometheus.Registerer
	GetHTTPClientPrometheus() *HTTPClientPrometheusMetrics
	RegisterDB(db *sql.DB, _, _, dbName string) error
	RegisterRedis(client *redis.Client, serviceName, namespace string) error
}

type metrics struct {
	reg               prometheus.Registerer
	httpClientMetrics *HTTPClientPrometheusMetrics
}

func New() Metrics {
	reg := prometheus.DefaultRegisterer

	return &metrics{
		reg:               reg,
		httpClientMetrics: newHTTPClientPrometheusMetrics(reg),
	}
}

func (m *metrics) PrometheusRegisterer() prometheus.Registerer {
	return m.reg
}

func (m *metrics) GetHTTPClientPrometheus() *HTTPClientPrometheusMetrics {
	return m.httpClientMetrics
}

func (m *metrics) RegisterDB(db *sql.DB, _, _, dbName string) error {
	return m.reg.Register(collectors.NewDBStatsCollector(db, dbName))
}

func (m *metrics) RegisterRedis(client *redis.Client, serviceName, namespace string) error {
	return m.reg.Register(redisprometheus.NewCollector(BuildFQName(serviceName, namespace), "redis", client))
}
