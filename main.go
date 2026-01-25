package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"southwest-bot/internal/checkin"
	"southwest-bot/internal/models"
	"southwest-bot/internal/scheduler"
)

func main() {
	// Define CLI flags
	confirmation := flag.String("c", "", "Southwest confirmation number (required)")
	confirmationLong := flag.String("confirmation", "", "Southwest confirmation number (required)")

	firstName := flag.String("f", "", "Passenger first name (required)")
	firstNameLong := flag.String("first", "", "Passenger first name (required)")

	lastName := flag.String("l", "", "Passenger last name (required)")
	lastNameLong := flag.String("last", "", "Passenger last name (required)")

	departure := flag.String("d", "", "Departure time (required). Format: 'YYYY-MM-DD HH:MM' or RFC3339")
	departureLong := flag.String("departure", "", "Departure time (required). Format: 'YYYY-MM-DD HH:MM' or RFC3339")

	verbose := flag.Bool("v", false, "Enable verbose logging")
	verboseLong := flag.Bool("verbose", false, "Enable verbose logging")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Southwest Check-in Bot\n")
		fmt.Fprintf(os.Stderr, "======================\n")
		fmt.Fprintf(os.Stderr, "Automatically checks in to Southwest flights 24 hours before departure.\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -c, --confirmation  Southwest confirmation number (required)\n")
		fmt.Fprintf(os.Stderr, "  -f, --first         Passenger first name (required)\n")
		fmt.Fprintf(os.Stderr, "  -l, --last          Passenger last name (required)\n")
		fmt.Fprintf(os.Stderr, "  -d, --departure     Departure time (required)\n")
		fmt.Fprintf(os.Stderr, "                      Format: 'YYYY-MM-DD HH:MM' or RFC3339\n")
		fmt.Fprintf(os.Stderr, "  -v, --verbose       Enable verbose logging\n")
		fmt.Fprintf(os.Stderr, "  -h, --help          Show this help message\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s -c ABC123 -f John -l Doe -d \"2024-01-15 14:30\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -c ABC123 -f John -l Doe -d \"2024-01-15T14:30:00-06:00\" -v\n", os.Args[0])
	}

	flag.Parse()

	// Merge short and long flag values
	conf := coalesce(*confirmation, *confirmationLong)
	first := coalesce(*firstName, *firstNameLong)
	last := coalesce(*lastName, *lastNameLong)
	dep := coalesce(*departure, *departureLong)
	verb := *verbose || *verboseLong

	// Validate required fields
	if conf == "" || first == "" || last == "" || dep == "" {
		fmt.Fprintf(os.Stderr, "Error: Missing required arguments\n\n")
		flag.Usage()
		os.Exit(1)
	}

	departureTime, err := parseTime(dep)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid departure time format: %v\n", err)
		fmt.Fprintf(os.Stderr, "Use format: 'YYYY-MM-DD HH:MM' or RFC3339\n")
		os.Exit(1)
	}

	reservation := models.Reservation{
		ConfirmationNumber: strings.ToUpper(conf),
		FirstName:          first,
		LastName:           last,
		DepartureTime:      departureTime,
	}

	fmt.Println("Southwest Check-in Bot")
	fmt.Println("======================")
	fmt.Printf("Confirmation: %s\n", reservation.ConfirmationNumber)
	fmt.Printf("Passenger: %s %s\n", reservation.FirstName, reservation.LastName)
	fmt.Printf("Departure: %s\n\n", reservation.DepartureTime.Format("2006-01-02 15:04:05 MST"))

	checkInTime := reservation.CheckInTime()
	sched := scheduler.New(checkInTime, verb)
	sched.WaitUntilCheckIn()

	fmt.Println()
	fmt.Println("Attempting check-in...")

	client := checkin.New(verb)
	boardingPasses, err := client.CheckIn(reservation)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError: Check-in failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("Check-in successful!")
	fmt.Println("====================")

	for _, pass := range boardingPasses {
		fmt.Printf("\nFlight %s: %s -> %s\n", pass.FlightNumber, pass.DepartureAirport, pass.ArrivalAirport)
		fmt.Printf("Departure: %s\n", pass.DepartureTime)
		fmt.Printf("Passenger: %s %s\n", pass.FirstName, pass.LastName)
		fmt.Printf("Boarding Position: %s%s\n", pass.BoardingGroup, pass.BoardingPosition)
	}

	fmt.Println()
}

// coalesce returns the first non-empty string.
func coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func parseTime(s string) (time.Time, error) {
	// Try RFC3339 first
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}

	// Try common formats
	formats := []string{
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
		"01/02/2006 15:04",
		"01/02/2006 3:04 PM",
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, s, time.Local); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}
