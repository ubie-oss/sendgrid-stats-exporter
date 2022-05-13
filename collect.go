package main

import (
	"context"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/ubie-oss/sendgrid-stats-exporter/sendgrid"
)

type Collector struct {
	logger     log.Logger
	client     *sendgrid.Client
	categories []string

	blocks           *prometheus.Desc
	bounceDrops      *prometheus.Desc
	bounces          *prometheus.Desc
	clicks           *prometheus.Desc
	deferred         *prometheus.Desc
	delivered        *prometheus.Desc
	invalidEmails    *prometheus.Desc
	opens            *prometheus.Desc
	processed        *prometheus.Desc
	requests         *prometheus.Desc
	spamReportDrops  *prometheus.Desc
	spamReports      *prometheus.Desc
	uniqueClicks     *prometheus.Desc
	uniqueOpens      *prometheus.Desc
	unsubscribeDrops *prometheus.Desc
	unsubscribes     *prometheus.Desc
}

func NewCollector(logger log.Logger, client *sendgrid.Client, categories []string) *Collector {
	return &Collector{
		logger:     logger,
		client:     client,
		categories: categories,

		blocks: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "blocks"),
			"blocks",
			[]string{"user_name", "category"},
			nil,
		),
		bounceDrops: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "bounce_drops"),
			"bounce_drops",
			[]string{"user_name", "category"},
			nil,
		),
		bounces: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "bounces"),
			"bounces",
			[]string{"user_name", "category"},
			nil,
		),
		clicks: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "clicks"),
			"clicks",
			[]string{"user_name", "category"},
			nil,
		),
		deferred: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "deferred"),
			"deferred",
			[]string{"user_name", "category"},
			nil,
		),
		delivered: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "delivered"),
			"delivered",
			[]string{"user_name", "category"},
			nil,
		),
		invalidEmails: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "invalid_emails"),
			"invalid_emails",
			[]string{"user_name", "category"},
			nil,
		),
		opens: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "opens"),
			"opens",
			[]string{"user_name", "category"},
			nil,
		),
		processed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "processed"),
			"processed",
			[]string{"user_name", "category"},
			nil,
		),
		requests: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "requests"),
			"requests",
			[]string{"user_name", "category"},
			nil,
		),
		spamReportDrops: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "spam_report_drops"),
			"spam_report_drops",
			[]string{"user_name", "category"},
			nil,
		),
		spamReports: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "spam_reports"),
			"spam_reports",
			[]string{"user_name", "category"},
			nil,
		),
		uniqueClicks: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "unique_clicks"),
			"unique_clicks",
			[]string{"user_name", "category"},
			nil,
		),
		uniqueOpens: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "unique_opens"),
			"unique_opens",
			[]string{"user_name", "category"},
			nil,
		),
		unsubscribeDrops: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "unsubscribe_drops"),
			"unsubscribe_drops",
			[]string{"user_name", "category"},
			nil,
		),
		unsubscribes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "unsubscribes"),
			"unsubscribes",
			[]string{"user_name", "category"},
			nil,
		),
	}
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	var today time.Time

	if *location != "" && *timeOffset != 0 {
		loc := time.FixedZone(*location, *timeOffset)
		today = time.Now().In(loc)
	} else {
		today = time.Now()
	}
	todayStr := today.Format("2006-01-02")

	ctx := context.Background()

	if err := c.collectGlobalStats(ctx, ch, todayStr); err != nil {
		level.Error(c.logger).Log(err)
	}

	if len(c.categories) > 0 {
		if err := c.collectCategoryStats(ctx, ch, todayStr); err != nil {
			level.Error(c.logger).Log(err)
		}
	}
}

func (c *Collector) collectGlobalStats(ctx context.Context, ch chan<- prometheus.Metric, dateStr string) error {
	globalStats, err := c.client.GetGlobalStats(ctx, &sendgrid.GetGlobalStatsArguments{
		StartDate:    dateStr,
		EndDate:      dateStr,
		AggregatedBy: "day",
	})
	if err != nil {
		return err
	}

	for _, stats := range globalStats[0].Stats {
		category := ""
		c.collectMetrics(ch, stats.Metrics, &category)
	}

	return nil
}

func (c *Collector) collectCategoryStats(ctx context.Context, ch chan<- prometheus.Metric, dateStr string) error {
	categoryStats, err := c.client.GetCategoryStats(ctx, &sendgrid.GetCategoryStatsArguments{
		StartDate:    dateStr,
		EndDate:      dateStr,
		AggregatedBy: "day",
		Categories:   c.categories,
	})
	if err != nil {
		return err
	}

	for _, stats := range categoryStats[0].Stats {
		c.collectMetrics(ch, stats.Metrics, &stats.Name)
	}

	return nil
}

func (c *Collector) collectMetrics(ch chan<- prometheus.Metric, metric sendgrid.StatMetric, category *string) {
	ch <- prometheus.MustNewConstMetric(
		c.blocks,
		prometheus.GaugeValue,
		float64(metric.Blocks),
		*sendGridUserName,
		*category,
	)
	ch <- prometheus.MustNewConstMetric(
		c.bounceDrops,
		prometheus.GaugeValue,
		float64(metric.BounceDrops),
		*sendGridUserName,
		*category,
	)
	ch <- prometheus.MustNewConstMetric(
		c.bounces,
		prometheus.GaugeValue,
		float64(metric.Bounces),
		*sendGridUserName,
		*category,
	)
	ch <- prometheus.MustNewConstMetric(
		c.clicks,
		prometheus.GaugeValue,
		float64(metric.Clicks),
		*sendGridUserName,
		*category,
	)
	ch <- prometheus.MustNewConstMetric(
		c.deferred,
		prometheus.GaugeValue,
		float64(metric.Deferred),
		*sendGridUserName,
		*category,
	)
	ch <- prometheus.MustNewConstMetric(
		c.delivered,
		prometheus.GaugeValue,
		float64(metric.Delivered),
		*sendGridUserName,
		*category,
	)
	ch <- prometheus.MustNewConstMetric(
		c.invalidEmails,
		prometheus.GaugeValue,
		float64(metric.InvalidEmails),
		*sendGridUserName,
		*category,
	)
	ch <- prometheus.MustNewConstMetric(
		c.opens,
		prometheus.GaugeValue,
		float64(metric.Opens),
		*sendGridUserName,
		*category,
	)
	ch <- prometheus.MustNewConstMetric(
		c.processed,
		prometheus.GaugeValue,
		float64(metric.Processed),
		*sendGridUserName,
		*category,
	)
	ch <- prometheus.MustNewConstMetric(
		c.requests,
		prometheus.GaugeValue,
		float64(metric.Requests),
		*sendGridUserName,
		*category,
	)
	ch <- prometheus.MustNewConstMetric(
		c.spamReportDrops,
		prometheus.GaugeValue,
		float64(metric.SpamReportDrops),
		*sendGridUserName,
		*category,
	)
	ch <- prometheus.MustNewConstMetric(
		c.spamReports,
		prometheus.GaugeValue,
		float64(metric.SpamReports),
		*sendGridUserName,
		*category,
	)
	ch <- prometheus.MustNewConstMetric(
		c.uniqueClicks,
		prometheus.GaugeValue,
		float64(metric.UniqueClicks),
		*sendGridUserName,
		*category,
	)
	ch <- prometheus.MustNewConstMetric(
		c.uniqueOpens,
		prometheus.GaugeValue,
		float64(metric.UniqueOpens),
		*sendGridUserName,
		*category,
	)
	ch <- prometheus.MustNewConstMetric(
		c.unsubscribeDrops,
		prometheus.GaugeValue,
		float64(metric.UnsubscribeDrops),
		*sendGridUserName,
		*category,
	)
	ch <- prometheus.MustNewConstMetric(
		c.unsubscribes,
		prometheus.GaugeValue,
		float64(metric.Unsubscribes),
		*sendGridUserName,
		*category,
	)
}
