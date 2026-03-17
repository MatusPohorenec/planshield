package redact

import (
	"regexp"
)

// DefaultPatterns returns built-in regex patterns that catch common secret-shaped
// tokens that providers may leak into plan values.
func DefaultPatterns() []*regexp.Regexp {
	raw := []string{
		// AWS-style access key IDs
		`(?i)AKIA[0-9A-Z]{16}`,
		// Generic long hex/base64 secrets (40+ chars)
		`(?i)(?:secret|password|token|key|api_key|apikey|auth)[\s]*[:=]\s*["']?[A-Za-z0-9/+=]{20,}`,
		// Connection strings with credentials
		`(?i)(?:mysql|postgres|mongodb|redis|amqp)://[^\s@]+:[^\s@]+@`,
		// Private key markers
		`-----BEGIN (?:RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`,
		// GitHub/GitLab tokens
		`(?:ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9_]{36,}`,
		`glpat-[A-Za-z0-9_\-]{20,}`,
		// Generic Bearer tokens
		`(?i)Bearer\s+[A-Za-z0-9\-._~+/]+=*`,
	}

	patterns := make([]*regexp.Regexp, 0, len(raw))
	for _, r := range raw {
		patterns = append(patterns, regexp.MustCompile(r))
	}
	return patterns
}

// CompilePatterns compiles a list of user-supplied regex strings.
// Returns an error on the first invalid pattern.
func CompilePatterns(rawPatterns []string) ([]*regexp.Regexp, error) {
	compiled := make([]*regexp.Regexp, 0, len(rawPatterns))
	for _, p := range rawPatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, err
		}
		compiled = append(compiled, re)
	}
	return compiled, nil
}
