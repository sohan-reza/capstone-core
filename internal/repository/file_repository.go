package repository

import (
	"github.com/sohan-reza/capstone-core/internal/model"

	"gorm.io/gorm"
)

type FileRepository interface {
	Create(file *model.File) error
	FindByID(id uint) (*model.File, error)
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

// In your repository
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
