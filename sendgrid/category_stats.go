package sendgrid

import (
	"context"
	"errors"
)

type CategoryStatMetrics struct {
	Metrics StatMetric `json:"metrics,omitempty"`
	Name    string     `json:"name"`
	Type    string     `json:"type"`
}

type CategoryStat struct {
	Date  string                `json:"date,omitempty"`
	Stats []CategoryStatMetrics `json:"stats,omitempty"`
}

type GetCategoryStatsArguments struct {
	StartDate    string
	EndDate      string
	AggregatedBy string
	Categories   []string
}

func (c *Client) GetCategoryStats(ctx context.Context, args *GetCategoryStatsArguments) ([]CategoryStat, error) {
	if args.StartDate == "" {
		return nil, errors.New("start_date is required")
	}

	if len(args.Categories) == 0 {
		return nil, errors.New("categories must not be empty")
	}

	params := requestParams{
		method:  "GET",
		subPath: "/categories/stats",
		queries: map[string]string{
			"start_date":    args.StartDate,
			"end_date":      args.EndDate,
			"aggregated_by": args.AggregatedBy,
		},
		arrayQueries: map[string][]string{
			"categories": args.Categories,
		},
	}

	var out []CategoryStat
	if err := c.doAPIRequest(ctx, &params, &out); err != nil {
		return nil, err
	}

	return out, nil
}
