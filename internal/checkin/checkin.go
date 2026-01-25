package checkin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"southwest-bot/internal/models"
)

const (
	baseURL   = "https://mobile.southwest.com/api/mobile-air-operations/v1/mobile-air-operations/page/check-in"
	userAgent = "Mozilla/5.0 (Linux; Android 10; SM-G960U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.120 Mobile Safari/537.36"
)

type Client struct {
	httpClient *http.Client
	verbose    bool
}

func New(verbose bool) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		verbose: verbose,
	}
}

func (c *Client) CheckIn(reservation models.Reservation) ([]models.BoardingPass, error) {
	// Step 1: Look up the reservation
	if c.verbose {
		fmt.Println("Looking up reservation...")
	}

	lookupResp, err := c.lookupReservation(reservation)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup reservation: %w", err)
	}

	if c.verbose {
		fmt.Println("Reservation found, attempting check-in...")
	}

	// Step 2: Perform check-in
	confirmResp, err := c.performCheckIn(lookupResp)
	if err != nil {
		return nil, fmt.Errorf("failed to check in: %w", err)
	}

	// Step 3: Extract boarding passes
	boardingPasses := c.extractBoardingPasses(confirmResp)

	return boardingPasses, nil
}

func (c *Client) lookupReservation(reservation models.Reservation) (*models.ReservationLookupResponse, error) {
	url := fmt.Sprintf("%s/%s?first-name=%s&last-name=%s",
		baseURL,
		strings.ToUpper(reservation.ConfirmationNumber),
		reservation.FirstName,
		reservation.LastName,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)

	resp, err := c.doWithRetry(req, 3)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if c.verbose {
		fmt.Printf("Lookup response status: %d\n", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lookup failed with status %d: %s", resp.StatusCode, string(body))
	}

	var lookupResp models.ReservationLookupResponse
	if err := json.Unmarshal(body, &lookupResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &lookupResp, nil
}

func (c *Client) performCheckIn(lookup *models.ReservationLookupResponse) (*models.CheckInConfirmationResponse, error) {
	checkInLink := lookup.CheckInViewReservationPage.Links.CheckIn
	url := "https://mobile.southwest.com" + checkInLink.Href
	bodyBytes, err := json.Marshal(checkInLink.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal check-in body: %w", err)
	}

	req, err := http.NewRequest(checkInLink.Method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doWithRetry(req, 3)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if c.verbose {
		fmt.Printf("Check-in response status: %d\n", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("check-in failed with status %d: %s", resp.StatusCode, string(body))
	}

	var confirmResp models.CheckInConfirmationResponse
	if err := json.Unmarshal(body, &confirmResp); err != nil {
		return nil, fmt.Errorf("failed to parse check-in response: %w", err)
	}

	return &confirmResp, nil
}

func (c *Client) extractBoardingPasses(confirm *models.CheckInConfirmationResponse) []models.BoardingPass {
	var passes []models.BoardingPass

	for _, flight := range confirm.CheckInConfirmationPage.Flights {
		for _, passenger := range flight.Passengers {
			pass := models.BoardingPass{
				FirstName:        extractFirstName(passenger.Name),
				LastName:         extractLastName(passenger.Name),
				BoardingGroup:    passenger.BoardingGroup,
				BoardingPosition: passenger.BoardingPosition,
				FlightNumber:     flight.FlightNumber,
				DepartureAirport: flight.OriginAirportCode,
				ArrivalAirport:   flight.DestinationAirportCode,
				DepartureTime:    flight.DepartureTime,
			}
			passes = append(passes, pass)
		}
	}

	return passes
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("X-API-Key", "l7xx944d175ea25f4b9c903a583ea82a1c4c")
	req.Header.Set("X-Channel-ID", "MWEB")
}

func (c *Client) doWithRetry(req *http.Request, maxRetries int) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			if c.verbose {
				fmt.Printf("Retrying in %v (attempt %d/%d)...\n", backoff, attempt+1, maxRetries)
			}
			time.Sleep(backoff)

			// Clone the request for retry (body needs to be re-readable)
			newReq, err := http.NewRequest(req.Method, req.URL.String(), nil)
			if err != nil {
				lastErr = err
				continue
			}
			newReq.Header = req.Header
			if req.Body != nil {
				body, _ := io.ReadAll(req.Body)
				newReq.Body = io.NopCloser(bytes.NewReader(body))
				req.Body = io.NopCloser(bytes.NewReader(body))
			}
			req = newReq
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		// Retry on 5xx errors
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func extractFirstName(fullName string) string {
	parts := strings.Fields(fullName)
	if len(parts) > 0 {
		return parts[0]
	}
	return fullName
}

func extractLastName(fullName string) string {
	parts := strings.Fields(fullName)
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}
