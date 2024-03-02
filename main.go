package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/op/go-logging"
	"github.com/sg3des/argum"
)

var version = "v0.1.0"
var log = logging.MustGetLogger("CAMERA-IMAGES")

var args struct {
	Addr    string `argum:"pos,req" help:"path to the directory with images"`
	Timeout int    `argum:"--timeout" help:"timeout between images[ms]" default:"100"`
}

func init() {
	// logFormat := `%{color}[%{module} %{shortfile}] %{message}%{color:reset}`
	logFormat := `[%{module} %{shortfile}] %{message}`
	logging.SetFormatter(logging.MustStringFormatter(logFormat))
	logging.SetBackend(logging.NewLogBackend(os.Stderr, "", 0))

	argum.MustParse(&args)
}

func main() {
	fmt.Fprintln(os.Stderr, "[log] "+version)
	fixWD()

	timeout := time.Duration(args.Timeout) * time.Millisecond

	for {
		if args.Timeout == 0 {
			// need some timeout to not load the CPU at 100%
			time.Sleep(50 * time.Millisecond)
		} else {
			time.Sleep(timeout)
		}

		if err := filepath.WalkDir(args.Addr, func(fp string, fd fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if fd.IsDir() || !isSuitableImage(fp) {
				return nil
			}

			if args.Timeout > 0 {
				time.Sleep(timeout)
			}

			fmt.Fprintln(os.Stderr, "[log] "+fp)

			if err := writeImage(fp); err != nil {
				return err
			}

			return os.Remove(fp)
		}); err != nil {
			fmt.Fprintln(os.Stderr, "[error] "+err.Error())
		}
	}
}

func fixWD() {
	if dir := os.Getenv("RTMIPDIR"); dir != "" {
		if err := os.Chdir(dir); err != nil {
			log.Error(err)
		}
	}
}

func isSuitableImage(fp string) bool {
	ext := strings.ToLower(filepath.Ext(fp))
	return ext == ".jpg" || ext == ".jpeg"
}

func writeImage(fp string) error {
	f, err := os.Open(fp)
	if err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, f)
	return err
}
