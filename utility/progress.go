package utility

import (
	"strings"

	"github.com/sirupsen/logrus"
)

// ProgressBar use for displaying working status
type ProgressBar struct {
	MaxLength    int
	currentIndex int
	buffer       []string
	Logger       *logrus.Logger
}

// Next update status for ProgressBar for on step
func (progressBar *ProgressBar) Next() {
	if progressBar.currentIndex == 0 {
		progressBar.buffer = make([]string, progressBar.MaxLength+2)
		progressBar.buffer[0] = "["
		progressBar.buffer[progressBar.MaxLength+1] = "]"
		for i := 1; i < progressBar.MaxLength+1; i++ {
			progressBar.buffer[i] = "-"
		}
	}
	progressBar.currentIndex++
	progressBar.buffer[progressBar.currentIndex] = "#"

	if progressBar.currentIndex == progressBar.MaxLength {
		progressBar.Logger.Infoln(strings.Join(progressBar.buffer, "") + "\n" + "\r")
	} else {
		progressBar.Logger.Infoln(strings.Join(progressBar.buffer, "") + "\r")
	}
}
