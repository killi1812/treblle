package seed

import "go.uber.org/zap"

// Insert will run seed functions and insert seeded data to database
func Insert() {
	if err := createSuperAdmin(); err != nil {
		zap.S().Panicf("Failed to create superadmin, err = %+v", err)
	}
}
