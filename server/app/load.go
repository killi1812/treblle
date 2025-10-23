package app

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// LoadConfig loads in program configuration should be a first thing called in the program
func LoadConfig() {
	zap.S().Debugf("Loading env variables")

	if err := godotenv.Load("../.env"); err != nil {
		zap.S().DPanicf("Env load err = %+v\n", err)
		zap.S().Infof("Can't load config using real env")
	}

	// App config
	Port = loadInt("PORT")

	// Database
	DbConn = loadString("DB_CONN")
	MongoConn = loadString("MONGO_CONN")

	zap.S().Debugf("Finished loading env variables")
}

func loadInt(name string) int {
	rez := os.Getenv(name)
	if rez == "" {
		zap.S().Errorf("Env variable %s is empty\n", name)
	}

	num, err := strconv.Atoi(rez)
	if err != nil {
		zap.S().Errorf("Failed to parse int %s, will use default (0)\n", rez)
		return 0
	}

	zap.S().Debugf("Loaded %s = %d", name, num)
	return num
}

func loadString(name string) string {
	rez := strings.TrimSpace(os.Getenv(name))
	if rez == "" {
		zap.S().Errorf("Env variable %s is empty", name)
		return rez
	}
	zap.S().Debugf("Loaded %s = %s", name, rez)
	return rez
}

func loadBool(name string) bool {
	rez := os.Getenv(name)
	if rez == "" {
		zap.S().Errorf("Env variable %s is empty", name)
	}
	return rez == "true"
}
