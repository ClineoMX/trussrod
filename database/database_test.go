package database

import (
	"net/url"
	"strings"
	"testing"
)

func TestGetURL(t *testing.T) {
	tests := []struct {
		name       string
		user       string
		password   string
		driver     string
		host       string
		port       string
		dbName     string
		sslmode    string
		searchpath string
		wantScheme string
		wantHost   string
		wantPath   string
		wantUser   string
		wantPass   string
		wantSSL    string
		wantSearch string
	}{
		{
			name:       "with password",
			user:       "admin",
			password:   "secret",
			host:       "localhost",
			port:       "5432",
			dbName:     "app",
			sslmode:    "disable",
			wantScheme: "postgres",
			wantHost:   "localhost:5432",
			wantPath:   "/app",
			wantUser:   "admin",
			wantPass:   "secret",
			wantSSL:    "disable",
		},
		{
			name:       "without password",
			user:       "admin",
			host:       "db.example.com",
			port:       "5432",
			dbName:     "app",
			sslmode:    "require",
			wantScheme: "postgres",
			wantHost:   "db.example.com:5432",
			wantPath:   "/app",
			wantUser:   "admin",
			wantSSL:    "require",
		},
		{
			name:       "custom driver",
			user:       "admin",
			password:   "secret",
			driver:     "postgresql",
			host:       "localhost",
			port:       "5432",
			dbName:     "app",
			sslmode:    "disable",
			wantScheme: "postgresql",
			wantHost:   "localhost:5432",
			wantPath:   "/app",
			wantUser:   "admin",
			wantPass:   "secret",
			wantSSL:    "disable",
		},
		{
			name:       "with search path",
			user:       "admin",
			password:   "secret",
			host:       "localhost",
			port:       "5432",
			dbName:     "app",
			sslmode:    "disable",
			searchpath: "public,audit",
			wantScheme: "postgres",
			wantHost:   "localhost:5432",
			wantPath:   "/app",
			wantUser:   "admin",
			wantPass:   "secret",
			wantSSL:    "disable",
			wantSearch: "-c search_path=public,audit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getURL(tt.user, tt.password, tt.driver, tt.host, tt.port, tt.dbName, tt.sslmode, tt.searchpath)
			parsed, err := url.Parse(got)
			if err != nil {
				t.Fatalf("failed to parse URL %q: %v", got, err)
			}

			if parsed.Scheme != tt.wantScheme {
				t.Fatalf("expected scheme %q, got %q", tt.wantScheme, parsed.Scheme)
			}
			if parsed.Host != tt.wantHost {
				t.Fatalf("expected host %q, got %q", tt.wantHost, parsed.Host)
			}
			if parsed.Path != tt.wantPath {
				t.Fatalf("expected path %q, got %q", tt.wantPath, parsed.Path)
			}

			user := parsed.User.Username()
			if user != tt.wantUser {
				t.Fatalf("expected user %q, got %q", tt.wantUser, user)
			}

			pass, ok := parsed.User.Password()
			if tt.wantPass == "" {
				if ok {
					t.Fatalf("expected no password, got %q", pass)
				}
			} else if pass != tt.wantPass {
				t.Fatalf("expected password %q, got %q", tt.wantPass, pass)
			}

			q := parsed.Query()
			if q.Get("sslmode") != tt.wantSSL {
				t.Fatalf("expected sslmode %q, got %q", tt.wantSSL, q.Get("sslmode"))
			}

			if tt.wantSearch == "" {
				if q.Get("options") != "" {
					t.Fatalf("expected no search path option, got %q", q.Get("options"))
				}
			} else if !strings.Contains(q.Get("options"), tt.wantSearch) {
				t.Fatalf("expected options to contain %q, got %q", tt.wantSearch, q.Get("options"))
			}
		})
	}
}
