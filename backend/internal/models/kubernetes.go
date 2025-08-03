package models

import (
	"time"

	"gorm.io/gorm"
)

type KubernetesCluster struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	UserID     uint           `json:"user_id" gorm:"not null"`
	Name       string         `json:"name" gorm:"not null"`
	KubeConfig string         `json:"kube_config" gorm:"type:text;not null"`
	ClusterURL string         `json:"cluster_url"`
	Version    string         `json:"version"`
	Status     string         `json:"status" gorm:"default:'pending'"`
	IsActive   bool           `json:"is_active" gorm:"default:true"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type ClusterValidationResponse struct {
	IsValid    bool   `json:"is_valid"`
	Version    string `json:"version,omitempty"`
	Error      string `json:"error,omitempty"`
	ClusterURL string `json:"cluster_url,omitempty"`
}

type ClusterStatus struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	IsActive bool   `json:"is_active"`
	Version  string `json:"version"`
}
