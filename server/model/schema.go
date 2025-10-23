package model

// NOTE: Here register all models to be used in migration

// GetAllModels returns an array of all models
func GetAllModels() []any {
	return []any{
		&User{},
		&Session{},
	}
}
