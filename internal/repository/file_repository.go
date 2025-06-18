package repository

import (
	"errors"
	"fmt"

	"github.com/sohan-reza/capstone-core/internal/model"

	"gorm.io/gorm"
)

type FileRepository interface {
	Create(file *model.File) error
	FindByID(id uint) (*model.File, error)
	DeleteByKey(key string) error
	GetFilesByTeamID(teamID string) ([]struct {
		DownloadURL string `json:"download_url"`
		FileType    string `json:"file_type"`
	}, error)
}

type fileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) Create(file *model.File) error {
	return r.db.Create(file).Error
}

func (r *fileRepository) FindByID(id uint) (*model.File, error) {
	var file model.File
	err := r.db.First(&file, id).Error
	return &file, err
}

func (r *fileRepository) DeleteByKey(key string) error {
	if key == "" {
		return errors.New("empty file key provided")
	}

	result := r.db.Where("storage_key = ?", key).Delete(&model.File{})
	if result.Error != nil {
		return fmt.Errorf("database deletion failed: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no record found with key: %s", key)
	}

	return nil
}

func (r *fileRepository) GetFilesByTeamID(teamID string) ([]struct {
	DownloadURL string `json:"download_url"`
	FileType    string `json:"file_type"`
}, error) {
	var results []struct {
		DownloadURL string `json:"download_url"`
		FileType    string `json:"file_type"`
	}

	err := r.db.Model(&model.File{}).
		Where("team_id = ?", teamID).
		Select("download_url, file_type").
		Find(&results).Error

	return results, err
}

// func (r *fileRepository) GetURLWithRefresh(id uint) (string, error) {
// 	file, err := r.FindByID(id)
// 	if err != nil {
// 		return "", err
// 	}

// 	// Check if URL needs refresh (e.g., expires in less than 1 day)
// 	if time.Until(file.URLExpiresAt) < 24*time.Hour {
// 		newURL, err := r.awsService.GenerateLongTermPresignedURL(file.StorageKey)
// 		if err != nil {
// 			return "", err
// 		}
// 		// Update the record with new URL
// 		file.DownloadURL = newURL
// 		file.URLExpiresAt = time.Now().Add(168 * time.Hour)
// 		r.db.Save(file)
// 	}

// 	return file.DownloadURL, nil
// }
