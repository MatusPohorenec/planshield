package redact

import (
	"testing"
)

func TestRedactSensitiveField(t *testing.T) {
	r := New(true, nil)

	got := r.RedactValue("my-secret-password", true)
	if got != "REDACTED" {
		t.Errorf("RedactValue(sensitive=true) = %v, want REDACTED", got)
	}
}

func TestRedactNonSensitiveClean(t *testing.T) {
	r := New(true, nil)

	got := r.RedactValue("hello-world", false)
	if got != "hello-world" {
		t.Errorf("RedactValue(clean string) = %v, want hello-world", got)
	}
}

func TestRedactPatternMatchAWSKey(t *testing.T) {
	r := New(true, nil)

	// AKIA followed by 16 uppercase alphanumeric chars
	got := r.RedactValue("AKIAIOSFODNN7EXAMPLE", false)
	if got != "REDACTED" {
		t.Errorf("RedactValue(AWS key) = %v, want REDACTED", got)
	}
}

func TestRedactPatternMatchGitHubToken(t *testing.T) {
	r := New(true, nil)

	got := r.RedactValue("ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmn", false)
	if got != "REDACTED" {
		t.Errorf("RedactValue(GitHub token) = %v, want REDACTED", got)
	}
}

func TestRedactPatternMatchPrivateKey(t *testing.T) {
	r := New(true, nil)

	got := r.RedactValue("-----BEGIN RSA PRIVATE KEY-----", false)
	if got != "REDACTED" {
		t.Errorf("RedactValue(private key) = %v, want REDACTED", got)
	}
}

func TestRedactMapWithSensitiveMarkers(t *testing.T) {
	r := New(false, nil)

	values := map[string]interface{}{
		"username": "admin",
		"password": "super-secret",
		"port":     float64(5432),
	}

	sensitiveMap := map[string]interface{}{
		"password": true,
	}

	result := r.RedactMap(values, sensitiveMap)

	if result["username"] != "admin" {
		t.Errorf("username = %v, want admin", result["username"])
	}
	if result["password"] != "REDACTED" {
		t.Errorf("password = %v, want REDACTED", result["password"])
	}
	if result["port"] != float64(5432) {
		t.Errorf("port = %v, want 5432", result["port"])
	}
}

func TestRedactMapNested(t *testing.T) {
	r := New(false, nil)

	values := map[string]interface{}{
		"config": map[string]interface{}{
			"host":     "db.example.com",
			"password": "nested-secret",
		},
	}

	sensitiveMap := map[string]interface{}{
		"config": map[string]interface{}{
			"password": true,
		},
	}

	result := r.RedactMap(values, sensitiveMap)

	config, ok := result["config"].(map[string]interface{})
	if !ok {
		t.Fatal("config should be a map")
	}
	if config["host"] != "db.example.com" {
		t.Errorf("config.host = %v, want db.example.com", config["host"])
	}
	if config["password"] != "REDACTED" {
		t.Errorf("config.password = %v, want REDACTED", config["password"])
	}
}

func TestRedactNoDefaults(t *testing.T) {
	r := New(false, nil)

	// AWS key should NOT be caught if defaults are disabled.
	got := r.RedactValue("AKIAIOSFODNN7EXAMPLE", false)
	if got != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("RedactValue(no defaults) = %v, want original value", got)
	}
}

func TestCompilePatternsValid(t *testing.T) {
	_, err := CompilePatterns([]string{`secret-\d+`, `password`})
	if err != nil {
		t.Errorf("CompilePatterns returned unexpected error: %v", err)
	}
}

func TestCompilePatternsInvalid(t *testing.T) {
	_, err := CompilePatterns([]string{`[invalid`})
	if err == nil {
		t.Error("CompilePatterns should return error for invalid regex")
	}
}
