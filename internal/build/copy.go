package build

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// copyTree copies all files from src to dst, preserving directory structure.
func copyTree(src, dst string, overwrite bool) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlinks not handled: %s", path)
		}

		rel, _ := filepath.Rel(src, path)
		outPath := filepath.Join(dst, rel)

		if !overwrite {
			// Do not overwrite existing files
			if _, err := os.Stat(outPath); err == nil {
				return nil
			}
		}

		return copyFile(outPath, path)
	})
}

// copyContent copies files in content/ that are NOT processed as templates.
func copyContent(contentDir, outDir string) error {
	return filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlinks not handled: %s", path)
		}

		// skip files we process
		if shouldProcessFile(path) {
			return nil
		}

		rel, _ := filepath.Rel(contentDir, path)
		outPath := filepath.Join(outDir, rel)

		return copyFile(outPath, path)
	})
}

func copyFile(dstPath, srcPath string) error {
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return err
	}
	in, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = in.Close()
	}()

	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := out.Close()
		if closeErr != nil {
			closeErr = fmt.Errorf("error closing file: %w", closeErr)
			if err != nil {
				err = errors.Join(err, closeErr)
			} else {
				err = closeErr
			}
		}
	}()

	_, err = io.Copy(out, in)
	return err
}

func CopyEmbeddedResources(dst string, src fs.FS) error {
	return fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("symlinks not handled: %s", path)
		}

		outPath := filepath.Join(dst, path)
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}

		in, err := src.Open(path)
		if err != nil {
			return err
		}
		defer func() {
			_ = in.Close()
		}()

		out, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer func() {
			closeErr := out.Close()
			if closeErr != nil {
				closeErr = fmt.Errorf("error closing file: %w", closeErr)
				if err != nil {
					err = errors.Join(err, closeErr)
				} else {
					err = closeErr
				}
			}
		}()

		_, err = io.Copy(out, in)
		return err
	})
}
