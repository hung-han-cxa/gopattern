package models

import (
	"github.com/jinzhu/gorm"
)

type Verification struct {
	gorm.Model
	Name   string `gorm:"size:100;not null"`
	Token  string `gorm:"size:255;not null"`
	User   User   `gorm:"ForeignKey:UserID" json:"-"`
	UserID uint   `gorm:"not null"`
}

// GetVerfication get verification data
func (verification Verification) GetVerificationByID(userID string, name string, db *gorm.DB) (*Verification, error) {
	if err := db.Debug().Table("verifications").Where("user_id = ?", userID).Where("name = ?", name).First(&verification).Error; err != nil {
		return nil, err
	}
	return &verification, nil
}

// GetVerfication get verification data
func (verification Verification) GetVerificationByToken(token string, db *gorm.DB) (*Verification, error) {
	if err := db.Debug().Table("verifications").Where("token = ?", token).First(&verification).Error; err != nil {
		return nil, err
	}
	return &verification, nil
}

// DeleteVerification delete verification data
func (verification *Verification) DeleteVerification(id string, db *gorm.DB) (*Verification, error) {
	if err := db.Debug().Table("verifications").Where("id = ?", id).Unscoped().Delete(&verification).Error; err != nil {
		return nil, err
	}
	return nil, nil
}