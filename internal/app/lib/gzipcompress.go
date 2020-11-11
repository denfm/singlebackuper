package lib

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Compress(path string, buf io.Writer, excludesPath []string) error {
	// tar > gzip > buf
	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	err := filepath.Walk(path, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == file {
			return nil
		}

		name := strings.TrimPrefix(file, path)
		isDir := fi.IsDir()

		if len(excludesPath) > 0 {
			for _, exName := range excludesPath {
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

	if err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}

	if err := zr.Close(); err != nil {
		return err
	}

	return nil
}
