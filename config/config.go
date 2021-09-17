package config

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend"
	"github.com/heetch/confita/backend/env"
	"github.com/heetch/confita/backend/file"
	"github.com/heetch/confita/backend/flags"

	"time"
)

type BlockCypher struct {
	Token string `config:"blockcypher_token" yaml:"blockcypher_token"`
}

type Metrics struct {
	Port int    `config:"metrics_port"`
	Host string `config:"metrics_host"`
}

type Logger struct {
	Host     string `config:"logger_host"`
	Port     int    `config:"logger_port"`
	Level    string `config:"logger_level"`
	CLILevel string `config:"cli_level"`
}

type Healthz struct {
	ReadTimeout  time.Duration `config:"healthz_read_timeout" yaml:"healthz_read_timeout"`
	WriteTimeout time.Duration `config:"healthz_write_timeout" yaml:"healthz_write_timeout"`
}

// AppConfiguration default config
// nolint
type Config struct {
	Name    string `default:"bitcoin wallet"`
	Version string `default:"0.0.0"`
	Port    int    `default:"8080"`
	Env     string `default:"integration"`
	Host    string `default:"0.0.0.0"`

	Metrics Metrics

	Logger Logger

	BlockCypher BlockCypher

	Healthz Healthz
}

func getDefaultConfig() *Config {
	return &Config{
		Name:    "kafka-overload",
		Version: "0.0.0",
		Host:    "0.0.0.0",
		Port:    8080,

		Metrics: Metrics{
			Port: 8081,
			Host: "0.0.0.0",
		},

		Logger: Logger{
			CLILevel: "INFO",
			Host:     "0.0.0.0",
			Port:     12201,
			Level:    "INFO",
		},

		BlockCypher: BlockCypher{
			Token: "",
		},

		Healthz: Healthz{
			ReadTimeout:  10,
			WriteTimeout: 10,
		},
	}
}

// New Load the config
func New() *Config {
	loaders := []backend.Backend{
		env.NewBackend(),
		flags.NewBackend(),
	}
	environment := os.Getenv("ENV")
	if environment != "" {
		configFile := findConfigFilePathRecursively(environment, 0)
		if configFile != "" {
			loaders = append(loaders, file.NewBackend(configFile))
		}
	}

	loader := confita.NewLoader(loaders...)

	cfg := getDefaultConfig()
	err := loader.Load(context.Background(), cfg)
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("%+v", cfg))
	return cfg
}

func findConfigFilePathRecursively(environment string, depth int) string {
	char := "../"
	if depth == 0 {
		char = "./"
	}
	if depth > 3 {
		return ""
	}

	filePath := strings.Repeat(char, depth) + "config/config." + environment + ".yaml"
	if _, err := os.Stat(filePath); err == nil {
		return filePath
	}
	depth++

	return findConfigFilePathRecursively(environment, depth)
}

func (c *Config) String() string {
	val := reflect.ValueOf(c).Elem()
	s := "\n-------------------------------\n"
	s += "-  Application configuration  -\n"
	s += "-------------------------------\n"
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		obfuscate := false

		tag := typeField.Tag.Get("config")
		if idx := strings.Index(tag, ","); idx != -1 {
			opts := strings.Split(tag[idx+1:], ",")

			for _, opt := range opts {
				if opt == "obfuscate" {
					obfuscate = true
				}
			}
		}
		if !obfuscate {
			switch {
			case typeField.Type.Kind() == reflect.String:
				s += fmt.Sprintf("%s: \"%v\"\n", typeField.Name, valueField.Interface())
				continue
			case typeField.Type.Kind() == reflect.Bool:
			case typeField.Type.Kind() == reflect.Int:
				s += fmt.Sprintf("%s: %v\n", typeField.Name, valueField.Interface())
				continue
			case typeField.Type.Kind() == reflect.Struct:
				s += c.DeepStructFields(typeField.Name, valueField.Interface())
			}
		}
	}
	return s
}

func (c *Config) DeepStructFields(parent string, iFace interface{}) string {
	s := ""
	ifv := reflect.ValueOf(iFace)
	ift := reflect.TypeOf(iFace)

	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		k := ift.Field(i)

		obfuscate := false

		tag := k.Tag.Get("config")
		if idx := strings.Index(tag, ","); idx != -1 {
			opts := strings.Split(tag[idx+1:], ",")

			for _, opt := range opts {
				if opt == "obfuscate" {
					obfuscate = true
				}
			}
		}
		if !obfuscate {

			switch v.Kind() {
			case reflect.String:
				s += fmt.Sprintf("%s: \"%v\"\n", parent+"-"+k.Name, v.Interface())
				continue
			case reflect.Bool:
			case reflect.Int:
				s += fmt.Sprintf("%s: %v\n", parent+"-"+k.Name, v.Interface())
				continue
			case reflect.Struct:
				s += c.DeepStructFields(parent+"-"+k.Name, v.Interface())
			}
		}
	}

	return s
}
