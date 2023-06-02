//go:build linux

package fileos

import (
	"os"
	"path/filepath"
	"time"

	"github.com/RomanIkonnikov93/tages/internal/models"
	"golang.org/x/sys/unix"
)

func FileInfo(files []os.FileInfo) ([]models.Record, error) {

	list := make([]models.Record, 0)

	for _, v := range files {

		path := filepath.Clean("storage/" + v.Name())

		var statx unix.Statx_t

		err := unix.Statx(unix.AT_FDCWD, path, 0, 0, &statx)
		if err != nil {
			return nil, err
		}

		created := ""

		t := time.Unix(int64(statx.Btime.Sec), int64(statx.Btime.Nsec))

		if t.String() == "1970-01-01 06:00:00 +0600 +06" {
			created = time.Unix(int64(statx.Ctime.Sec), int64(statx.Ctime.Nsec)).String()
		} else {
			created = t.String()
		}

		updated := time.Unix(int64(statx.Atime.Sec), int64(statx.Atime.Nsec))

		list = append(list, models.Record{
			FileName: v.Name(),
			Created:  created,
			Updated:  updated.String(),
		})

	}

	return list, nil
}
