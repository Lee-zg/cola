// backup 包只做本地 SQLite 文件复制和恢复，不解析或上传书签内容。
package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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
