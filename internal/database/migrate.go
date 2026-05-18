package database

import (
	"log"

	"gorm.io/gorm"
	"kirimaja-go/models"
)

// AutoMigrate creates/updates all tables from the GORM models.
// Order matters: parents before children so FK constraints resolve.
func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.Role{},
		&models.Permission{},
		&models.RolePermission{},
		&models.User{},
		&models.Branch{},
		&models.EmployeeBranch{},
		&models.UserAddress{},
		&models.Shipment{},
		&models.ShipmentDetail{},
		&models.ShipmentHistory{},
		&models.Payment{},
		&models.ShipmentBranchLog{},
	); err != nil {
		return err
	}
	log.Println("Database migrated")
	return nil
}
