package model

import (
	"errors"
	"fmt"
	"github.com/NubeDev/flexy/helpers"
	"gorm.io/gorm"
	"log"
)

type Host struct {
	UUID      string    `gorm:"primary_key" json:"uuid"`
	CreatedAt JSONTime  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt JSONTime  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *JSONTime `sql:"index" json:"deleted_at"`
	Name      string    `json:"name"`
	IP        string    `json:"ip"`
}

func (Host) TableName() string {
	return TablePrefix + "host"
}

func GetHosts() []*Host {
	var hosts []*Host
	result := db.Model(&Host{}).Find(&hosts)
	if result.Error != nil {
		log.Printf("Error fetching hosts: %v", result.Error)
		return nil
	}
	return hosts
}

func GetHost(uuid string) (*Host, error) {
	var host Host
	result := db.Where("uuid = ?", uuid).First(&host)
	// Check for errors during the query
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("host with UUID %s not found", uuid)
		}
		log.Printf("Error fetching host with UUID %s: %v", uuid, result.Error)
		return nil, result.Error
	}

	return &host, nil
}

func UpdateHost(uuid string, body *Host) (*Host, error) {
	// Find the host by UUID first
	var host Host
	result := db.Where("uuid = ?", uuid).First(&host)
	// Check if the host exists
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("host with UUID %s not found", uuid)
		}
		log.Printf("Error fetching host with UUID %s: %v", uuid, result.Error)
		return nil, result.Error
	}
	// Update the host with the provided data
	result = db.Model(&host).Updates(body)

	// Check if there's an error during the update
	if result.Error != nil {
		log.Printf("Error updating host with UUID %s: %v", uuid, result.Error)
		return nil, result.Error
	}
	// Return the updated host and nil error
	return &host, nil
}

func CreateHost(body *Host) error {
	if body.UUID == "" {
		body.UUID = helpers.UUID()
	}
	result := db.Create(body)
	if result.Error != nil {
		log.Printf("Error creating host: %v", result.Error)
		return result.Error
	}

	return nil
}
