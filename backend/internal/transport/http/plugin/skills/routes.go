package skills

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

type manifest struct {
	SkillID            string              `json:"skill_id" yaml:"id"`
	Provider           string              `json:"provider,omitempty" yaml:"provider"`
	Version            string              `json:"version" yaml:"version"`
	Title              string              `json:"title,omitempty" yaml:"title"`
	Description        string              `json:"description" yaml:"description"`
	IntentExamples     []string            `json:"intent_examples,omitempty" yaml:"intent_examples"`
	ResponseGuidance   map[string][]string `json:"response_guidance,omitempty" yaml:"response_guidance"`
	ActionRequiredArgs map[string][]string `json:"action_required_args,omitempty" yaml:"action_required_args"`
	Capability         string              `json:"capability,omitempty" yaml:"capability"`
	ActionCapabilities map[string]string   `json:"action_capabilities,omitempty" yaml:"action_capabilities"`
	Visibility         string              `json:"visibility,omitempty" yaml:"visibility"`
	Status             string              `json:"status,omitempty" yaml:"status"`
	Executor           map[string]any      `json:"executor" yaml:"executor"`
	InputSchema        map[string]any      `json:"input_schema" yaml:"-"`
	OutputSchema       map[string]any      `json:"output_schema,omitempty" yaml:"-"`

	inputSchemaRef  string
	outputSchemaRef string
}

type schemaResponse struct {
	SkillID      string         `json:"skill_id"`
	Version      string         `json:"version"`
	InputSchema  map[string]any `json:"input_schema"`
	OutputSchema map[string]any `json:"output_schema,omitempty"`
}

type frontmatter struct {
	ID                 string              `yaml:"id"`
	Provider           string              `yaml:"provider"`
	Version            string              `yaml:"version"`
	Title              string              `yaml:"title"`
	Description        string              `yaml:"description"`
	IntentExamples     []string            `yaml:"intent_examples"`
	ResponseGuidance   map[string][]string `yaml:"response_guidance"`
	ActionRequiredArgs map[string][]string `yaml:"action_required_args"`
	Capability         string              `yaml:"capability"`
	ActionCapabilities map[string]string   `yaml:"action_capabilities"`
	Visibility         string              `yaml:"visibility"`
	Status             string              `yaml:"status"`
	Executor           map[string]any      `yaml:"executor"`
	InputSchema        string              `yaml:"input_schema"`
	OutputSchema       string              `yaml:"output_schema"`
}

func RegisterAPIRoutes(rg *gin.RouterGroup, _ *app.Deps) {
	h := handler{}
	group := rg.Group("/plugin/skills")
	group.GET("", h.list)
	group.GET("/", h.list)
	group.GET("/:skill_id/schema", h.schema)
}

type handler struct{}

func (handler) list(c *gin.Context) {
	items, err := loadManifests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (handler) schema(c *gin.Context) {
	skillID := strings.TrimSpace(c.Param("skill_id"))
	version := strings.TrimSpace(c.Query("version"))
	items, err := loadManifests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, item := range items {
		if item.SkillID != skillID {
			continue
		}
		if version != "" && item.Version != version {
			continue
		}
		c.JSON(http.StatusOK, schemaResponse{
			SkillID:      item.SkillID,
			Version:      item.Version,
			InputSchema:  item.InputSchema,
			OutputSchema: item.OutputSchema,
		})
		return
	}
	c.JSON(http.StatusNotFound, gin.H{
		"success":  false,
		"skill_id": skillID,
		"error": gin.H{
			"code":    "skill.not_found",
			"message": "skill is not registered",
		},
	})
}

func loadManifests() ([]manifest, error) {
	root, err := findSkillsRoot()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	items := make([]manifest, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		item, err := loadManifest(filepath.Join(root, entry.Name()))
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].SkillID == items[j].SkillID {
			return items[i].Version < items[j].Version
		}
		return items[i].SkillID < items[j].SkillID
	})
	return items, nil
}

func loadManifest(packagePath string) (manifest, error) {
	raw, err := os.ReadFile(filepath.Join(packagePath, "SKILL.md"))
	if err != nil {
		return manifest{}, err
	}
	front, err := splitFrontmatter(string(raw))
	if err != nil {
		return manifest{}, err
	}
	var fm frontmatter
	if err := yaml.Unmarshal([]byte(front), &fm); err != nil {
		return manifest{}, err
	}
	item := manifest{
		SkillID:            strings.TrimSpace(fm.ID),
		Provider:           strings.TrimSpace(fm.Provider),
		Version:            firstNonEmpty(fm.Version, "1.0.0"),
		Title:              strings.TrimSpace(fm.Title),
		Description:        strings.TrimSpace(fm.Description),
		IntentExamples:     fm.IntentExamples,
		ResponseGuidance:   fm.ResponseGuidance,
		ActionRequiredArgs: fm.ActionRequiredArgs,
		Capability:         strings.TrimSpace(fm.Capability),
		ActionCapabilities: fm.ActionCapabilities,
		Visibility:         firstNonEmpty(fm.Visibility, "tenant"),
		Status:             firstNonEmpty(fm.Status, "active"),
		Executor:           fm.Executor,
		inputSchemaRef:     strings.TrimSpace(fm.InputSchema),
		outputSchemaRef:    strings.TrimSpace(fm.OutputSchema),
	}
	if item.InputSchema, err = readSchema(packagePath, item.inputSchemaRef); err != nil {
		return manifest{}, err
	}
	if item.OutputSchema, err = readSchema(packagePath, item.outputSchemaRef); err != nil {
		return manifest{}, err
	}
	return item, nil
}

func findSkillsRoot() (string, error) {
	if value := strings.TrimSpace(os.Getenv("PLUGIN_SKILLS_DIR")); value != "" {
		return filepath.Abs(value)
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, "skills")
		if st, err := os.Stat(candidate); err == nil && st.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", os.ErrNotExist
}

func splitFrontmatter(raw string) (string, error) {
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return "", os.ErrInvalid
	}
	rest := strings.TrimPrefix(normalized, "---\n")
	idx := strings.Index(rest, "\n---\n")
	if idx < 0 {
		return "", os.ErrInvalid
	}
	return rest[:idx], nil
}

func readSchema(packagePath, ref string) (map[string]any, error) {
	if strings.TrimSpace(ref) == "" {
		return nil, nil
	}
	clean := filepath.Clean(ref)
	if filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." {
		return nil, os.ErrPermission
	}
	raw, err := os.ReadFile(filepath.Join(packagePath, clean))
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
