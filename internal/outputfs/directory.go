// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package outputfs

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// ReplaceDirectory generates a complete output set in a sibling directory.
// Unmanaged entries are preserved. Existing results are backed up during the
// short publication step and restored if publication fails. After an abrupt
// interruption, the next invocation recovers the backup under the same lock.
func ReplaceDirectory(
	dir string,
	clean bool,
	managed func(relativePath string) bool,
	generate func(stage string) ([]string, error),
) ([]string, error) {
	absolute, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	parent, base := filepath.Dir(absolute), filepath.Base(absolute)
	if parent == absolute {
		return nil, errors.New("filesystem root cannot be an output directory")
	}
	if mkdirErr := os.MkdirAll(parent, 0o700); mkdirErr != nil {
		return nil, mkdirErr
	}
	unlock, err := lockDirectory(filepath.Join(parent, "."+base+".codeknit.lock"))
	if err != nil {
		return nil, err
	}
	defer unlock()

	backup := filepath.Join(parent, "."+base+".codeknit-backup")
	if recoveryErr := recoverDirectory(absolute, backup); recoveryErr != nil {
		return nil, recoveryErr
	}
	existing, err := os.Lstat(absolute)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if existing != nil && !existing.IsDir() {
		return nil, fmt.Errorf("output %s must be a directory, not a symlink or file", dir)
	}
	stage, err := os.MkdirTemp(parent, "."+base+".codeknit-stage-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(stage) }()

	written, err := generate(stage)
	if err != nil {
		return nil, err
	}
	outputs := make([]string, len(written))
	for i, path := range written {
		rel, err := filepath.Rel(stage, path)
		if err != nil || !filepath.IsLocal(rel) {
			return nil, fmt.Errorf("generated path %q is outside staging directory", path)
		}
		outputs[i] = filepath.Join(dir, rel)
	}
	if existing != nil {
		if err := preserveUnmanaged(absolute, stage, clean, managed); err != nil {
			return nil, err
		}
		if err := os.Chmod(stage, existing.Mode().Perm()); err != nil {
			return nil, err
		}
	}
	if err := publishDirectory(absolute, stage, backup, existing != nil, os.Rename); err != nil {
		return nil, err
	}
	return outputs, nil
}

func recoverDirectory(dir, backup string) error {
	info, err := os.Lstat(backup) //nolint:gosec // validated sibling backup chosen by ReplaceDirectory
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("output backup %s is not a directory", backup)
	}
	if current, err := os.Lstat(dir); os.IsNotExist(err) { //nolint:gosec // explicit, validated output destination
		return os.Rename(backup, dir) //nolint:gosec // restore the validated sibling backup
	} else if err != nil {
		return err
	} else if !current.IsDir() {
		return fmt.Errorf("cannot recover output: %s is not a directory; previous output remains at %s", dir, backup)
	}
	return os.RemoveAll(backup) //nolint:gosec // validated sibling backup after successful publication
}

func publishDirectory(dir, stage, backup string, exists bool, rename func(string, string) error) error {
	if exists {
		if err := rename(dir, backup); err != nil {
			return err
		}
	}
	if err := rename(stage, dir); err != nil {
		if exists {
			if rollbackErr := rename(backup, dir); rollbackErr != nil {
				return errors.Join(err, fmt.Errorf("previous output remains at %s: %w", backup, rollbackErr))
			}
		}
		return err
	}
	if exists {
		if err := os.RemoveAll(backup); err != nil { //nolint:gosec // validated sibling backup after successful publication
			return fmt.Errorf("output published; could not remove backup %s: %w", backup, err)
		}
	}
	return nil
}

func preserveUnmanaged(source, stage string, clean bool, managed func(string) bool) error {
	original, err := os.OpenRoot(source)
	if err != nil {
		return err
	}
	defer func() { _ = original.Close() }()
	staged, err := os.OpenRoot(stage)
	if err != nil {
		return err
	}
	defer func() { _ = staged.Close() }()
	var directories []string
	err = fs.WalkDir(original.FS(), ".", func(rel string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || rel == "." {
			return walkErr
		}
		if entry.IsDir() {
			directories = append(directories, rel)
			return staged.MkdirAll(rel, 0o700)
		}
		if managed(rel) {
			if !clean {
				return fmt.Errorf("output contains %s from a previous run; pass --clean or choose another directory", rel)
			}
			return nil
		}
		if _, statErr := staged.Lstat(rel); statErr == nil {
			return fmt.Errorf("generated output would overwrite unrelated file %s", filepath.Join(source, rel))
		} else if !os.IsNotExist(statErr) {
			return statErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			link, linkErr := original.Readlink(rel)
			if linkErr != nil {
				return linkErr
			}
			return staged.Symlink(link, rel)
		}
		return copyFile(original, staged, rel)
	})
	if err != nil {
		return err
	}
	// Restore directory permissions after copying children, from deepest up.
	for i := len(directories) - 1; i >= 0; i-- {
		info, err := original.Lstat(directories[i])
		if err != nil {
			return err
		}
		if err := staged.Chmod(directories[i], info.Mode().Perm()); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(source, target *os.Root, rel string) error {
	info, err := source.Lstat(rel)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("cannot preserve non-regular output entry %s", rel)
	}
	in, err := source.Open(rel)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	info, err = in.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("cannot preserve non-regular output entry %s", rel)
	}
	out, err := target.OpenFile(rel, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	return errors.Join(copyErr, out.Sync(), out.Close())
}
