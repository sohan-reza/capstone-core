package model

import "time"

type File struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	OriginalName string    `json:"original_name"`
	StorageKey   string    `json:"storage_key"`
	DownloadURL  string    `json:"download_url"`
	Size         int64     `json:"size"`
	ContentType  string    `json:"content_type"`
	CreatedAt    time.Time `json:"created_at"`
}
