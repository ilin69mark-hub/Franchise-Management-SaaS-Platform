package database

import (
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedUsers(db *gorm.DB) error {
	users := []struct {
		Email     string
		Password  string
		Role      string
		FirstName string
		LastName  string
	}{
		{Email: "admin@mail.ru", Password: "qwerty", Role: "super_admin", FirstName: "Super", LastName: "Admin"},
		{Email: "fr@mail.ru", Password: "qwerty", Role: "franchiser", FirstName: "Franchise", LastName: "Owner"},
	}

	for _, u := range users {
		var count int64
		db.Table("users").Where("email = ?", u.Email).Count(&count)
		if count > 0 {
			log.Printf("User %s already exists, skipping", u.Email)
			continue
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		if u.Role == "franchiser" {
			var tenantCount int64
			db.Table("tenants").Count(&tenantCount)
			tenantUUID := "00000000-0000-0000-0000-000000000001"
			db.Exec(`INSERT INTO tenants (id, name, status) VALUES ($1, $2, 'active') ON CONFLICT DO NOTHING`,
				tenantUUID, u.FirstName+" "+u.LastName)
			db.Exec(`INSERT INTO users (id, email, password_hash, role, tenant_id, first_name, last_name)
				VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)`,
				u.Email, string(hashed), u.Role, tenantUUID, u.FirstName, u.LastName)
		} else {
			db.Exec(`INSERT INTO users (id, email, password_hash, role, first_name, last_name)
				VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)`,
				u.Email, string(hashed), u.Role, u.FirstName, u.LastName)
		}

		log.Printf("User created: %s (%s)", u.Email, u.Role)
	}

	return nil
}