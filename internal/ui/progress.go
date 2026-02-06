// Package ui provides user interface utilities for the CLI.
package ui

import (
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/progress"
)

const (
	sleepDuration = 100 * time.Millisecond
	trackerLength = 25
)

// WithSpinner runs a function with an indeterminate spinner.
// The spinner writes to stderr, keeping stdout clean.
func WithSpinner(message string, fn func() error) error {
	pw := progress.NewWriter()
	pw.SetOutputWriter(os.Stderr)
	pw.SetAutoStop(true)
	pw.SetTrackerLength(trackerLength)
	pw.SetStyle(progress.StyleDefault)
	pw.Style().Visibility.ETA = false
	pw.Style().Visibility.Percentage = false
	pw.Style().Visibility.Value = false

	go pw.Render()

	tracker := &progress.Tracker{
		Message: message,
		Total:   0, // Indeterminate
	}
	pw.AppendTracker(tracker)

	err := fn()

	if err != nil {
		tracker.MarkAsErrored()
	} else {
		tracker.MarkAsDone()
	}

	time.Sleep(sleepDuration)
	pw.Stop()

	return err
}

// WithProgressBar runs a function with a determinate progress bar.
func WithProgressBar(message string, total int64, fn func(increment func(n int64)) error) error {
	pw := progress.NewWriter()
	pw.SetOutputWriter(os.Stderr)
	pw.SetAutoStop(true)
	pw.SetTrackerLength(trackerLength)

	go pw.Render()

	tracker := &progress.Tracker{
		Message: message,
		Total:   total,
	}
	pw.AppendTracker(tracker)

	increment := func(n int64) {
		tracker.Increment(n)
	}

	err := fn(increment)

	if err != nil {
		tracker.MarkAsErrored()
	} else {
		tracker.MarkAsDone()
	}

	time.Sleep(sleepDuration)
	pw.Stop()

	return err
}
