package lib

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/denfm/singlebackuper/internal/app/service"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type CompressOptions struct {
	Paths        []string
	Buf          *bytes.Buffer
	ExcludesPath []string
	IsMultiType  bool
}

func Compress(opt *CompressOptions) error {
	// tar > gzip > buf
	zr := gzip.NewWriter(opt.Buf)
	tw := tar.NewWriter(zr)

	for _, path := range opt.Paths {
		if !service.HasDir(path) {
			return fmt.Errorf("the specified backup directory \"%s\" does not exist", path)
		}

		if path == "/" {
			return fmt.Errorf("bad path")
		}

		err := compressWalk(path, tw, opt)
		if err != nil {
			return err
		}
	}

	if err := tw.Close(); err != nil {
		return err
	}

	if err := zr.Close(); err != nil {
		return err
	}

	return nil
}

func compressWalk(path string, tw *tar.Writer, opt *CompressOptions) error {
	err := filepath.Walk(path, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == file {
			return nil
		}

		var name string

		if opt.IsMultiType {
			pathExplode := strings.Split(strings.TrimLeft(file, "/"), "/")
			name = fmt.Sprintf("%s/%s", pathExplode[len(pathExplode)-2], fi.Name())
		} else {
			name = strings.TrimLeft(strings.TrimPrefix(file, path), "/")
		}

		isDir := fi.IsDir()

		if len(opt.ExcludesPath) > 0 {
			for _, exName := range opt.ExcludesPath {
				if isDir && exName == name {
					return filepath.SkipDir
				}

				if exName == name {
					return nil
				}
			}
		}

		if fi.Mode()&os.ModeSymlink != 0 {
			if file, err = os.Readlink(file); err != nil {
				return err
			}
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		header.Name = name

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})

	return err
}
