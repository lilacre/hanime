package downloader

import (
	"os"
)

func clean(tmpPATH string) error {
	return os.RemoveAll(tmpPATH)
}
