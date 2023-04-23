package api

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DownloadProgress struct for progress tracking
type DownloadProgress struct {
	sync.Mutex
	isDownloadCompleted bool
	lastCheck           time.Time
	context.CancelFunc
	result []byte
}

type DownloadManager struct {
	progress map[string]*DownloadProgress
}

func NewDownloadManager() *DownloadManager {
	return &DownloadManager{
		progress: make(map[string]*DownloadProgress),
	}
}

func (dm *DownloadManager) SetProgress(token string, isDownloadCompleted bool) error {
	if _, ok := dm.progress[token]; !ok {
		return fmt.Errorf("token not found")
	}
	v := dm.progress[token]
	v.Mutex.Lock()
	defer v.Mutex.Unlock()
	v.isDownloadCompleted = isDownloadCompleted
	return nil
}

func (dm *DownloadManager) GetProgress(token string) (bool, error) {
	if _, ok := dm.progress[token]; !ok {
		return false, fmt.Errorf("token not found")
	}
	v := dm.progress[token]
	v.Mutex.Lock()
	defer v.Mutex.Unlock()

	v.lastCheck = time.Now()

	return v.isDownloadCompleted, nil
}

func (dm *DownloadManager) SetResult(token string, result []byte) error {
	if _, ok := dm.progress[token]; !ok {
		return fmt.Errorf("token not found")
	}
	v := dm.progress[token]
	v.Mutex.Lock()
	defer v.Mutex.Unlock()
	v.result = result
	return nil
}

func (dm *DownloadManager) GetResult(token string) ([]byte, error) {
	if _, ok := dm.progress[token]; !ok {
		return []byte{}, fmt.Errorf("token not found")
	}
	v := dm.progress[token]
	v.Mutex.Lock()
	defer v.Mutex.Unlock()

	return v.result, nil
}

func (dm *DownloadManager) GetLastCheck(token string) (time.Time, error) {
	if _, ok := dm.progress[token]; !ok {
		return time.Now(), fmt.Errorf("token not found")
	}
	v := dm.progress[token]
	v.Mutex.Lock()
	defer v.Mutex.Unlock()

	return v.lastCheck, nil
}

func (dm *DownloadManager) CancelDownload(token string) error {
	if _, ok := dm.progress[token]; !ok {
		return fmt.Errorf("token not found")
	}
	v := dm.progress[token]
	v.Mutex.Lock()
	defer v.Mutex.Unlock()

	v.CancelFunc()

	return nil
}
