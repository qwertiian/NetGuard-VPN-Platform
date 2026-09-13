package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Manager struct {
	dataDir string
	dbPath  string
}

func NewManager(dataDir, dbPath string) *Manager {
	return &Manager{
		dataDir: dataDir,
		dbPath:  dbPath,
	}
}

func (m *Manager) CreateBackup(outputPath string) error {
	out, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer out.Close()

	gw := gzip.NewWriter(out)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Add DB
	if err := addFileToTar(tw, m.dbPath, filepath.Base(m.dbPath)); err != nil {
		// Ignore if DB doesn't exist yet
	}

	// Add key and conf files from dataDir
	err = filepath.Walk(m.dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		
		ext := filepath.Ext(path)
		if ext == ".key" || ext == ".psk" || ext == ".yaml" || ext == ".yml" {
			rel, _ := filepath.Rel(m.dataDir, path)
			if err := addFileToTar(tw, path, rel); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	fmt.Println("Warning: Backup archive contains sensitive key material.")
	return nil
}

func (m *Manager) RestoreBackup(archivePath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Security: Check for zip slip
		if strings.Contains(header.Name, "..") {
			continue
		}

		target := filepath.Join(m.dataDir, header.Name)
		
		if header.FileInfo().IsDir() {
			os.MkdirAll(target, 0700)
			continue
		}

		os.MkdirAll(filepath.Dir(target), 0700)
		
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return err
		}
		out.Close()
	}

	return nil
}

type BackupInfo struct {
	Path      string
	Size      int64
	CreatedAt time.Time
}

func (m *Manager) ListBackups(backupDir string) ([]BackupInfo, error) {
	var backups []BackupInfo
	
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return backups, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tar.gz") {
			continue
		}
		
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		backups = append(backups, BackupInfo{
			Path:      filepath.Join(backupDir, entry.Name()),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
		})
	}
	
	return backups, nil
}

func addFileToTar(tw *tar.Writer, path string, name string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}
	header.Name = name
	
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	
	_, err = io.Copy(tw, f)
	return err
}
