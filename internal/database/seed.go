package database

import (
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"kirimaja-go/models"
)

type seedRole struct{ Name, Key string }
type seedPermission struct{ Name, Key, Resource string }
type seedUser struct{ Name, Email, Password, Phone, RoleKey string }
type seedBranch struct{ Name, Address, Phone string }
type seedEmployee struct {
	Name, Email, Password, Phone, RoleKey, BranchName, Type string
}

var rolesData = []seedRole{
	{"Super Admin", "super-admin"},
	{"Customer", "customer"},
	{"Courier", "courier"},
	{"Admin Branch", "admin-branch"},
}

var permissionsData = []seedPermission{
	{"Create Shipment", "shipments.create", "shipments"},
	{"Read Shipment", "shipments.read", "shipments"},
	{"Create Employee", "employee.create", "employee"},
	{"Read Employee", "employee.read", "employee"},
	{"Update Employee", "employee.update", "employee"},
	{"Delete Employee", "employee.delete", "employee"},
	{"Create Branch", "branches.create", "branches"},
	{"Read Branch", "branches.read", "branches"},
	{"Update Branch", "branches.update", "branches"},
	{"Delete Branch", "branches.delete", "branches"},
	{"Read History", "history.read", "history"},
	{"Read Delivery", "delivery.read", "delivery"},
	{"Update Delivery", "delivery.update", "delivery"},
	{"Read Shipment Branch", "shipment-branch.read", "shipment-branch"},
	{"Input Shipment Branch", "shipment-branch.input", "shipment-branch"},
	{"Read Permissions", "permissions.read", "permissions"},
	{"Manage Permissions", "permissions.manage", "permissions"},
	{"Track Packages", "packages.track", "packages"},
	{"Scan Packages", "packages.scan", "packages"},
}

// rolePermissionsData mirrors role-permissions.json. Keys that don't exist as
// permissions (e.g. shipments.update/delete) are silently skipped, matching
// the original Prisma seeder behaviour.
var rolePermissionsData = map[string][]string{
	"super-admin": {
		"shipments.create", "shipments.read", "shipments.update", "shipments.delete",
		"employee.create", "employee.read", "employee.update", "employee.delete",
		"branches.create", "branches.read", "branches.update", "branches.delete",
		"history.read", "delivery.read", "delivery.update",
		"shipment-branch.input", "shipment-branch.read",
		"permissions.read", "permissions.manage", "packages.track", "packages.scan",
	},
	"admin-branch": {
		"shipments.create", "shipments.read", "shipments.update",
		"shipment-branch.input", "shipment-branch.read",
		"packages.track", "packages.scan",
		"employee.create", "employee.read", "employee.update", "employee.delete",
	},
	"courier":  {"delivery.read", "delivery.update", "packages.track"},
	"customer": {"shipments.create", "shipments.read", "packages.track"},
}

var usersData = []seedUser{
	{"Super Admin", "superadmin@mail.com", "password123", "0811111111", "super-admin"},
	{"Customer", "customer@mail.com", "password123", "0822222222", "customer"},
}

var branchesData = []seedBranch{
	{"Cabang Jakarta Pusat", "Jl. Sudirman No. 123, Jakarta Pusat, DKI Jakarta", "08121234567"},
	{"Cabang Surabaya", "Jl. Pemuda No. 45, Surabaya, Jawa Timur", "08129876543"},
	{"Cabang Bandung", "Jl. Dago No. 78, Bandung, Jawa Barat", "08125555666"},
	{"Cabang Malang", "Jl. Kertanegara No. 34, Malang, Jawa Timur", "08127777888"},
}

var employeeBranchesData = []seedEmployee{
	{"Courier Jakarta", "courier.jakarta@mail.com", "password123", "0833333333", "courier", "Cabang Jakarta Pusat", "courier"},
	{"Admin Branch Jakarta", "adminbranch.jakarta@mail.com", "password123", "0844444444", "admin-branch", "Cabang Jakarta Pusat", "admin"},
	{"Courier Surabaya", "courier.surabaya@mail.com", "password123", "0835555555", "courier", "Cabang Surabaya", "courier"},
	{"Admin Branch Surabaya", "adminbranch.surabaya@mail.com", "password123", "0846666666", "admin-branch", "Cabang Surabaya", "admin"},
	{"Courier Bandung", "courier.bandung@mail.com", "password123", "0837777777", "courier", "Cabang Bandung", "courier"},
	{"Admin Branch Bandung", "adminbranch.bandung@mail.com", "password123", "0848888888", "admin-branch", "Cabang Bandung", "admin"},
}

// Seed idempotently inserts roles, permissions, role-permissions, users,
// branches and employee-branches. Safe to run repeatedly.
func Seed(db *gorm.DB) error {
	// 1. Roles
	for _, r := range rolesData {
		var role models.Role
		if err := db.Where("key = ?", r.Key).First(&role).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&models.Role{Name: r.Name, Key: r.Key}).Error; err != nil {
				return err
			}
		}
	}

	// 2. Permissions
	for _, p := range permissionsData {
		var perm models.Permission
		if err := db.Where("key = ?", p.Key).First(&perm).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&models.Permission{Name: p.Name, Key: p.Key, Resource: p.Resource}).Error; err != nil {
				return err
			}
		}
	}

	// 3. Role-permission mappings
	for roleKey, permKeys := range rolePermissionsData {
		var role models.Role
		if err := db.Where("key = ?", roleKey).First(&role).Error; err != nil {
			log.Printf("[seed] role %s not found, skipping", roleKey)
			continue
		}
		for _, pk := range permKeys {
			var perm models.Permission
			if err := db.Where("key = ?", pk).First(&perm).Error; err != nil {
				continue // permission doesn't exist (e.g. shipments.update) — skip
			}
			var rp models.RolePermission
			err := db.Where("role_id = ? AND permission_id = ?", role.ID, perm.ID).First(&rp).Error
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&models.RolePermission{RoleID: role.ID, PermissionID: perm.ID}).Error; err != nil {
					return err
				}
			}
		}
	}

	// 4. Users (bcrypt cost 12, matching users-seed.ts)
	for _, u := range usersData {
		var role models.Role
		if err := db.Where("key = ?", u.RoleKey).First(&role).Error; err != nil {
			log.Printf("[seed] role %s not found, skipping user %s", u.RoleKey, u.Email)
			continue
		}
		if err := upsertUser(db, u.Name, u.Email, u.Password, u.Phone, role.ID, 12); err != nil {
			return err
		}
	}

	// 5. Branches
	for _, b := range branchesData {
		var branch models.Branch
		if err := db.Where("name = ?", b.Name).First(&branch).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&models.Branch{Name: b.Name, Address: b.Address, PhoneNumber: b.Phone}).Error; err != nil {
				return err
			}
		}
	}

	// 6. Employee branches (bcrypt cost 10, matching employee-branches-seed.ts)
	for _, e := range employeeBranchesData {
		var role models.Role
		if err := db.Where("key = ?", e.RoleKey).First(&role).Error; err != nil {
			log.Printf("[seed] role %s not found, skipping employee %s", e.RoleKey, e.Email)
			continue
		}
		var branch models.Branch
		if err := db.Where("name = ?", e.BranchName).First(&branch).Error; err != nil {
			log.Printf("[seed] branch %s not found, skipping employee %s", e.BranchName, e.Email)
			continue
		}
		if err := upsertUser(db, e.Name, e.Email, e.Password, e.Phone, role.ID, 10); err != nil {
			return err
		}
		var user models.User
		if err := db.Where("email = ?", e.Email).First(&user).Error; err != nil {
			return err
		}
		var eb models.EmployeeBranch
		err := db.Where("user_id = ? AND branch_id = ?", user.ID, branch.ID).First(&eb).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&models.EmployeeBranch{
				UserID: user.ID, BranchID: branch.ID, Type: e.Type,
			}).Error; err != nil {
				return err
			}
		}
	}

	log.Println("Database seeded")
	return nil
}

// upsertUser creates the user if their email is new; existing users are left
// untouched (matches prisma upsert with empty update).
func upsertUser(db *gorm.DB, name, email, password, phone string, roleID uint, cost int) error {
	var existing models.User
	if err := db.Where("email = ?", email).First(&existing).Error; err == nil {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return err
	}
	return db.Create(&models.User{
		Name:        name,
		Email:       email,
		Password:    string(hash),
		PhoneNumber: phone,
		RoleID:      roleID,
	}).Error
}
