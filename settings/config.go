package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type DatabaseConfig struct {
	Driver     string
	Host       string `json:"DB_HOST"`
	Port       string `json:"DB_PORT"`
	Name       string `json:"DB_NAME"`
	SSLMode    string
	User       string `json:"DB_USER"`
	Password   string `json:"DB_PASSWORD"`
	SearchPath string `json:"DB_SEARCHPATH"`
	MaxConns   int    `json:"DB_MAX_CONNS"`
}

type CacheConfig struct {
	Driver   string
	Host     string `json:"CACHE_HOST"`
	Port     string `json:"CACHE_PORT"`
	User     string `json:"CACHE_USER"`
	Password string `json:"CACHE_PASSWORD"`
}

type OAuthConfig struct {
	App              string `json:"OAUTH_APP_ID"`
	Flow             string `json:"OAUTH_AUTH_FLOW"`
	Issuer           string `json:"OAUTH_ISSUER"`
	Secret           string `json:"OAUTH_SECRET_ID"`
	UserPool         string `json:"OAUTH_USER_POOL"`
	PatientsUserPool string `json:"OAUTH_PATIENTS_USER_POOL"`
}

type CloudConfig struct {
	Region string
}

type EventsConfig struct {
	QueueURL string `json:"QUEUE_URL"`
}

type EncryptionConfig struct {
	Key string `json:"ENCRYPTION_KEY_ID"`
}

type IdentityConfig struct {
	UserPool     string `json:"USER_POOL"`
	IdentityPool string `json:"IDENTITY_POOL"`
}

type VectorConfig struct {
	Host string `json:"VECTORS_HOST"`
	Port string `json:"VECTORS_PORT"`
}

type StorageConfig struct {
	Bucket string
}

type DomedikConfig struct {
	Cache       *CacheConfig
	Cloud       *CloudConfig
	Crypto      *EncryptionConfig
	DB          *DatabaseConfig
	OAuth       *OAuthConfig
	Events      *EventsConfig
	Vectors     *VectorConfig
	Identity    *IdentityConfig
	Storage     *StorageConfig
	BindPort    string
	ApiKey      string
	Environment string
	LogLevel    string
}

func getFromProvider(deps []string) (*DomedikConfig, error) {
	ctx := context.Background()
	region := os.Getenv("DOMEDIK_REGION")
	port := os.Getenv("DOMEDIK_PORT")
	secretId := os.Getenv("SECRET_ID")

	// Remove after correct deployment
	apikey := os.Getenv("API_KEY")
	var environment string
	if value := os.Getenv("DOMEDIK_ENV"); value != "" {
		environment = strings.ToLower(value)
	} else {
		environment = "development"
	}
	var loglevel string
	if value := os.Getenv("DOMEDIK_LOGLEVEL"); value != "" {
		loglevel = strings.ToLower(value)
	} else {
		loglevel = "debug"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		panic("failed to load AWS config")
	}
	client := secretsmanager.NewFromConfig(cfg)

	secret, err := func(secretId string) (string, error) {
		input := &secretsmanager.GetSecretValueInput{
			SecretId: aws.String(secretId),
		}

		result, err := client.GetSecretValue(ctx, input)
		if err != nil {
			return "", err
		}

		if result.SecretString == nil {
			return "", err
		}

		return *result.SecretString, nil
	}(secretId)

	if err != nil {
		return nil, err
	}

	// Unmarshaling database config
	dbconf := DatabaseConfig{}
	if slices.Contains(deps, "database") {
		if err := json.Unmarshal([]byte(secret), &dbconf); err != nil {
			return nil, errors.New("failed to unmarshal database config")
		}
		dbconf.SSLMode = "require"
	}

	// Unmarshaling cache config
	cacheconf := CacheConfig{}
	if slices.Contains(deps, "cache") {
		if err := json.Unmarshal([]byte(secret), &cacheconf); err != nil {
			return nil, errors.New("failed to unmarshal cache config")
		}
	}

	// Unmarshaling oauth config
	oauthconf := OAuthConfig{}
	if slices.Contains(deps, "oauth") {
		if err := json.Unmarshal([]byte(secret), &oauthconf); err != nil {
			return nil, errors.New("failed to unmarshal oauth config")
		}
	}

	// Unmarshaling events config
	eventsconf := EventsConfig{}
	if slices.Contains(deps, "events") {
		if err := json.Unmarshal([]byte(secret), &eventsconf); err != nil {
			return nil, errors.New("failed to unmarshal oauth config")
		}
	}

	// Unmarshaling crypto config
	encconf := EncryptionConfig{}
	if slices.Contains(deps, "encryption") {
		if err := json.Unmarshal([]byte(secret), &encconf); err != nil {
			return nil, errors.New("failed to unmarshal oauth config")
		}
	}

	// Unmarshaling vector config
	vectorconf := VectorConfig{}
	if slices.Contains(deps, "vectors") {
		if err := json.Unmarshal([]byte(secret), &vectorconf); err != nil {
			return nil, errors.New("failed to unmarshal vector config")
		}
	}

	idconf := IdentityConfig{}
	if slices.Contains(deps, "identity") {
		if err := json.Unmarshal([]byte(secret), &idconf); err != nil {
			return nil, errors.New("failed to unmarshal id config")
		}
	}

	sconfig := StorageConfig{}
	if slices.Contains(deps, "storage") {
		if err := json.Unmarshal([]byte(secret), &sconfig); err != nil {
			return nil, errors.New("failed to unmarshal id config")
		}
	}

	return &DomedikConfig{
		BindPort:    port,
		ApiKey:      apikey,
		Environment: environment,
		LogLevel:    loglevel,
		Cloud:       &CloudConfig{Region: region},
		DB:          &dbconf,
		Cache:       &cacheconf,
		OAuth:       &oauthconf,
		Events:      &eventsconf,
		Vectors:     &vectorconf,
		Crypto:      &encconf,
		Identity:    &idconf,
		Storage:     &sconfig,
	}, nil
}

func getFromEnv(deps []string) *DomedikConfig {
	port := os.Getenv("DOMEDIK_PORT")
	region := os.Getenv("DOMEDIK_REGION")
	// Remove after correct deployment
	apikey := os.Getenv("DOMEDIK_API_KEY")
	var environment string
	if value := os.Getenv("DOMEDIK_ENV"); value != "" {
		environment = strings.ToLower(value)
	} else {
		environment = "development"
	}
	var loglevel string
	if value := os.Getenv("DOMEDIK_LOGLEVEL"); value != "" {
		loglevel = strings.ToLower(value)
	} else {
		loglevel = "debug"
	}

	dbconf := &DatabaseConfig{}

	maxConns, _ := strconv.Atoi(os.Getenv("DB_MAX_CONNS"))
	if slices.Contains(deps, "database") {
		dbconf.Driver = "postgres"
		dbconf.Host = os.Getenv("DB_HOST")
		dbconf.Port = os.Getenv("DB_PORT")
		dbconf.Name = os.Getenv("DB_NAME")
		dbconf.User = os.Getenv("DB_USER")
		dbconf.Password = os.Getenv("DB_PASSWORD")
		dbconf.SearchPath = os.Getenv("DB_SEARCHPATH")
		dbconf.MaxConns = maxConns
		dbconf.SSLMode = "disable"
	}

	cacheconf := &CacheConfig{}
	if slices.Contains(deps, "cache") {
		cacheconf.Driver = "redis"
		cacheconf.Host = os.Getenv("CACHE_HOST")
		cacheconf.Port = os.Getenv("CACHE_PORT")
		cacheconf.User = os.Getenv("CACHE_USER")
		cacheconf.Password = os.Getenv("CACHE_PASSWORD")
	}

	oauthconf := &OAuthConfig{}
	if slices.Contains(deps, "oauth") {
		oauthconf.App = os.Getenv("OAUTH_APP_ID")
		oauthconf.Flow = os.Getenv("OAUTH_AUTH_FLOW")
		oauthconf.Issuer = os.Getenv("OAUTH_ISSUER")
		oauthconf.Secret = os.Getenv("OAUTH_SECRET_ID")
		oauthconf.UserPool = os.Getenv("OAUTH_USER_POOL")
		oauthconf.PatientsUserPool = os.Getenv("OAUTH_PATIENTS_USER_POOL")
	}

	encconf := &EncryptionConfig{}
	if slices.Contains(deps, "encryption") {
		encconf.Key = os.Getenv("ENCRYPTION_KEY_ID")
	}

	eventsconf := &EventsConfig{}
	if slices.Contains(deps, "events") {
		eventsconf.QueueURL = os.Getenv("QUEUE_URL")
	}

	vectorconf := &VectorConfig{}
	if slices.Contains(deps, "events") {
		vectorconf.Host = os.Getenv("VECTORS_HOST")
		vectorconf.Port = os.Getenv("VECTORS_PORT")
	}

	idconf := IdentityConfig{}
	if slices.Contains(deps, "identity") {
		idconf.UserPool = os.Getenv("USER_POOL")
		idconf.IdentityPool = os.Getenv("IDENTITY_POOL")
	}

	sconf := StorageConfig{}
	if slices.Contains(deps, "storage") {
		sconf.Bucket = os.Getenv("STORAGE_BUCKET")
	}

	return &DomedikConfig{
		ApiKey:      apikey,
		BindPort:    port,
		Environment: environment,
		LogLevel:    loglevel,
		DB:          dbconf,
		Cache:       cacheconf,
		OAuth:       oauthconf,
		Crypto:      encconf,
		Events:      eventsconf,
		Vectors:     vectorconf,
		Identity:    &idconf,
		Cloud:       &CloudConfig{Region: region},
		Storage:     &sconf,
	}
}

func Resolve(deps []string) (*DomedikConfig, error) {
	env := os.Getenv("DOMEDIK_ENV")
	fmt.Print(env)
	if env == "prod" {
		return getFromProvider(deps)
	} else {
		return getFromEnv(deps), nil
	}
}
