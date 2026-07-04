package osrm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/abubakar508/dispatch-ops/internal/models"
)

var (
	ErrServiceUnavailable = errors.New("osrm: routing service unavailable")
	ErrNoRoute            = errors.New("osrm: no route could be found between the given points")
	ErrInvalidResponse    = errors.New("osrm: received an invalid response")
)

type TableResult struct {
	Durations [][]float64
	Distances [][]float64
}

type RouteResult struct {
	Geometry [][]float64
	Metrics  models.RouteMetrics
}

type Client interface {
	Table(ctx context.Context, points []models.Coordinate) (TableResult, error)
	Route(ctx context.Context, points []models.Coordinate) (RouteResult, error)
}

type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPClient(baseURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: timeout},
	}
}

type tableResponse struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Durations [][]float64 `json:"durations"`
	Distances [][]float64 `json:"distances"`
}

type routeResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Routes  []struct {
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
		Geometry struct {
			Coordinates [][]float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"routes"`
}

func (c *HTTPClient) Table(ctx context.Context, points []models.Coordinate) (TableResult, error) {
	if len(points) < 2 {
		return TableResult{}, fmt.Errorf("osrm: table requires at least two points, got %d", len(points))
	}

	endpoint := fmt.Sprintf("%s/table/v1/driving/%s", c.baseURL, joinCoordinates(points))
	q := url.Values{}
	q.Set("annotations", "duration,distance")

	var resp tableResponse
	if err := c.do(ctx, endpoint, q, &resp); err != nil {
		return TableResult{}, err
	}

	if resp.Code != "Ok" {
		return TableResult{}, fmt.Errorf("%w: %s %s", ErrInvalidResponse, resp.Code, resp.Message)
	}
	if len(resp.Durations) == 0 {
		return TableResult{}, fmt.Errorf("%w: empty duration matrix", ErrInvalidResponse)
	}

	return TableResult{Durations: resp.Durations, Distances: resp.Distances}, nil
}

func (c *HTTPClient) Route(ctx context.Context, points []models.Coordinate) (RouteResult, error) {
	if len(points) < 2 {
		return RouteResult{}, fmt.Errorf("osrm: route requires at least two points, got %d", len(points))
	}

	endpoint := fmt.Sprintf("%s/route/v1/driving/%s", c.baseURL, joinCoordinates(points))
	q := url.Values{}
	q.Set("overview", "full")
	q.Set("geometries", "geojson")

	var resp routeResponse
	if err := c.do(ctx, endpoint, q, &resp); err != nil {
		return RouteResult{}, err
	}

	if resp.Code != "Ok" {
		if resp.Code == "NoRoute" {
			return RouteResult{}, ErrNoRoute
		}
		return RouteResult{}, fmt.Errorf("%w: %s %s", ErrInvalidResponse, resp.Code, resp.Message)
	}
	if len(resp.Routes) == 0 {
		return RouteResult{}, ErrNoRoute
	}

	route := resp.Routes[0]
	return RouteResult{
		Geometry: route.Geometry.Coordinates,
		Metrics: models.RouteMetrics{
			DistanceMeters:  route.Distance,
			DurationSeconds: route.Duration,
		},
	}, nil
}

func (c *HTTPClient) do(ctx context.Context, endpoint string, q url.Values, target any) error {
	fullURL := endpoint + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("osrm: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "dispatch-ops/1.0")

	res, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return fmt.Errorf("osrm: request timed out: %w", err)
		}
		return fmt.Errorf("%w: %v", ErrServiceUnavailable, err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 500 {
		return fmt.Errorf("%w: upstream status %d", ErrServiceUnavailable, res.StatusCode)
	}
	if res.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("%w: rate limited by routing provider", ErrServiceUnavailable)
	}

	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return fmt.Errorf("%w: decode body: %v", ErrInvalidResponse, err)
	}
	return nil
}

func joinCoordinates(points []models.Coordinate) string {
	parts := make([]string, len(points))
	for i, p := range points {
		parts[i] = fmt.Sprintf("%f,%f", p.Lng, p.Lat)
	}
	return strings.Join(parts, ";")
}
