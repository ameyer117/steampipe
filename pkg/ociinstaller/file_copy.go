package ociinstaller

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func CopyFile(sourcePath string, destPath string) error {
	info, err := os.Lstat(sourcePath)
	if err != nil {
		return err
	}

	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(sourcePath)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}
		if err := os.RemoveAll(destPath); err != nil {
			return err
		}
		return os.Symlink(target, destPath)
	}

	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
	if err != nil {
		return err
	}
	defer dest.Close()

	if _, err := io.Copy(dest, source); err != nil {
		return err
	}
	return nil
}

func CopyFolder(sourcePath string, destPath string) error {
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(destPath, sourceInfo.Mode()); err != nil {
		return err
	}

	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		targetPath := filepath.Join(destPath, relPath)
		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}
		if err := CopyFile(path, targetPath); err != nil {
			return fmt.Errorf("copy %s to %s: %w", path, targetPath, err)
		}
		return nil
	})
}
