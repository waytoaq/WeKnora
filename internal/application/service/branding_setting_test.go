package service

import (
	"strings"
	"testing"
)

func TestValidateBrandingSystemName(t *testing.T) {
	if err := validateBrandingSystemName(""); err != nil {
		t.Fatalf("empty name should be allowed: %v", err)
	}
	if err := validateBrandingSystemName("知识"); err != nil {
		t.Fatalf("short name should be allowed: %v", err)
	}
	if err := validateBrandingSystemName(strings.Repeat("名", brandingSystemNameMaxRunes+1)); err == nil {
		t.Fatal("overlong name should be rejected")
	}
	if err := validateBrandingSystemName("bad\nname"); err == nil {
		t.Fatal("control characters should be rejected")
	}
}

func TestValidateBrandingLogo(t *testing.T) {
	if err := validateBrandingLogo(""); err != nil {
		t.Fatalf("empty logo should be allowed: %v", err)
	}
	if err := validateBrandingLogo("data:image/png;base64,AAAA"); err != nil {
		t.Fatalf("data URI should be allowed: %v", err)
	}
	if err := validateBrandingLogo("https://example.com/logo.png"); err != nil {
		t.Fatalf("https URL should be allowed: %v", err)
	}
	if err := validateBrandingLogo("ftp://example.com/logo.png"); err == nil {
		t.Fatal("non-http schemes should be rejected")
	}
	if err := validateBrandingLogo(strings.Repeat("a", brandingLogoMaxChars+1)); err == nil {
		t.Fatal("oversized logo should be rejected")
	}
}

func TestBrandingKeysRegistered(t *testing.T) {
	for _, key := range []string{"branding.system_name", "branding.logo"} {
		spec, ok := registry[key]
		if !ok {
			t.Fatalf("missing registry key %q", key)
		}
		if spec.Type != "string" {
			t.Fatalf("%s type = %q, want string", key, spec.Type)
		}
		if spec.Category != "branding" {
			t.Fatalf("%s category = %q, want branding", key, spec.Category)
		}
	}
}
