package app

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v3"

	"github.com/vladislav-chunikhin/lib-go/pkg/logger"
)

const (
	EnvAppName    = "APP_NAME"
	EnvAppVersion = "APP_VERSION"
	EnvEnvName    = "ENV_NAME"
	EnvAppPort    = "APP_PORT"
	EnvAppMtPort  = "APP_MAINTENANCE_PORT"
	EnvConfigFile = "CONFIG_FILE"

	defaultPublicPort = "8080"
	defaultMtPort     = "8081"

	defaultHTTPServerReadTimeout  = 10
	defaultHTTPServerWriteTimeout = 20

	defaultFileExt = "yaml"
)

type Config struct {
	AppName        string
	AppEnvironment string
	AppVersion     string
	LoggerLevel    string
	Port           string
	MtPort         string

	ConfigFile string

	HTTPServerReadTimeout  int64
	HTTPServerWriteTimeout int64
}

// configFromEnv creates a configuration structure with data from environment variables
func createFromEnv() *Config {
	config := &Config{
		LoggerLevel:            logger.InfoLevel,
		Port:                   defaultPublicPort,
		MtPort:                 defaultMtPort,
		HTTPServerReadTimeout:  defaultHTTPServerReadTimeout,
		HTTPServerWriteTimeout: defaultHTTPServerWriteTimeout,
	}

	if v, ok := os.LookupEnv(EnvAppName); ok {
		config.AppName = v
	}

	if v, ok := os.LookupEnv(EnvAppVersion); ok {
		config.AppVersion = v
	}

	if v, ok := os.LookupEnv(EnvEnvName); ok {
		config.AppEnvironment = v
	}

	if port, ok := os.LookupEnv(EnvAppPort); ok {
		config.Port = port
	}

	if portMt, ok := os.LookupEnv(EnvAppMtPort); ok {
		config.MtPort = portMt
	}

	if v, ok := os.LookupEnv(EnvConfigFile); ok {
		config.ConfigFile = v
	}

	return config
}

func configure(configPtr interface{}, configFilePath string) error {
	{ // target type validation
		t := reflect.TypeOf(configPtr)
		if t.Kind() != reflect.Ptr {
			return errors.New("configPtr arg must be pointer")
		}
		if t.Elem().Kind() != reflect.Struct {
			return errors.New("configPtr arg must be pointer to struct")
		}
	}

	loadedFromFile, err := loadFromFile(configPtr, configFilePath)
	if err != nil {
		return err
	}

	if !loadedFromFile {
		return errors.New("cannot load from config")
	}

	return nil
}

func loadFromFile(
	target interface{},
	configPath string,
) (bool, error) {
	if err := fileExists(configPath); err != nil {
		return false, nil
	}

	if err := checkFileExt(configPath); err != nil {
		return false, err
	}

	file, err := os.Open(configPath)
	if err != nil {
		return false, err
	}
	if err = yaml.NewDecoder(file).Decode(target); err != nil {
		return false, err
	}

	return true, nil
}

func fileExists(path string) (err error) {
	if path == "" {
		return errors.New("empty path")
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return
	}
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("file not exists: " + absPath)
		}
		return
	}
	if info.IsDir() {
		return errors.New("must be file: " + absPath)
	}
	return
}

func checkFileExt(path string) error {
	if path == "" {
		return errors.New("empty path")
	}
	ext := filepath.Ext(path)
	if len(ext) > 1 {
		ext = ext[1:]
	}

	if ext != defaultFileExt {
		return errors.New("config file must be yaml")
	}

	return nil
}
