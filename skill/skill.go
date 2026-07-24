package skill

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Skill 表示一个 Agent Skill
type Skill struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Body        string            `yaml:"-"` // markdown 正文（去掉 frontmatter）
	Location    string            `yaml:"-"` // SKILL.md 的绝对路径
	BaseDir     string            `yaml:"-"` // Skill 所在目录
	Metadata    map[string]string `yaml:"metadata,omitempty"`
}

// Manager 管理所有已发现的 Skill
type Manager struct {
	skills map[string]*Skill
}

// NewManager 扫描指定目录下所有子目录中的 SKILL.md，返回 Manager
func NewManager(skillsDir string) (*Manager, error) {
	m := &Manager{skills: make(map[string]*Skill)}

	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		log.Printf("[skill] skills directory not found: %s, skipping", skillsDir)
		return m, nil
	}

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("read skills dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillMD := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
		if _, err := os.Stat(skillMD); os.IsNotExist(err) {
			continue
		}

		s, err := ParseSkillMD(skillMD)
		if err != nil {
			log.Printf("[skill] failed to parse %s: %v", skillMD, err)
			continue
		}
		m.skills[s.Name] = s
		log.Printf("[skill] loaded: %s", s.Name)
	}

	log.Printf("[skill] total skills loaded: %d", len(m.skills))
	return m, nil
}

// frontmatter 用于解析 YAML frontmatter
type frontmatter struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Metadata    map[string]string `yaml:"metadata,omitempty"`
}

// ParseSkillMD 解析 SKILL.md 文件，提取 frontmatter 和 body
func ParseSkillMD(path string) (*Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	content := string(data)

	// 查找 frontmatter 的起止位置
	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("missing YAML frontmatter in %s", path)
	}

	endIdx := strings.Index(content[3:], "---")
	if endIdx == -1 {
		return nil, fmt.Errorf("unterminated YAML frontmatter in %s", path)
	}

	yamlBlock := content[3 : endIdx+3]
	body := strings.TrimSpace(content[endIdx+3+3:]) // skip closing "---\n"

	var fm frontmatter
	if err := yaml.Unmarshal([]byte(yamlBlock), &fm); err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}

	if fm.Name == "" || fm.Description == "" {
		return nil, fmt.Errorf("name and description are required in %s", path)
	}

	absPath, _ := filepath.Abs(path)
	baseDir := filepath.Dir(absPath)

	return &Skill{
		Name:        fm.Name,
		Description: fm.Description,
		Body:        body,
		Location:    absPath,
		BaseDir:     baseDir,
		Metadata:    fm.Metadata,
	}, nil
}

// Catalog 返回所有 Skill 的 name+description 摘要，用于注入系统提示词
func (m *Manager) Catalog() string {
	if len(m.skills) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## 可用 Skills\n\n")
	sb.WriteString("以下是可用的专业技能。当任务匹配某个 Skill 的描述时，调用 activate_skill 工具加载完整指令。\n\n")
	sb.WriteString("<available_skills>\n")
	for _, s := range m.skills {
		sb.WriteString("  <skill>\n")
		sb.WriteString(fmt.Sprintf("    <name>%s</name>\n", s.Name))
		sb.WriteString(fmt.Sprintf("    <description>%s</description>\n", s.Description))
		sb.WriteString("  </skill>\n")
	}
	sb.WriteString("</available_skills>\n")
	return sb.String()
}

// GetSkill 按名称获取完整 Skill
func (m *Manager) GetSkill(name string) (*Skill, error) {
	s, ok := m.skills[name]
	if !ok {
		return nil, fmt.Errorf("skill not found: %s", name)
	}
	return s, nil
}

// ListNames 返回所有 Skill 名称
func (m *Manager) ListNames() []string {
	names := make([]string, 0, len(m.skills))
	for name := range m.skills {
		names = append(names, name)
	}
	return names
}

// HasSkills 是否有可用的 Skill
func (m *Manager) HasSkills() bool {
	return len(m.skills) > 0
}

// ListResources 列出 Skill 目录下的 scripts/references/assets 文件
func (m *Manager) ListResources(name string) []string {
	s, ok := m.skills[name]
	if !ok {
		return nil
	}

	var resources []string
	for _, dir := range []string{"scripts", "references", "assets"} {
		dirPath := filepath.Join(s.BaseDir, dir)
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				resources = append(resources, filepath.Join(dir, e.Name()))
			}
		}
	}
	return resources
}
