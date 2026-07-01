package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	httpregistry "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http"
	"gopkg.in/yaml.v3"
)

const (
	pluginID     = "com.powerx.plugins.ecommerce"
	httpBaseExpr = "${POWERX_PLUGIN_HTTP_BASE:-/api/v1}"
)

type routeSpec struct {
	Method string
	Path   string
}

type capabilitySpec struct {
	ID       string
	Resource string
	Action   string
	Routes   []routeSpec
}

func main() {
	root := flag.String("root", ".", "plugin root")
	prefix := flag.String("prefix", "/api/v1", "API prefix")
	flag.Parse()

	if err := run(*root, *prefix); err != nil {
		fmt.Fprintf(os.Stderr, "capability_catalog: %v\n", err)
		os.Exit(1)
	}
}

func run(root, prefix string) error {
	root = filepath.Clean(root)
	specs := collectSpecs(prefix, httpregistry.StaticRBACEntries(prefix))
	if len(specs) == 0 {
		return fmt.Errorf("no RBAC-backed routes found")
	}
	if err := os.MkdirAll(filepath.Join(root, "contracts", "capabilities"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(root, "contracts", "schema", "input"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(root, "contracts", "schema", "output"), 0o755); err != nil {
		return err
	}
	if err := removeGeneratedCapabilities(root); err != nil {
		return err
	}
	for _, spec := range specs {
		if isTemplateCapability(spec.ID) {
			continue
		}
		if err := writeCapability(root, spec); err != nil {
			return err
		}
	}
	fmt.Printf("capability_catalog: generated %d RBAC-backed capability descriptors\n", len(specs))
	return nil
}

func collectSpecs(prefix string, entries map[string]authx.Permission) []capabilitySpec {
	prefix = strings.TrimRight(strings.TrimSpace(prefix), "/")
	if prefix == "" {
		prefix = "/api/v1"
	}
	byID := map[string]*capabilitySpec{}
	for key, perm := range entries {
		method, pathValue, ok := splitRouteKey(key)
		if !ok {
			continue
		}
		resource := strings.TrimSpace(perm.Resource)
		action := strings.TrimSpace(perm.Action)
		if resource == "" || action == "" {
			continue
		}
		id := capabilityID(resource, action)
		if id == "" {
			continue
		}
		item := byID[id]
		if item == nil {
			item = &capabilitySpec{ID: id, Resource: resource, Action: action}
			byID[id] = item
		}
		item.Routes = append(item.Routes, routeSpec{
			Method: strings.ToUpper(strings.TrimSpace(method)),
			Path:   trimRoutePrefix(prefix, strings.TrimSpace(pathValue)),
		})
	}
	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]capabilitySpec, 0, len(ids))
	for _, id := range ids {
		item := byID[id]
		sort.Slice(item.Routes, func(i, j int) bool {
			if item.Routes[i].Path == item.Routes[j].Path {
				return item.Routes[i].Method < item.Routes[j].Method
			}
			return item.Routes[i].Path < item.Routes[j].Path
		})
		out = append(out, *item)
	}
	return out
}

func writeCapability(root string, spec capabilitySpec) error {
	inputRel := "schema/input/" + spec.ID + ".json"
	outputRel := "schema/output/" + spec.ID + ".json"
	doc := map[string]any{
		"id":          spec.ID,
		"type":        "API",
		"version":     "1.0.0",
		"status":      "active",
		"description": fmt.Sprintf("%s %s capability", spec.Resource, spec.Action),
		"rbac": map[string]any{
			"resource": spec.Resource,
			"actions":  []string{spec.Action},
		},
		"consumes": []map[string]any{{
			"id":   inputRel,
			"path": inputRel,
			"kind": "input",
		}},
		"provides": []map[string]any{{
			"id":   outputRel,
			"path": outputRel,
			"kind": "output",
		}},
		"metadata": map[string]any{
			"protocols": map[string]any{
				"rest": restEntries(spec.Routes),
			},
			"source": map[string]any{
				"generated_from": "route_rbac",
				"resource":       spec.Resource,
				"action":         spec.Action,
			},
		},
	}
	raw, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(root, "contracts", "capabilities", spec.ID+".yaml"), raw, 0o644); err != nil {
		return err
	}
	if err := writeSchema(filepath.Join(root, "contracts", inputRel), spec.ID+"Input"); err != nil {
		return err
	}
	return writeSchema(filepath.Join(root, "contracts", outputRel), spec.ID+"Output")
}

func removeGeneratedCapabilities(root string) error {
	dir := filepath.Join(root, "contracts", "capabilities")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var doc struct {
			ID       string `yaml:"id"`
			Metadata struct {
				Source struct {
					GeneratedFrom string `yaml:"generated_from"`
				} `yaml:"source"`
			} `yaml:"metadata"`
		}
		if err := yaml.Unmarshal(raw, &doc); err != nil {
			return fmt.Errorf("read generated marker from %s: %w", path, err)
		}
		if doc.Metadata.Source.GeneratedFrom != "route_rbac" {
			continue
		}
		id := strings.TrimSpace(doc.ID)
		if id == "" {
			id = strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		for _, schemaDir := range []string{"input", "output"} {
			schemaPath := filepath.Join(root, "contracts", "schema", schemaDir, id+".json")
			if err := os.Remove(schemaPath); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	return nil
}

func restEntries(routes []routeSpec) []map[string]any {
	out := make([]map[string]any, 0, len(routes))
	for _, route := range routes {
		out = append(out, map[string]any{
			"method": route.Method,
			"path":   httpBaseExpr + route.Path,
		})
	}
	return out
}

func writeSchema(path, title string) error {
	doc := map[string]any{
		"$schema":              "https://json-schema.org/draft/2020-12/schema",
		"title":                title,
		"type":                 "object",
		"additionalProperties": true,
	}
	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(path, raw, 0o644)
}

func capabilityID(resource, action string) string {
	resource = strings.TrimSpace(resource)
	action = strings.TrimSpace(action)
	if resource == "" || action == "" {
		return ""
	}
	const prefix = pluginID + ":"
	if strings.HasPrefix(resource, prefix) {
		return pluginID + "." + sanitizeID(strings.TrimPrefix(resource, prefix)) + "." + sanitizeID(action)
	}
	return pluginID + "." + sanitizeID(resource) + "." + sanitizeID(action)
}

func sanitizeID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(":", ".", "/", ".", "_", ".", "-", ".")
	value = replacer.Replace(value)
	parts := strings.FieldsFunc(value, func(r rune) bool { return r == '.' || r == ' ' || r == '\t' || r == '\n' })
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.Trim(part, ".")
		if part != "" {
			cleaned = append(cleaned, part)
		}
	}
	return strings.Join(cleaned, ".")
}

func isTemplateCapability(id string) bool {
	return id == pluginID+".template.manage" || id == pluginID+".template.read"
}

func splitRouteKey(route string) (string, string, bool) {
	route = strings.TrimSpace(route)
	idx := strings.Index(route, ":/")
	if idx < 0 {
		return "", "", false
	}
	return route[:idx], route[idx+1:], true
}

func trimRoutePrefix(prefix, pathValue string) string {
	if prefix == "" || prefix == "/" {
		return pathValue
	}
	if pathValue == prefix {
		return "/"
	}
	if strings.HasPrefix(pathValue, prefix+"/") {
		return strings.TrimPrefix(pathValue, prefix)
	}
	return pathValue
}
