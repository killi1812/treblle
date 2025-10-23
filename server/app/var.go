package app

const (
	BuildDev  = "dev"
	BuildProd = "prod"
)

var (
	// Build describes app build type
	//
	// Could be: dev, prod
	Build = BuildDev
	// Version is a semver version of the app
	Version = "0.0.0"
	// CommitHash is latest build commit hash
	CommitHash = "n/a"
	// BuildTimestamp stores when the app was build
	BuildTimestamp = "n/a"
)

// Envirment variables

var (
	Port      int    // Port is app port
	DbConn    string // Postgress Connection string
	MongoConn string // MongoConn is mongo db connection string
)
