// backup 包只做本地 SQLite 文件复制和恢复，不解析或上传书签内容。
package backup

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Create(dbPath, targetPath string) (string, error) {
	if targetPath == "" {
		targetPath = filepath.Join(filepath.Dir(dbPath), "backups", "cola-"+time.Now().UTC().Format("20060102-150405")+".db")
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return "", err
	}
	if err := copyFile(dbPath, targetPath); err != nil {
		return "", err
	}
	return targetPath, nil
}

func CreateWithAssets(dbPath, previewsDir, targetPath string) (string, error) {
	if targetPath == "" {
		targetPath = filepath.Join(filepath.Dir(dbPath), "backups", "cola-"+time.Now().UTC().Format("20060102-150405")+".zip")
	}
	if strings.EqualFold(filepath.Ext(targetPath), ".db") {
		return Create(dbPath, targetPath)
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return "", err
	}
	out, err := os.Create(targetPath)
	if err != nil {
		return "", err
	}
	defer out.Close()
	writer := zip.NewWriter(out)
	defer writer.Close()
	if err := addFileToZip(writer, dbPath, "cola.db"); err != nil {
		return "", err
	}
	if stat, err := os.Stat(previewsDir); err == nil && stat.IsDir() {
		if err := filepath.WalkDir(previewsDir, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(filepath.Dir(previewsDir), path)
			if err != nil {
				return err
			}
			return addFileToZip(writer, path, filepath.ToSlash(rel))
		}); err != nil {
			return "", err
		}
	}
	return targetPath, nil
}

// Restore 在覆盖当前数据库前先创建快照，便于恢复失败或误操作后人工找回原始文件。
func Restore(dbPath, backupPath string) (string, error) {
	if backupPath == "" {
		return "", fmt.Errorf("backup path is required")
	}
	if _, err := os.Stat(backupPath); err != nil {
		return "", err
	}
	snapshot := dbPath + ".snapshot-" + time.Now().UTC().Format("20060102-150405")
	if _, err := Create(dbPath, snapshot); err != nil {
		return "", err
	}
	if err := copyFile(backupPath, dbPath); err != nil {
		return snapshot, err
	}
	return snapshot, nil
}

func RestoreWithAssets(dbPath, previewsDir, backupPath string) (string, error) {
	if strings.EqualFold(filepath.Ext(backupPath), ".zip") {
		snapshot := dbPath + ".snapshot-" + time.Now().UTC().Format("20060102-150405")
		if _, err := CreateWithAssets(dbPath, previewsDir, snapshot+".zip"); err != nil {
			return "", err
		}
		reader, err := zip.OpenReader(backupPath)
		if err != nil {
			return snapshot, err
		}
		defer reader.Close()
		for _, file := range reader.File {
			switch {
			case file.Name == "cola.db":
				if err := extractZipFile(file, dbPath); err != nil {
					return snapshot, err
				}
			case strings.HasPrefix(file.Name, "previews/"):
				target := filepath.Join(filepath.Dir(previewsDir), filepath.FromSlash(file.Name))
				if err := extractZipFile(file, target); err != nil {
					return snapshot, err
				}
			}
		}
		return snapshot, nil
	}
	return Restore(dbPath, backupPath)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func addFileToZip(writer *zip.Writer, src, name string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := writer.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	return err
}

func extractZipFile(file *zip.File, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	in, err := file.Open()
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(target)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
