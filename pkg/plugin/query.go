package plugin

import (
	"fmt"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/VictoriaMetrics/victorialogs-datasource/pkg/utils"
)

const (
	instantQueryPath = "/select/logsql/query"
	tailQueryPath    = "/select/logsql/tail"
	defaultMaxLines  = 1000
)

// Query represents backend query object
type Query struct {
	RefID        string `json:"refId"`
	Expr         string `json:"expr"`
	LegendFormat string `json:"legendFormat"`
	MaxLines     int    `json:"maxLines"`
	TimeRange    TimeRange
	IntervalMs   int `json:"intervalMs"`
	url          *url.URL
}

// TimeRange represents time range backend object
type TimeRange struct {
	From time.Time
	To   time.Time
}

// GetQueryURL calculates step and clear expression from template variables,
// and after builds query url depends on query type
func (q *Query) getQueryURL(rawURL string, queryParams string) (string, error) {
	if rawURL == "" {
		return "", fmt.Errorf("url can't be blank")
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse datasource url: %s", err)
	}
	params, err := url.ParseQuery(queryParams)
	if err != nil {
		return "", fmt.Errorf("failed to parse query params: %s", err.Error())
	}

	q.url = u

	return q.queryInstantURL(params), nil
}

// queryInstantURL prepare query url for instant query
func (q *Query) queryTailURL(rawURL string, queryParams string) (string, error) {
	if rawURL == "" {
		return "", fmt.Errorf("url can't be blank")
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse datasource url: %s", err)
	}
	params, err := url.ParseQuery(queryParams)
	if err != nil {
		return "", fmt.Errorf("failed to parse query params: %s", err.Error())
	}

	q.url = u

	q.url.Path = path.Join(q.url.Path, tailQueryPath)
	values := q.url.Query()

	for k, vl := range params {
		for _, v := range vl {
			values.Add(k, v)
		}
	}

	q.Expr = utils.ReplaceTemplateVariable(q.Expr, q.IntervalMs)
	values.Set("query", q.Expr)

	q.url.RawQuery = values.Encode()
	return q.url.String(), nil
}

// queryInstantURL prepare query url for instant query
func (q *Query) queryInstantURL(queryParams url.Values) string {
	q.url.Path = path.Join(q.url.Path, instantQueryPath)
	values := q.url.Query()

	for k, vl := range queryParams {
		for _, v := range vl {
			values.Add(k, v)
		}
	}

	if q.MaxLines <= 0 {
		q.MaxLines = defaultMaxLines
	}

	now := time.Now()
	if q.TimeRange.From.IsZero() {
		q.TimeRange.From = now.Add(-time.Minute * 5)
	}
	if q.TimeRange.To.IsZero() {
		q.TimeRange.To = now
	}

	q.Expr = utils.ReplaceTemplateVariable(q.Expr, q.IntervalMs)
	values.Set("query", q.Expr)
	values.Set("limit", strconv.Itoa(q.MaxLines))
	values.Set("start", strconv.FormatInt(q.TimeRange.From.Unix(), 10))
	values.Set("end", strconv.FormatInt(q.TimeRange.To.Unix(), 10))

	q.url.RawQuery = values.Encode()
	return q.url.String()
}
