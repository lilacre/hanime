package downloader

import (
	"os/exec"
)

func concat(filelistsPATH string, outputPATH string) ([]byte, error) {
	cmd := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", filelistsPATH, "-c", "copy", outputPATH)
	stdout, err := cmd.Output()
	if err != nil {
		return stdout, err
	}
	return nil, nil
}
