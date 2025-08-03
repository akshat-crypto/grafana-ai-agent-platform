package models

import (
	"time"

	"gorm.io/gorm"
)

type AgentQuery struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	ClusterID uint           `json:"cluster_id"`
	Query     string         `json:"query" gorm:"type:text;not null"`
	Response  string         `json:"response" gorm:"type:text"`
	Status    string         `json:"status" gorm:"default:'pending'"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User    User              `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Cluster KubernetesCluster `json:"cluster,omitempty" gorm:"foreignKey:ClusterID"`
}

type Deployment struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	ClusterID uint           `json:"cluster_id" gorm:"not null"`
	StackName string         `json:"stack_name" gorm:"not null"`
	Status    string         `json:"status" gorm:"default:'pending'"`
	Manifest  string         `json:"manifest" gorm:"type:text"`
	Error     string         `json:"error" gorm:"type:text"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User    User              `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Cluster KubernetesCluster `json:"cluster,omitempty" gorm:"foreignKey:ClusterID"`
}

type AgentRequest struct {
	Query     string `json:"query" binding:"required"`
	ClusterID uint   `json:"cluster_id,omitempty"`
}

type AgentResponse struct {
	Response string `json:"response"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}
