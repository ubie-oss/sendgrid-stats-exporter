package sendgrid

import (
	"context"
	"errors"
)

type GlobalStatMetric struct {
	Blocks           int64 `json:"blocks,omitempty"`
	BounceDrops      int64 `json:"bounce_drops,omitempty"`
	Bounces          int64 `json:"bounces,omitempty"`
	Clicks           int64 `json:"clicks,omitempty"`
	Deferred         int64 `json:"deferred,omitempty"`
	Delivered        int64 `json:"delivered,omitempty"`
	InvalidEmails    int64 `json:"invalid_emails,omitempty"`
	Opens            int64 `json:"opens,omitempty"`
	Processed        int64 `json:"processed,omitempty"`
	Requests         int64 `json:"requests,omitempty"`
	SpamReportDrops  int64 `json:"spam_report_drops,omitempty"`
	SpamReports      int64 `json:"spam_reports,omitempty"`
	UniqueClicks     int64 `json:"unique_clicks,omitempty"`
	UniqueOpens      int64 `json:"unique_opens,omitempty"`
	UnsubscribeDrops int64 `json:"unsubscribe_drops,omitempty"`
	Unsubscribes     int64 `json:"unsubscribes,omitempty"`
}

type GlobalStatMetrics struct {
	Metrics GlobalStatMetric `json:"metrics,omitempty"`
}

type GlobalStat struct {
	Date  string              `json:"date,omitempty"`
	Stats []GlobalStatMetrics `json:"stats,omitempty"`
}

type GetGlobalStatsArguments struct {
	StartDate    string
	EndDate      string
	AggregatedBy string
}

func (c *Client) GetGlobalStats(ctx context.Context, args *GetGlobalStatsArguments) ([]GlobalStat, error) {
	if args.StartDate == "" {
		return nil, errors.New("start_date is required")
	}

	params := requestParams{
		method:  "GET",
		subPath: "/stats",
		queries: map[string]string{
			"start_date":    args.StartDate,
			"end_date":      args.EndDate,
			"aggregated_by": args.AggregatedBy,
		},
	}
	var out []GlobalStat
	if err := c.doAPIRequest(ctx, &params, &out); err != nil {
		return nil, err
	}

	return out, nil
}
