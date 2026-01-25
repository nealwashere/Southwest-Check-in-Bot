package scheduler

import (
	"fmt"
	"time"
)

type Scheduler struct {
	CheckInTime time.Time
	Verbose     bool
}

func New(checkInTime time.Time, verbose bool) *Scheduler {
	return &Scheduler{
		CheckInTime: checkInTime,
		Verbose:     verbose,
	}
}

// WaitUntilCheckIn waits until the check-in time, displaying a countdown.
// Returns immediately if the check-in time has already passed.
func (s *Scheduler) WaitUntilCheckIn() {
	remaining := time.Until(s.CheckInTime)

	if remaining <= 0 {
		fmt.Println("Check-in window is already open!")
		return
	}

	fmt.Printf("Check-in opens: %s\n", s.CheckInTime.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("Time until check-in: %s\n\n", formatDuration(remaining))
	fmt.Println("Waiting for check-in window...")

	// Update countdown every second until we're close, then more frequently
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		remaining = time.Until(s.CheckInTime)

		if remaining <= 0 {
			fmt.Print("\r" + clearLine())
			fmt.Println("Check-in window is NOW OPEN!")
			return
		}

		s.displayProgress(remaining)

		// When we're very close, switch to busy-wait for precision
		if remaining <= 100*time.Millisecond {
			s.busyWait()
			fmt.Print("\r" + clearLine())
			fmt.Println("Check-in window is NOW OPEN!")
			return
		}

		<-ticker.C
	}
}

func (s *Scheduler) busyWait() {
	for time.Now().Before(s.CheckInTime) {
		// Tight loop for precision timing
	}
}

func (s *Scheduler) displayProgress(remaining time.Duration) {
	// Calculate progress (assuming max wait of 24 hours)
	maxWait := 24 * time.Hour
	elapsed := maxWait - remaining
	if elapsed < 0 {
		elapsed = 0
	}
	progress := float64(elapsed) / float64(maxWait)
	if progress > 1 {
		progress = 1
	}

	barWidth := 30
	filled := int(progress * float64(barWidth))

	bar := ""
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "="
		} else if i == filled {
			bar += ">"
		} else {
			bar += " "
		}
	}

	fmt.Printf("\r[%s] %3.0f%% %s remaining", bar, progress*100, formatDuration(remaining))
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		return "0s"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func clearLine() string {
	return "\033[K"
}
