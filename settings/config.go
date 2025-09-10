package settings

import (
	"context"
	"encoding/json"
	"os"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type DatabaseConfig struct {
	Driver   string
	Host     string `json:"DB_HOST"`
	Port     string `json:"DB_PORT"`
	Name     string `json:"DB_NAME"`
	SSLMode  string
	User     string `json:"DB_USER"`
	Password string `json:"DB_PASSWORD"`
}

type CacheConfig struct {
	Driver   string
	Host     string `json:"CACHE_HOST"`
	Port     string `json:"CACHE_PORT"`
	User     string `json:"CACHE_USER"`
	Password string `json:"CACHE_PASSWORD"`
}

type OAuthConfig struct {
	App      string `json:"OAUTH_APP_ID"`
	Flow     string `json:"OAUTH_AUTH_FLOW"`
	Issuer   string `json:"OAUTH_ISSUER"`
	Secret   string `json:"OAUTH_SECRET_ID"`
	UserPool string `json:"OAUTH_USER_POOL"`
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

type VectorConfig struct {
	Host string `json:"VECTORS_HOST"`
	Port string `json:"VECTORS_PORT"`
}

type DomedikConfig struct {
	Cache    *CacheConfig
	Cloud    *CloudConfig
	Crypto   *EncryptionConfig
	DB       *DatabaseConfig
	OAuth    *OAuthConfig
	Events   *EventsConfig
	Vectors  *VectorConfig
	BindPort string
	ApiKey   string
}

func getFromProvider(deps []string) *DomedikConfig {
	ctx := context.Background()
	region := os.Getenv("DOMEDIK_REGION")
	port := os.Getenv("DOMEDIK_PORT")
	secretId := os.Getenv("SECRET_ID")

	// Remove after correct deployment
	apikey := os.Getenv("API_KEY")

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		panic("failed to load AWS config")
	}
	client := secretsmanager.NewFromConfig(cfg)

	secret := func(secretId string) string {
		input := &secretsmanager.GetSecretValueInput{
			SecretId: aws.String(secretId),
		}

		result, err := client.GetSecretValue(ctx, input)
		if err != nil {
			panic("failed to get AWS secret: " + secretId)
		}

		if result.SecretString == nil {
			panic("secret string is nil: " + secretId)
		}

		return *result.SecretString
	}(secretId)

	// Unmarshaling database config
	var dbconf *DatabaseConfig
	if slices.Contains(deps, "database") {
		if err := json.Unmarshal([]byte(secret), &dbconf); err != nil {
			panic("failed to unmarshal database config")
		}
		dbconf.SSLMode = "require"
	}

	// Unmarshaling cache config
	var cacheconf *CacheConfig
	if slices.Contains(deps, "cache") {
		if err := json.Unmarshal([]byte(secret), &cacheconf); err != nil {
			panic("failed to unmarshal cache config")
		}
	}

	// Unmarshaling oauth config
	var oauthconf *OAuthConfig
	if slices.Contains(deps, "oauth") {
		if err := json.Unmarshal([]byte(secret), &oauthconf); err != nil {
			panic("failed to unmarshal oauth config")
		}
	}

	// Unmarshaling events config
	var eventsconf *EventsConfig
	if slices.Contains(deps, "events") {
		if err := json.Unmarshal([]byte(secret), &eventsconf); err != nil {
			panic("failed to unmarshal oauth config")
		}
	}

	// Unmarshaling crypto config
	var encconf *EncryptionConfig
	if slices.Contains(deps, "encryption") {
		if err := json.Unmarshal([]byte(secret), &encconf); err != nil {
			panic("failed to unmarshal oauth config")
		}
	}

	// Unmarshaling vector config
	var vectorconf *VectorConfig
	if slices.Contains(deps, "vectors") {
		if err := json.Unmarshal([]byte(secret), &vectorconf); err != nil {
			panic("failed to unmarshal vector config")
		}
	}

	return &DomedikConfig{
		BindPort: port,
		ApiKey:   apikey,
		Cloud:    &CloudConfig{Region: region},
		DB:       dbconf,
		Cache:    cacheconf,
		OAuth:    oauthconf,
		Events:   eventsconf,
		Vectors:  vectorconf,
		Crypto:   encconf,
	}
}

func getFromEnv(deps []string) *DomedikConfig {
	port := os.Getenv("DOMEDIK_PORT")
	region := os.Getenv("DOMEDIK_REGION")
	// Remove after correct deployment
	apikey := os.Getenv("API_KEY")

	dbconf := &DatabaseConfig{}
	if slices.Contains(deps, "database") {
		dbconf.Driver = "postgres"
		dbconf.Host = os.Getenv("DB_HOST")
		dbconf.Port = os.Getenv("DB_PORT")
		dbconf.Name = os.Getenv("DB_NAME")
		dbconf.User = os.Getenv("DB_USER")
		dbconf.Password = os.Getenv("DB_PASSWORD")
		dbconf.SSLMode = "disable"
	}

	cacheconf := &CacheConfig{}
	if slices.Contains(deps, "cache") {
		cacheconf.Driver = "redis"
		cacheconf.Host = os.Getenv("DB_HOST")
		cacheconf.Port = os.Getenv("DB_PORT")
		cacheconf.User = os.Getenv("DB_USER")
		cacheconf.Password = os.Getenv("DB_PASSWORD")
	}

	oauthconf := &OAuthConfig{}
	if slices.Contains(deps, "oauth") {
		oauthconf.App = os.Getenv("COGNITO_APP_ID")
		oauthconf.Flow = os.Getenv("COGNITO_AUTH_FLOW")
		oauthconf.Issuer = os.Getenv("COGNITO_ISSUER")
		oauthconf.Secret = os.Getenv("COGNITO_SECRET_ID")
		oauthconf.UserPool = os.Getenv("COGNITO_USER_POOL")
	}

	encconf := &EncryptionConfig{}
	if slices.Contains(deps, "encryption") {
		encconf.Key = os.Getenv("KMS_KEY")
	}

	eventsconf := &EventsConfig{}
	if slices.Contains(deps, "events") {
		eventsconf.QueueURL = os.Getenv("SQS_URL")
	}

	vectorconf := &VectorConfig{}
	if slices.Contains(deps, "events") {
		vectorconf.Host = os.Getenv("QDRANT_HOST")
		vectorconf.Port = os.Getenv("QDRANT_PORT")
	}

	return &DomedikConfig{
		ApiKey:   apikey,
		BindPort: port,
		DB:       dbconf,
		Cache:    cacheconf,
		OAuth:    oauthconf,
		Crypto:   encconf,
		Events:   eventsconf,
		Vectors:  vectorconf,
		Cloud:    &CloudConfig{Region: region},
	}
}

func Resolve(deps []string) *DomedikConfig {
	env := os.Getenv("DOMEDIK_ENV")
	if env == "prod" {
		return getFromProvider(deps)
	} else {
		return getFromEnv(deps)
	}
}
