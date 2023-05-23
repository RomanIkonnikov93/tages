//go:build windows

package fileos

import (
	"os"
	"syscall"
	"time"

	"github.com/RomanIkonnikov93/tages/internal/models"
	"github.com/RomanIkonnikov93/tages/pkg/pkg/logging"
)

func FileInfo(files []os.FileInfo, logger *logging.Logger) []models.Record {

	list := make([]models.Record, 0)

	for _, v := range files {
		d := v.Sys().(*syscall.Win32FileAttributeData)

		list = append(list, models.Record{
			FileName: v.Name(),
			Created:  time.Unix(0, d.CreationTime.Nanoseconds()).String(),
			Updated:  v.ModTime().String(),
		})
	}
	return list
}
