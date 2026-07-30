package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Storage handles file storage with sharded directories
type Storage struct {
	dataDir string
	tmpDir  string
}

func NewStorage(dataDir, tmpDir string) (*Storage, error) {
	if err := os.MkdirAll(dataDir, 0750); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	if err := os.MkdirAll(tmpDir, 0750); err != nil {
		return nil, fmt.Errorf("create tmp dir: %w", err)
	}
	return &Storage{dataDir: dataDir, tmpDir: tmpDir}, nil
}

// generateShardPath creates a sharded path like /data/ab/cd/abcdef...
func (s *Storage) generateShardPath(fileID string) string {
	if len(fileID) < 4 {
		return filepath.Join(s.dataDir, fileID)
	}
	return filepath.Join(s.dataDir, fileID[:2], fileID[2:4], fileID)
}

// GenerateFileID generates a secure random file ID
func GenerateFileID() (string, error) {
	bytes := make([]byte, 16) // 128 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateTempFile creates a temporary file for upload
func (s *Storage) CreateTempFile(uploadID int64) (*os.File, string, error) {
	filename := fmt.Sprintf("%d.part", uploadID)
	path := filepath.Join(s.tmpDir, filename)
	
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		return nil, "", err
	}
	
	return f, path, nil
}

// GetTempFilePath returns the path to a temp file
func (s *Storage) GetTempFilePath(uploadID int64) string {
	return filepath.Join(s.tmpDir, fmt.Sprintf("%d.part", uploadID))
}

// FinalizeFile moves a completed upload to its final location
func (s *Storage) FinalizeFile(tempPath, fileID string) (string, error) {
	finalPath := s.generateShardPath(fileID)
	
	// Ensure shard directories exist
	if err := os.MkdirAll(filepath.Dir(finalPath), 0750); err != nil {
		return "", fmt.Errorf("create shard dir: %w", err)
	}
	
	// Atomic rename (must be on same filesystem)
	if err := os.Rename(tempPath, finalPath); err != nil {
		// If rename fails (cross-device), copy and delete
		if err := copyFile(finalPath, tempPath); err != nil {
			os.Remove(finalPath)
			return "", fmt.Errorf("copy file: %w", err)
		}
		os.Remove(tempPath)
	}
	
	// Set proper permissions
	if err := os.Chmod(finalPath, 0640); err != nil {
		return "", fmt.Errorf("chmod: %w", err)
	}
	
	return finalPath, nil
}

// GetFilePath returns the full path to a stored file
func (s *Storage) GetFilePath(fileID string) string {
	return s.generateShardPath(fileID)
}

// OpenFile opens a stored file for reading
func (s *Storage) OpenFile(fileID string) (*os.File, error) {
	path := s.generateShardPath(fileID)
	return os.Open(path)
}

// FileExists checks if a file exists
func (s *Storage) FileExists(fileID string) bool {
	path := s.generateShardPath(fileID)
	_, err := os.Stat(path)
	return err == nil
}

// DeleteFile removes a stored file
func (s *Storage) DeleteFile(fileID string) error {
	path := s.generateShardPath(fileID)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	
	// Try to clean up empty parent directories
	parent := filepath.Dir(path)
	os.Remove(parent)
	os.Remove(filepath.Dir(parent))
	
	return nil
}

// DeleteTempFile removes a temporary upload file
func (s *Storage) DeleteTempFile(uploadID int64) error {
	path := s.GetTempFilePath(uploadID)
	return os.Remove(path)
}

// GetFileSize returns the size of a stored file
func (s *Storage) GetFileSize(fileID string) (int64, error) {
	path := s.generateShardPath(fileID)
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetTempFileSize returns the size of a temp file
func (s *Storage) GetTempFileSize(uploadID int64) (int64, error) {
	path := s.GetTempFilePath(uploadID)
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// ListTempFiles lists all temporary files
func (s *Storage) ListTempFiles() ([]string, error) {
	var files []string
	
	err := filepath.Walk(s.tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".part") {
			files = append(files, path)
		}
		return nil
	})
	
	return files, err
}

// GetDiskStats returns disk usage statistics
func (s *Storage) GetDiskStats() (total, free uint64, err error) {
	// Use syscall for disk stats - simplified version
	// In production, use golang.org/x/sys/unix
	statfs := &syscallStatfs{}
	if err := syscallStatfsGet(s.dataDir, statfs); err != nil {
		return 0, 0, err
	}
	
	total = statfs.Blocks * statfs.Bsize
	free = statfs.Bfree * statfs.Bsize
	
	return total, free, nil
}

// copyFile copies a file from src to dst
func copyFile(dst, src string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	return err
}

// syscallStatfs is a placeholder for unix.Statfs_t
type syscallStatfs struct {
	Blocks uint64
	Bsize  uint64
	Bfree  uint64
}

// syscallStatfsGet is a placeholder for unix.Statfs
func syscallStatfsGet(path string, stat *syscallStatfs) error {
	// In production, use:
	// import "golang.org/x/sys/unix"
	// var fs unix.Statfs_t
	// if err := unix.Statfs(path, &fs); err != nil { return err }
	// stat.Blocks = fs.Blocks
	// stat.Bsize = uint64(fs.Bsize)
	// stat.Bfree = fs.Bfree
	
	// Fallback: use df command or return dummy values
	return nil
}

// WriteAt writes data at offset in a file
func (s *Storage) WriteAt(f *os.File, data []byte, offset int64) (int, error) {
	return f.WriteAt(data, offset)
}

// ReadRange reads a range from a file
func (s *Storage) ReadRange(fileID string, start, end int64) (io.ReadCloser, error) {
	f, err := s.OpenFile(fileID)
	if err != nil {
		return nil, err
	}
	
	if start > 0 {
		if _, err := f.Seek(start, io.SeekStart); err != nil {
			f.Close()
			return nil, err
		}
	}
	
	return &rangeReader{f: f, remaining: end - start + 1}, nil
}

type rangeReader struct {
	f         *os.File
	remaining int64
}

func (r *rangeReader) Read(p []byte) (int, error) {
	if r.remaining <= 0 {
		return 0, io.EOF
	}
	
	if int64(len(p)) > r.remaining {
		p = p[:r.remaining]
	}
	
	n, err := r.f.Read(p)
	r.remaining -= int64(n)
	
	if err == io.EOF && r.remaining > 0 {
		err = nil
	}
	
	return n, err
}

func (r *rangeReader) Close() error {
	return r.f.Close()
}
