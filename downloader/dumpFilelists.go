package downloader

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func dumpFilelists(num int, filelistsPATH string, tmpPATH string) error {
	fo, err := os.Create(filelistsPATH)
	defer fo.Close()
	if err != nil {
		return err
	}

	w := bufio.NewWriter(fo)
	for i := 0; i < num; i++ {
		p := "'" + filepath.Join(tmpPATH, strconv.Itoa(i)+".ts") + "'"
		fmt.Fprintln(w, "file "+p)
	}
	w.Flush()
	return nil
}
