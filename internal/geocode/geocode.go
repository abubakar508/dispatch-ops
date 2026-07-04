package geocode

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/abubakar508/dispatch-ops/internal/models"
)

var (
	ErrServiceUnavailable = errors.New("geocode: search service unavailable")
	ErrEmptyQuery         = errors.New("geocode: query must not be empty")
)

type Geocoder interface {
	Search(ctx context.Context, query string, limit int) ([]models.Place, error)
}

type NominatimClient struct {
	baseURL    string
	userAgent  string
	viewbox    string
	httpClient *http.Client
}

func NewNominatimClient(baseURL, userAgent string, timeout time.Duration) *NominatimClient {
	return &NominatimClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		userAgent:  userAgent,
		viewbox:    "36.6509,-1.1636,37.1050,-1.4448",
		httpClient: &http.Client{Timeout: timeout},
	}
}

type nominatimResult struct {
	DisplayName string `json:"display_name"`
	Name        string `json:"name"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	Category    string `json:"category"`
	Type        string `json:"type"`
}

func (c *NominatimClient) Search(ctx context.Context, query string, limit int) ([]models.Place, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, ErrEmptyQuery
	}
	if limit <= 0 || limit > 10 {
		limit = 6
	}

	q := url.Values{}
	q.Set("q", query)
	q.Set("format", "jsonv2")
	q.Set("addressdetails", "0")
	q.Set("limit", strconv.Itoa(limit))
	q.Set("viewbox", c.viewbox)
	q.Set("countrycodes", "ke")

	endpoint := fmt.Sprintf("%s/search?%s", c.baseURL, q.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("geocode: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	res, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("geocode: request timed out: %w", err)
		}
		return nil, fmt.Errorf("%w: %v", ErrServiceUnavailable, err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 500 || res.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("%w: upstream status %d", ErrServiceUnavailable, res.StatusCode)
	}

	var raw []nominatimResult
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("%w: decode body: %v", ErrServiceUnavailable, err)
	}

	places := make([]models.Place, 0, len(raw))
	for _, r := range raw {
		lat, errLat := strconv.ParseFloat(r.Lat, 64)
		lng, errLng := strconv.ParseFloat(r.Lon, 64)
		if errLat != nil || errLng != nil {
			continue
		}
		category := r.Category
		if category == "" {
			category = r.Type
		}
		places = append(places, models.Place{
			DisplayName: r.DisplayName,
			Name:        placeName(r),
			Coordinate:  models.Coordinate{Lat: lat, Lng: lng},
			Category:    category,
		})
	}
	return places, nil
}

func placeName(r nominatimResult) string {
	if r.Name != "" {
		return r.Name
	}
	if idx := strings.Index(r.DisplayName, ","); idx > 0 {
		return r.DisplayName[:idx]
	}
	return r.DisplayName
}
