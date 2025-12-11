package spu

import "strings"

// LocaleContent captures localized fields collected from the UI.
type LocaleContent struct {
	Locale      string         `json:"locale"`
	Title       string         `json:"title"`
	Subtitle    string         `json:"subtitle,omitempty"`
	Description string         `json:"description,omitempty"`
	Attributes  map[string]any `json:"attributes,omitempty"`
}

// LocaleService provides normalization and validation helpers for per-locale payloads.
type LocaleService struct {
	required []string
}

// NewLocaleService constructs a validator that enforces the provided required locales.
func NewLocaleService(required []string) *LocaleService {
	clean := make([]string, 0, len(required))
	for _, locale := range required {
		if trimmed := strings.TrimSpace(locale); trimmed != "" {
			clean = append(clean, trimmed)
		}
	}
	if len(clean) == 0 {
		clean = []string{"zh-CN"}
	}
	return &LocaleService{required: clean}
}

// Normalize trims whitespace and removes empty attributes helper.
func (s *LocaleService) Normalize(locales []LocaleContent) []LocaleContent {
	for i := range locales {
		locales[i].Locale = strings.TrimSpace(locales[i].Locale)
		locales[i].Title = strings.TrimSpace(locales[i].Title)
		locales[i].Subtitle = strings.TrimSpace(locales[i].Subtitle)
		locales[i].Description = strings.TrimSpace(locales[i].Description)
		if len(locales[i].Attributes) == 0 {
			locales[i].Attributes = nil
		}
	}
	return locales
}

// Validate ensures required locales have mandatory fields before submission.
func (s *LocaleService) Validate(defaultLocale string, locales []LocaleContent) ValidationErrors {
	var errs ValidationErrors
	seen := map[string]LocaleContent{}
	for _, entry := range locales {
		key := strings.TrimSpace(strings.ToLower(entry.Locale))
		if key == "" {
			errs = errs.add("locale", "locale code is required")
			continue
		}
		if _, ok := seen[key]; ok {
			errs = errs.add("locale", "duplicate locale: "+entry.Locale)
		}
		if strings.TrimSpace(entry.Title) == "" {
			errs = errs.add(entry.Locale+".title", "title is required")
		}
		if key == strings.ToLower(strings.TrimSpace(defaultLocale)) && strings.TrimSpace(entry.Description) == "" {
			errs = errs.add(entry.Locale+".description", "default locale description is required")
		}
		seen[key] = entry
	}
	for _, needed := range s.required {
		if _, ok := seen[strings.ToLower(needed)]; !ok {
			errs = errs.add("locale", "missing required locale: "+needed)
		}
	}
	return errs
}
