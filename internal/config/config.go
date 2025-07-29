package config

import (
	"encoding/json"
	"time"

	"log"
	"os"
	"reflect"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"

	"github.com/go-playground/validator/v10"
)

// Config holds the application configuration loaded from environment variables
// and optionally overridden by command-line flags.
type Config struct {
	ServerURL       string        `env:"SERVER_URL" validate:"url" json:"server_url"`
	GRPCAddress     string        `env:"GRPC_ADDRESS" json:"grpc_address"`
	MigrationsDir   string        `env:"MIGRATIONS_DIR" json:"migrations_dir"`
	DatabaseDSN     string        `env:"DATABASE_DSN" json:"database_dsn"`
	GRPCDialTimeout time.Duration `env:"GRPC_DIAL_TIMEOUT" json:"grpc_dial_timeout"`
	JWTSecret       string        `env:"JWT_SECRET" json:"jwt_secret"`
}

var defaultConfig = Config{
	ServerURL:       "http://localhost:8080",
	GRPCAddress:     ":50051",
	MigrationsDir:   "migrations",
	GRPCDialTimeout: 5 * time.Second,
}

// New loads the application configuration from the environment,
// Returns a validated Config instance.
func New() (*Config, error) {
	config := &Config{}

	parseJSON(config)

	err := parseENV(config)
	if err != nil {
		return nil, err
	}

	applyDefaults(config, defaultConfig)

	return config, config.Validate()
}

// Validate validates the configuration struct fields using custom and built-in rules.
func (conf *Config) Validate() error {
	validate := validator.New()

	err := validate.RegisterValidation("loglevel", validateLogLevel)
	if err != nil {
		return err
	}

	err = validate.RegisterValidation("filepath", validateFilePath)
	if err != nil {
		return err
	}

	return validate.Struct(conf)
}

func validateFilePath(fieldLevel validator.FieldLevel) bool {
	path := fieldLevel.Field().String()
	_, err := os.Stat(path)

	return err == nil || os.IsNotExist(err)
}

func validateLogLevel(fieldLevel validator.FieldLevel) bool {
	value := fieldLevel.Field().String()

	allowedLogLevels := map[string]bool{
		"debug":   true,
		"info":    true,
		"warning": true,
		"error":   true,
		"fatal":   true,
	}

	return allowedLogLevels[value]
}

func applyDefaults(config *Config, defaultConfig Config) {
	vCfg := reflect.ValueOf(config).Elem()
	vDef := reflect.ValueOf(defaultConfig)

	for i := 0; i < vCfg.NumField(); i++ {
		field := vCfg.Field(i)
		defVal := vDef.Field(i)

		if field.CanSet() && field.IsZero() {
			field.Set(defVal)
		}
	}
}

func parseENV(config *Config) error {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Unable to load .env file: %v", err)
	}

	err = env.Parse(config)
	if err != nil {
		return err
	}

	return nil
}

func parseJSON(values *Config) {
	var configFileName string
	if val := os.Getenv("CONFIG"); val != "" {
		configFileName = val
	}
	for i, arg := range os.Args {
		if (arg == "-c" || arg == "-config") && i+1 < len(os.Args) {
			configFileName = os.Args[i+1]
		}
	}

	if file, err := os.Open(configFileName); err == nil {
		defer file.Close()
		if err := json.NewDecoder(file).Decode(values); err != nil {
			log.Printf("Failed to parse %s: %v", configFileName, err)
		} else {
			log.Printf("Loaded config from %s", configFileName)
		}
	} else if !os.IsNotExist(err) {
		log.Printf("Failed to open %s: %v", configFileName, err)
	}
}
