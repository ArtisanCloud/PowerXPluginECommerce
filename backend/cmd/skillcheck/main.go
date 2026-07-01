package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type skillFrontmatter struct {
	ID                 string            `yaml:"id"`
	Provider           string            `yaml:"provider"`
	Version            string            `yaml:"version"`
	Title              string            `yaml:"title"`
	Description        string            `yaml:"description"`
	IntentExamples     []string          `yaml:"intent_examples"`
	Executor           skillExecutor     `yaml:"executor"`
	ActionCapabilities map[string]string `yaml:"action_capabilities"`
	InputSchema        string            `yaml:"input_schema"`
	OutputSchema       string            `yaml:"output_schema"`
}

type skillExecutor struct {
	Type              string            `yaml:"type"`
	Capability        string            `yaml:"capability"`
	PrepareCapability string            `yaml:"prepare_capability"`
	ActionMap         map[string]string `yaml:"action_map"`
}

type capabilityCatalog struct {
	Capabilities struct {
		Provides []struct {
			ID string `yaml:"id"`
		} `yaml:"provides"`
	} `yaml:"capabilities"`
}

func main() {
	root := flag.String("root", ".", "plugin package root")
	flag.Parse()

	if err := run(*root); err != nil {
		fmt.Fprintf(os.Stderr, "skillcheck: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("skillcheck: ok")
}

func run(root string) error {
	root = filepath.Clean(root)
	skillsDir := filepath.Join(root, "skills")
	if st, err := os.Stat(skillsDir); err != nil || !st.IsDir() {
		return fmt.Errorf("missing skills directory: %s", skillsDir)
	}
	catalogIDs, err := loadCapabilityIDs(filepath.Join(root, "plugin.d", "capabilities.yaml"))
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return err
	}
	var packages []string
	for _, entry := range entries {
		if entry.IsDir() {
			packages = append(packages, entry.Name())
		}
	}
	sort.Strings(packages)
	if len(packages) == 0 {
		return errors.New("skills directory has no skill packages")
	}
	for _, name := range packages {
		if err := validatePackage(filepath.Join(skillsDir, name), catalogIDs); err != nil {
			return fmt.Errorf("skills/%s: %w", name, err)
		}
	}
	return nil
}

func validatePackage(packagePath string, catalogIDs map[string]struct{}) error {
	raw, err := os.ReadFile(filepath.Join(packagePath, "SKILL.md"))
	if err != nil {
		return err
	}
	front, body, err := splitFrontmatter(string(raw))
	if err != nil {
		return err
	}
	if strings.TrimSpace(body) == "" {
		return errors.New("Markdown body is required")
	}
	var fm skillFrontmatter
	if err := yaml.Unmarshal([]byte(front), &fm); err != nil {
		return err
	}
	required := map[string]string{
		"id":                          fm.ID,
		"provider":                    fm.Provider,
		"version":                     fm.Version,
		"title":                       fm.Title,
		"description":                 fm.Description,
		"executor.type":               fm.Executor.Type,
		"executor.capability":         fm.Executor.Capability,
		"executor.prepare_capability": fm.Executor.PrepareCapability,
		"input_schema":                fm.InputSchema,
		"output_schema":               fm.OutputSchema,
	}
	for field, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", field)
		}
	}
	if len(fm.IntentExamples) == 0 {
		return errors.New("intent_examples is required")
	}
	if strings.TrimSpace(fm.Executor.Type) != "capability" {
		return errors.New("executor.type must be capability")
	}
	if len(fm.Executor.ActionMap) == 0 {
		return errors.New("executor.action_map is required")
	}
	if _, ok := catalogIDs[fm.Executor.PrepareCapability]; !ok {
		return fmt.Errorf("executor.prepare_capability %q is not provided by plugin.d/capabilities.yaml", fm.Executor.PrepareCapability)
	}
	for action, capabilityID := range fm.Executor.ActionMap {
		if strings.TrimSpace(action) == "" || strings.TrimSpace(capabilityID) == "" {
			return errors.New("executor.action_map contains empty action or capability")
		}
		if _, ok := catalogIDs[capabilityID]; !ok {
			return fmt.Errorf("executor.action_map.%s %q is not provided by plugin.d/capabilities.yaml", action, capabilityID)
		}
	}
	for action, capabilityID := range fm.ActionCapabilities {
		if _, ok := fm.Executor.ActionMap[action]; !ok {
			return fmt.Errorf("action_capabilities.%s has no matching executor.action_map entry", action)
		}
		if _, ok := catalogIDs[capabilityID]; !ok {
			return fmt.Errorf("action_capabilities.%s %q is not provided by plugin.d/capabilities.yaml", action, capabilityID)
		}
	}
	if err := validateSchema(filepath.Join(packagePath, fm.InputSchema)); err != nil {
		return fmt.Errorf("input_schema: %w", err)
	}
	if err := validateSchema(filepath.Join(packagePath, fm.OutputSchema)); err != nil {
		return fmt.Errorf("output_schema: %w", err)
	}
	return nil
}

func loadCapabilityIDs(path string) (map[string]struct{}, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var catalog capabilityCatalog
	if err := yaml.Unmarshal(raw, &catalog); err != nil {
		return nil, err
	}
	out := map[string]struct{}{}
	for _, item := range catalog.Capabilities.Provides {
		id := strings.TrimSpace(item.ID)
		if id != "" {
			out[id] = struct{}{}
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s has no capabilities.provides entries", path)
	}
	return out, nil
}

func splitFrontmatter(raw string) (string, string, error) {
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return "", "", errors.New("SKILL.md frontmatter is required")
	}
	rest := strings.TrimPrefix(normalized, "---\n")
	idx := strings.Index(rest, "\n---\n")
	if idx < 0 {
		return "", "", errors.New("SKILL.md frontmatter closing marker is required")
	}
	return rest[:idx], rest[idx+len("\n---\n"):], nil
}

func validateSchema(path string) error {
	clean := filepath.Clean(path)
	raw, err := os.ReadFile(clean)
	if err != nil {
		return err
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return err
	}
	if strings.TrimSpace(fmt.Sprint(doc["type"])) == "" {
		return errors.New("schema type is required")
	}
	return nil
}
