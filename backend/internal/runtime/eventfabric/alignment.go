package eventfabric

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type TopicDecl struct {
	Topic       string
	Description string
	Actions     []string
}

type manifestDoc struct {
	Events struct {
		Topics []struct {
			Key         string   `yaml:"key"`
			Description string   `yaml:"description"`
			Actions     []string `yaml:"actions"`
		} `yaml:"topics"`
	} `yaml:"events"`
}

type executionDoc struct {
	Topics []struct {
		Topic       string `yaml:"topic"`
		Description string `yaml:"description"`
		ACL         []struct {
			Actions []string `yaml:"actions"`
		} `yaml:"acl"`
	} `yaml:"topics"`
}

func ResolvePaths() (string, string) {
	return resolveFirstExisting(
			"plugin.yaml",
			filepath.Join("..", "plugin.yaml"),
			filepath.Join("..", "..", "plugin.yaml"),
		),
		resolveFirstExisting(
			filepath.Join("config", "event_fabric.yaml"),
			filepath.Join("..", "config", "event_fabric.yaml"),
			filepath.Join("..", "..", "config", "event_fabric.yaml"),
		)
}

func LoadManifestTopics(path string) ([]TopicDecl, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc manifestDoc
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	out := make([]TopicDecl, 0, len(doc.Events.Topics))
	for _, item := range doc.Events.Topics {
		topic := strings.TrimSpace(item.Key)
		if topic == "" {
			continue
		}
		out = append(out, TopicDecl{
			Topic:       topic,
			Description: strings.TrimSpace(item.Description),
			Actions:     normalizeActions(item.Actions),
		})
	}
	return out, nil
}

func LoadExecutionTopics(path string) ([]TopicDecl, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc executionDoc
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	out := make([]TopicDecl, 0, len(doc.Topics))
	for _, item := range doc.Topics {
		topic := strings.TrimSpace(item.Topic)
		if topic == "" {
			continue
		}
		actions := make([]string, 0, 4)
		for _, acl := range item.ACL {
			actions = append(actions, acl.Actions...)
		}
		out = append(out, TopicDecl{
			Topic:       topic,
			Description: strings.TrimSpace(item.Description),
			Actions:     normalizeActions(actions),
		})
	}
	return out, nil
}

func ValidateConsistency(manifestTopics, executionTopics []TopicDecl) error {
	manifestMap := indexTopics(manifestTopics)
	executionMap := indexTopics(executionTopics)

	var issues []string
	for topic, manifestDecl := range manifestMap {
		execDecl, ok := executionMap[topic]
		if !ok {
			issues = append(issues, fmt.Sprintf("missing in execution config: %s", topic))
			continue
		}
		if !sameActionSet(manifestDecl.Actions, execDecl.Actions) {
			issues = append(issues, fmt.Sprintf("actions mismatch for %s: manifest=%v execution=%v", topic, manifestDecl.Actions, execDecl.Actions))
		}
	}
	for topic := range executionMap {
		if _, ok := manifestMap[topic]; !ok {
			issues = append(issues, fmt.Sprintf("missing in plugin manifest: %s", topic))
		}
	}
	if len(issues) == 0 {
		return nil
	}
	sort.Strings(issues)
	return fmt.Errorf("%s", strings.Join(issues, "; "))
}

func resolveFirstExisting(candidates ...string) string {
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) == "" {
			continue
		}
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return ""
}

func indexTopics(items []TopicDecl) map[string]TopicDecl {
	out := map[string]TopicDecl{}
	for _, item := range items {
		topic := strings.TrimSpace(item.Topic)
		if topic == "" {
			continue
		}
		out[topic] = TopicDecl{
			Topic:       topic,
			Description: strings.TrimSpace(item.Description),
			Actions:     normalizeActions(item.Actions),
		}
	}
	return out
}

func sameActionSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	am := map[string]struct{}{}
	for _, v := range a {
		am[v] = struct{}{}
	}
	for _, v := range b {
		if _, ok := am[v]; !ok {
			return false
		}
	}
	return true
}

func normalizeActions(actions []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(actions))
	for _, action := range actions {
		v := strings.ToLower(strings.TrimSpace(action))
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}
