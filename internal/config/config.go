package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
)

var (
	k          = koanf.New(".")
	once       sync.Once
	configPath string
)

func Load(path string) error {
	var loadErr error

	once.Do(func() {
		configPath = path
		provider := file.Provider(path)
		parser := yaml.Parser()

		if err := k.Load(provider, parser); err != nil {
			loadErr = fmt.Errorf("erro ao ler config: %w", err)
		}
	})
	return loadErr
}

func Get() *koanf.Koanf {
	return k
}

type Config struct {
	Server            ServerConfig         `yaml:"server" json:"server"`
	Session           SessionConfig        `yaml:"session" json:"session"`
	Campaign          CampaignConfig       `yaml:"campaign" json:"campaign"`
	EmailScheduler    EmailSchedulerConfig `yaml:"email_scheduler" json:"email_scheduler"`
	Security          SecurityConfig       `yaml:"security" json:"security"`
	TemplateDir       string               `yaml:"template_dir" json:"template_dir"`
	TemplateAssetsDir string               `yaml:"template_assets_dir" json:"template_assets_dir"`
}

type ServerConfig struct {
	Host          string `yaml:"host" json:"host"`
	DashboardPort int    `yaml:"dashboard_port" json:"dashboard_port"`
	APIport       int    `yaml:"api_port" json:"api_port"`
	CampaignPort  int    `yaml:"campaign_port" json:"campaign_port"`
}

type SessionConfig struct {
	CookieName     string `yaml:"cookie_name" json:"cookie_name"`
	CookieDomain   string `yaml:"cookie_domain" json:"cookie_domain"`
	CookieSecure   bool   `yaml:"cookie_secure" json:"cookie_secure"`
	CookieHTTPOnly bool   `yaml:"cookie_http_only" json:"cookie_http_only"`
	TTL            string `yaml:"ttl" json:"ttl"`
}

type CampaignConfig struct {
	BaseDomain    string `yaml:"base_domain" json:"base_domain"`
	SubdomainMode bool   `yaml:"subdomain_mode" json:"subdomain_mode"`
}

type EmailSchedulerConfig struct {
	Enabled              bool `yaml:"enabled" json:"enabled"`
	PollIntervalSeconds  int  `yaml:"poll_interval_seconds" json:"poll_interval_seconds"`
	EmailsPerMinute      int  `yaml:"emails_per_minute" json:"emails_per_minute"`
	BatchSize            int  `yaml:"batch_size" json:"batch_size"`
	BatchPauseMS         int  `yaml:"batch_pause_ms" json:"batch_pause_ms"`
	MaxParallelCampaigns int  `yaml:"max_parallel_campaigns" json:"max_parallel_campaigns"`
}

type SecurityConfig struct {
	TestModeToken string `yaml:"test_mode_token" json:"test_mode_token"`
	JwtSecret     string `yaml:"jwt_secret,omitempty" json:"jwt_secret,omitempty"`
}

var configStruct Config

func GetConfigStruct() *Config {
	return &configStruct
}

func GetBool(key string) bool {
	return k.Bool(key)
}

func GetInt(key string) int {
	return k.Int(key)
}

func GetString(key string) string {
	return k.String(key)
}

func GetKeys(section string) []string {
	val := k.Get(section)
	if subMap, ok := val.(map[string]interface{}); ok {
		keys := make([]string, 0, len(subMap))
		for k := range subMap {
			keys = append(keys, k)
		}
		return keys
	}
	return nil
}

func GetStringSlice(key string) []string {
	return k.Strings(key)
}

func GetStringMap(key string) map[string]interface{} {
	val := k.Get(key)
	if m, ok := val.(map[string]interface{}); ok {
		return m
	}
	return nil
}

func GetConfigField(path string) interface{} {
	return k.Get(path)
}

func SetConfigField(key string, value interface{}) error {
	path := strings.Split(key, ".")
	fieldType, err := getFieldTypeByYAMLTag(Config{}, path)
	if err != nil {
		return err
	}

	convertedValue, err := convertValue(fieldType, value)
	if err != nil {
		return err
	}

	k.Set(key, convertedValue)
	return saveToDisk()
}

func SetConfig(key string, value interface{}) error {
	k.Set(key, value)
	return saveToDisk()
}

func SaveStructConfig(baseKey string, data map[string]interface{}) error {
	for key, value := range data {
		fullKey := baseKey + "." + key
		if err := SetConfigField(fullKey, value); err != nil {
			return err
		}
	}
	return nil
}

func getFieldTypeByYAMLTag(structVal interface{}, fieldPath []string) (reflect.Type, error) {
	currentType := reflect.TypeOf(structVal)

	for i := 0; i < len(fieldPath); i++ {
		yamlField := fieldPath[i]

		if currentType.Kind() == reflect.Ptr {
			currentType = currentType.Elem()
		}

		if currentType.Kind() != reflect.Struct {
			return nil, fmt.Errorf("type %s is not a struct", currentType.Name())
		}

		found := false
		for j := 0; j < currentType.NumField(); j++ {
			field := currentType.Field(j)
			tag := field.Tag.Get("yaml")
			tagName := strings.Split(tag, ",")[0]

			if tagName == yamlField {
				if field.Type.Kind() == reflect.Map {
					if i+1 >= len(fieldPath) {
						return nil, fmt.Errorf("expected a key after map '%s'", yamlField)
					}
					currentType = field.Type.Elem()
					i++
					found = true
					break
				}

				currentType = field.Type
				found = true
				break
			}
		}

		if !found {
			if _, ok := currentType.FieldByName("Extras"); ok {
				return reflect.TypeOf(new(interface{})).Elem(), nil
			}
			return nil, fmt.Errorf("yaml field '%s' not found in type %s", yamlField, currentType.Name())
		}
	}

	return currentType, nil
}

func convertValue(fieldType reflect.Type, value interface{}) (interface{}, error) {
	switch fieldType.Kind() {
	case reflect.String:
		return fmt.Sprintf("%v", value), nil

	case reflect.Int:
		switch v := value.(type) {
		case string:
			return strconv.Atoi(v)
		case float64:
			return int(v), nil
		}
		return value, nil

	case reflect.Bool:
		if strVal, ok := value.(string); ok {
			return strconv.ParseBool(strVal)
		}
		return value, nil

	case reflect.Slice:
		if fieldType.Elem().Kind() == reflect.String {
			if rawList, ok := value.([]interface{}); ok {
				strList := make([]string, len(rawList))
				for i, item := range rawList {
					strList[i] = fmt.Sprintf("%v", item)
				}
				return strList, nil
			}
		}
		return value, nil

	case reflect.Map, reflect.Struct:
		return value, nil

	case reflect.Interface:
		return value, nil

	default:
		return nil, fmt.Errorf("unsupported kind: %v", fieldType.Kind())
	}
}

func SetConfigFromKey(key string, data interface{}) error {
	path := strings.Split(key, ".")
	fieldType, err := getFieldTypeByYAMLTag(data, path)
	if err != nil {
		return err
	}

	convertedValue, err := convertValue(fieldType, data)
	if err != nil {
		return err
	}

	k.Set(key, convertedValue)
	return saveToDisk()
}

func saveToDisk() error {
	raw, err := k.Marshal(yaml.Parser())
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, raw, 0644)
}
