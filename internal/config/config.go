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
}

var defaultConfig = Config{
	ServerURL:       "http://localhost:8080",
	GRPCAddress:     ":50051",
	MigrationsDir:   "migrations",
	GRPCDialTimeout: 5 * time.Second,
}

type initOptions struct {
	disableFlagsParsing bool
}

// New loads the application configuration from the environment and command-line flags,
// applying optional InitOptions. Returns a validated Config instance.
func New(optionsProto ...InitOption) (*Config, error) {
	options := &initOptions{
		disableFlagsParsing: false,
	}
	for _, protoOption := range optionsProto {
		protoOption(options)
	}

	config := &Config{}

	parseJSON(config)

	err := parseENV(config)
	if err != nil {
		return nil, err
	}

	if !options.disableFlagsParsing {
		parseFlags(config)
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

// InitOption is a functional option type for configuring New().
type InitOption func(*initOptions)

// WithDisableFlagsParsing disables parsing of command-line flags when creating a Config.
func WithDisableFlagsParsing(disableFlagsParsing bool) InitOption {
	return func(options *initOptions) {
		options.disableFlagsParsing = disableFlagsParsing
	}
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

func parseFlags(config *Config) {
	//flag.StringVar(&config.RunAddr, "a", config.RunAddr, "address and port to run server")
	//flag.StringVar(&config.ShortURLBase, "b", config.ShortURLBase, "base address of the resulting shortened URL")
	//flag.StringVar(&config.LogLevel, "l", config.LogLevel, "logger level")
	//flag.StringVar(&config.DBFileName, "f", config.DBFileName, "JSON file name with database")
	//flag.StringVar(&config.DatabaseDSN, "d", config.DatabaseDSN, "A string with the database connection details")
	//flag.BoolVar(&config.EnableHTTPS, "s", config.EnableHTTPS, "HTTPS enabling flag")
	//
	//JSONConfigFilePathDesc := "JSON configuration file path"
	//flag.StringVar(&config.JSONConfigFilePath, "c", config.JSONConfigFilePath, JSONConfigFilePathDesc)
	//flag.StringVar(&config.JSONConfigFilePath, "config", config.JSONConfigFilePath, JSONConfigFilePathDesc)
	//
	//flag.StringVar(&config.TrustedSubnet, "t", config.TrustedSubnet, "CIDR for the trusted subnet")
	//
	//flag.Parse()
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
