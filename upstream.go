package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const maxCalendarBytes = 10 << 20

var upstreamHTTPClient = &http.Client{
	Timeout: 15 * time.Second,
}

func fetchUpstreamCalendar(ctx context.Context, feedURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ical-filter-proxy/"+version)
	req.Header.Set("Accept", "text/calendar, text/plain;q=0.8, */*;q=0.1")

	resp, err := upstreamHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("upstream returned %s", resp.Status)
	}

	feedData, err := io.ReadAll(io.LimitReader(resp.Body, maxCalendarBytes+1))
	if err != nil {
		return nil, err
	}
	if len(feedData) > maxCalendarBytes {
		return nil, fmt.Errorf("upstream calendar exceeds %d bytes", maxCalendarBytes)
	}

	return feedData, nil
}
