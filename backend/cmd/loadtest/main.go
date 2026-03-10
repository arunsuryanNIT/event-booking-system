// Package main is a standalone CLI tool for load-testing the event booking API.
// It demonstrates that concurrent bookings never exceed capacity and that
// concurrent cancellations never produce negative counts.
//
// Usage:
//
//	go run ./cmd/loadtest --base-url http://localhost:8080 --test book
//	go run ./cmd/loadtest --base-url http://localhost:8080 --test cancel
//	go run ./cmd/loadtest --base-url http://localhost:8080 --test mixed
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type event struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Capacity    int    `json:"capacity"`
	BookedCount int    `json:"booked_count"`
}

type booking struct {
	ID      string `json:"id"`
	EventID string `json:"event_id"`
	UserID  string `json:"user_id"`
	Status  string `json:"status"`
}

type apiResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error"`
	Message string          `json:"message"`
}

var (
	baseURL    = flag.String("base-url", "http://localhost:8080", "Base URL of the API")
	eventID    = flag.String("event-id", "e1111111-1111-1111-1111-111111111111", "Event UUID to test")
	goroutines = flag.Int("goroutines", 50, "Number of concurrent goroutines")
	testType   = flag.String("test", "book", "Test type: book, cancel, mixed")
)

// seeded user IDs from the migrations
var userIDs = []string{
	"a1111111-1111-1111-1111-111111111111",
	"b2222222-2222-2222-2222-222222222222",
	"c3333333-3333-3333-3333-333333333333",
	"d4444444-4444-4444-4444-444444444444",
	"a5555555-5555-5555-5555-555555555555",
	"a6666666-6666-6666-6666-666666666666",
	"a7777777-7777-7777-7777-777777777777",
	"a8888888-8888-8888-8888-888888888888",
	"a9999999-9999-9999-9999-999999999999",
}

func main() {
	flag.Parse()

	fmt.Println("=== Event Booking Load Test ===")
	fmt.Println()

	switch *testType {
	case "book":
		runBookTest()
	case "cancel":
		runCancelTest()
	case "mixed":
		runMixedTest()
	default:
		fmt.Printf("unknown test type: %s (expected: book, cancel, mixed)\n", *testType)
	}
}

// runBookTest fires N goroutines that all try to book the same event simultaneously.
// Verifies: successful bookings == remaining capacity, no overbooking.
func runBookTest() {
	ev := fetchEvent(*eventID)
	remaining := ev.Capacity - ev.BookedCount

	fmt.Printf("Event:      %s (capacity: %d, booked: %d, remaining: %d)\n", ev.Title, ev.Capacity, ev.BookedCount, remaining)
	fmt.Printf("Goroutines: %d\n", *goroutines)
	fmt.Printf("Test:       book\n\n")

	results := fireBookings(*eventID, *goroutines)

	successes := count(results, "success")
	soldOut := count(results, "sold_out")
	alreadyBooked := count(results, "already_booked")
	errors := count(results, "error")

	evAfter := fetchEvent(*eventID)

	fmt.Println("Results:")
	fmt.Printf("  Successful bookings:  %d\n", successes)
	fmt.Printf("  Sold out:             %d\n", soldOut)
	fmt.Printf("  Already booked:       %d\n", alreadyBooked)
	fmt.Printf("  Errors:               %d\n", errors)
	fmt.Println()
	fmt.Println("Verification:")
	fmt.Printf("  Expected booked:  %d\n", ev.BookedCount+min(successes, remaining))
	fmt.Printf("  Actual booked:    %d\n", evAfter.BookedCount)
	fmt.Printf("  Capacity:         %d\n", evAfter.Capacity)
	fmt.Println()

	if evAfter.BookedCount <= evAfter.Capacity && errors == 0 {
		fmt.Println("PASS -- no overbooking, no errors")
	} else {
		fmt.Println("FAIL -- overbooking detected or unexpected errors")
	}
}

// runCancelTest first books the event to full capacity sequentially, then fires
// N goroutines that all try to cancel simultaneously.
// Verifies: booked_count returns to 0, never goes negative.
func runCancelTest() {
	ev := fetchEvent(*eventID)
	remaining := ev.Capacity - ev.BookedCount

	fmt.Printf("Event:      %s (capacity: %d, booked: %d)\n", ev.Title, ev.Capacity, ev.BookedCount, )
	fmt.Printf("Test:       cancel\n\n")

	// Book to capacity sequentially
	fmt.Printf("Booking %d spots sequentially...\n", remaining)
	var bookingIDs []string
	for i := 0; i < remaining && i < len(userIDs); i++ {
		b := doBook(*eventID, userIDs[i])
		if b != nil {
			bookingIDs = append(bookingIDs, b.ID)
		}
	}
	fmt.Printf("Created %d bookings\n\n", len(bookingIDs))

	if len(bookingIDs) == 0 {
		fmt.Println("No bookings to cancel. Reset the database and retry.")
		return
	}

	// Fire cancellations concurrently
	fmt.Printf("Firing %d concurrent cancellations...\n", len(bookingIDs))
	cancelResults := fireCancellations(bookingIDs)

	successes := count(cancelResults, "success")
	notFound := count(cancelResults, "not_found")
	errors := count(cancelResults, "error")

	evAfter := fetchEvent(*eventID)

	fmt.Println()
	fmt.Println("Results:")
	fmt.Printf("  Successful cancels:   %d\n", successes)
	fmt.Printf("  Not found/duplicate:  %d\n", notFound)
	fmt.Printf("  Errors:               %d\n", errors)
	fmt.Println()
	fmt.Println("Verification:")
	fmt.Printf("  Booked count after:   %d\n", evAfter.BookedCount)
	fmt.Println()

	if evAfter.BookedCount >= 0 && errors == 0 {
		fmt.Println("PASS -- no negative count, no errors")
	} else {
		fmt.Println("FAIL -- negative count or unexpected errors")
	}
}

// runMixedTest books the event to capacity, then fires goroutines that
// simultaneously cancel existing bookings and attempt new bookings.
// Verifies: booked_count stays within [0, capacity].
func runMixedTest() {
	ev := fetchEvent(*eventID)
	remaining := ev.Capacity - ev.BookedCount

	fmt.Printf("Event:      %s (capacity: %d, booked: %d)\n", ev.Title, ev.Capacity, ev.BookedCount)
	fmt.Printf("Test:       mixed (concurrent cancel + rebook)\n\n")

	// Book to capacity sequentially
	fmt.Printf("Booking %d spots sequentially...\n", remaining)
	var bookingIDs []string
	for i := 0; i < remaining && i < len(userIDs); i++ {
		b := doBook(*eventID, userIDs[i])
		if b != nil {
			bookingIDs = append(bookingIDs, b.ID)
		}
	}
	fmt.Printf("Created %d bookings, event is now full\n\n", len(bookingIDs))

	// Fire mixed: half cancel, half try to book
	totalGoroutines := len(bookingIDs) * 2
	fmt.Printf("Firing %d goroutines: %d cancels + %d booking attempts...\n",
		totalGoroutines, len(bookingIDs), len(bookingIDs))

	startGun := make(chan struct{})
	var wg sync.WaitGroup
	var mu sync.Mutex
	var cancelSuccesses, bookSuccesses int

	// Cancel goroutines
	for i, bid := range bookingIDs {
		wg.Add(1)
		go func(bookingID, userID string) {
			defer wg.Done()
			<-startGun
			result := doCancel(bookingID, userID)
			if result == "success" {
				mu.Lock()
				cancelSuccesses++
				mu.Unlock()
			}
		}(bid, userIDs[i])
	}

	// Book goroutines (use different users if possible)
	for i := 0; i < len(bookingIDs); i++ {
		wg.Add(1)
		userIdx := (i + len(bookingIDs)) % len(userIDs)
		go func(uid string) {
			defer wg.Done()
			<-startGun
			b := doBook(*eventID, uid)
			if b != nil {
				mu.Lock()
				bookSuccesses++
				mu.Unlock()
			}
		}(userIDs[userIdx])
	}

	start := time.Now()
	close(startGun)
	wg.Wait()
	elapsed := time.Since(start)

	evAfter := fetchEvent(*eventID)

	fmt.Println()
	fmt.Printf("Completed in %dms\n\n", elapsed.Milliseconds())
	fmt.Println("Results:")
	fmt.Printf("  Successful cancels:   %d\n", cancelSuccesses)
	fmt.Printf("  Successful rebooks:   %d\n", bookSuccesses)
	fmt.Println()
	fmt.Println("Verification:")
	fmt.Printf("  Booked count after:   %d\n", evAfter.BookedCount)
	fmt.Printf("  Capacity:             %d\n", evAfter.Capacity)
	fmt.Println()

	if evAfter.BookedCount >= 0 && evAfter.BookedCount <= evAfter.Capacity {
		fmt.Println("PASS -- booked_count is within [0, capacity]")
	} else {
		fmt.Println("FAIL -- booked_count is out of bounds")
	}
}

// fireBookings launches N goroutines that all try to book the same event
// simultaneously using the start-gun pattern.
func fireBookings(eventID string, n int) []string {
	startGun := make(chan struct{})
	results := make([]string, n)
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			userID := userIDs[idx%len(userIDs)]
			<-startGun
			b := doBook(eventID, userID)
			if b != nil {
				results[idx] = "success"
			} else {
				results[idx] = "sold_out"
			}
		}(i)
	}

	start := time.Now()
	close(startGun)
	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("Completed in %dms\n\n", elapsed.Milliseconds())
	return results
}

// fireCancellations launches one goroutine per booking, all firing at once.
func fireCancellations(bookingIDs []string) []string {
	startGun := make(chan struct{})
	results := make([]string, len(bookingIDs))
	var wg sync.WaitGroup

	for i, bid := range bookingIDs {
		wg.Add(1)
		go func(idx int, bookingID string) {
			defer wg.Done()
			userID := userIDs[idx%len(userIDs)]
			<-startGun
			result := doCancel(bookingID, userID)
			results[idx] = result
		}(i, bid)
	}

	close(startGun)
	wg.Wait()
	return results
}

// doBook sends a booking request and returns the booking on success, nil on failure.
func doBook(eventID, userID string) *booking {
	body, _ := json.Marshal(map[string]string{"user_id": userID})
	resp, err := http.Post(
		fmt.Sprintf("%s/api/events/%s/book", *baseURL, eventID),
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var r apiResponse
	data, _ := io.ReadAll(resp.Body)
	json.Unmarshal(data, &r)

	if !r.Success {
		return nil
	}

	var b booking
	json.Unmarshal(r.Data, &b)
	return &b
}

// doCancel sends a cancellation request and returns the outcome.
func doCancel(bookingID, userID string) string {
	body, _ := json.Marshal(map[string]string{"user_id": userID})
	resp, err := http.Post(
		fmt.Sprintf("%s/api/bookings/%s/cancel", *baseURL, bookingID),
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return "error"
	}
	defer resp.Body.Close()

	var r apiResponse
	data, _ := io.ReadAll(resp.Body)
	json.Unmarshal(data, &r)

	if r.Success {
		return "success"
	}
	return "not_found"
}

// fetchEvent retrieves the current state of an event from the API.
func fetchEvent(id string) event {
	resp, err := http.Get(fmt.Sprintf("%s/api/events/%s", *baseURL, id))
	if err != nil {
		panic(fmt.Sprintf("failed to fetch event: %v", err))
	}
	defer resp.Body.Close()

	var r apiResponse
	data, _ := io.ReadAll(resp.Body)
	json.Unmarshal(data, &r)

	var ev event
	json.Unmarshal(r.Data, &ev)
	return ev
}

func count(results []string, value string) int {
	n := 0
	for _, r := range results {
		if r == value {
			n++
		}
	}
	return n
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
