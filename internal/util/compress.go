package util

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

func CompressDirectory(srcDirectory string, destPath string) (string, error) {
	file, err := os.OpenFile(destPath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	err = filepath.Walk(srcDirectory, createRecursiveWalker(w, srcDirectory))
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}

func createRecursiveWalker(w *zip.Writer, rootPath string) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		f, err := w.Create(relPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		return nil
	}
}
