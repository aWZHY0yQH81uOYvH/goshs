//go:build windows

package httpserver

import (
	"github.com/aWZHY0yQH81uOYvH/goshs/logger"
)

func (fs *FileServer) dropPrivs() {
	if fs.DropUser != "" {
		logger.Warn("Dropping privileges with --user only works for unix systems, sorry.")
	}
}
