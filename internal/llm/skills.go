package llm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"go.yaml.in/yaml/v3"
)

const (
	skillToolLoad      = "load_skill"
	skillToolRead      = "read_skill_file"
	skillFileMaxBytes  = 48 << 10
	skillRoundsDefault = 3
	skillMemoryMax     = 256
	skillKnownMaxBytes = 32 << 10
)

// SkillLibrary is a folder of skills, one folder per skill with a SKILL.md
// whose front matter has a name and description. Only Markdown files are
// served.
type SkillLibrary struct {
	dir    string
	skills []skillInfo
	byName map[string]skillInfo
}

type skillInfo struct {
	Name        string
	Description string
	dir         string
}

// LoadSkillLibrary reads every dir/*/SKILL.md.
func LoadSkillLibrary(dir string) (*SkillLibrary, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if abs, err = filepath.EvalSymlinks(abs); err != nil {
		return nil, fmt.Errorf("skills folder %s: %w", dir, err)
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, fmt.Errorf("skills folder %s: %w", dir, err)
	}
	lib := &SkillLibrary{dir: abs, byName: map[string]skillInfo{}}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(abs, e.Name(), "SKILL.md"))
		if err != nil {
			continue
		}
		front, _ := splitFrontMatter(data)
		var meta struct {
			Name        string `yaml:"name"`
			Description string `yaml:"description"`
		}
		_ = yaml.Unmarshal(front, &meta)
		if meta.Name == "" {
			meta.Name = e.Name()
		}
		info := skillInfo{Name: meta.Name, Description: strings.Join(strings.Fields(meta.Description), " "), dir: filepath.Join(abs, e.Name())}
		lib.skills = append(lib.skills, info)
		lib.byName[info.Name] = info
	}
	if len(lib.skills) == 0 {
		return nil, fmt.Errorf("skills folder %s has no */SKILL.md", dir)
	}
	sort.Slice(lib.skills, func(i, j int) bool { return lib.skills[i].Name < lib.skills[j].Name })
	return lib, nil
}

func splitFrontMatter(data []byte) (front, body []byte) {
	if !bytes.HasPrefix(data, []byte("---\n")) {
		return nil, data
	}
	end := bytes.Index(data[4:], []byte("\n---"))
	if end < 0 {
		return nil, data
	}
	front = data[4 : 4+end]
	body = bytes.TrimLeft(data[4+end+4:], "\r\n")
	return front, body
}

// load returns a skill's SKILL.md without its front matter.
func (l *SkillLibrary) load(name string) (string, error) {
	info, ok := l.byName[name]
	if !ok {
		return "", fmt.Errorf("no skill named %q", name)
	}
	data, err := os.ReadFile(filepath.Join(info.dir, "SKILL.md"))
	if err != nil {
		return "", err
	}
	_, body := splitFrontMatter(data)
	return fmt.Sprintf("Skill %s (file SKILL.md). Read the files it links to with %s, using paths relative to the skill folder.\n\n%s",
		name, skillToolRead, body), nil
}

// readFile returns a Markdown file inside a skill's folder.
func (l *SkillLibrary) readFile(name, rel string) (string, error) {
	info, ok := l.byName[name]
	if !ok {
		return "", fmt.Errorf("no skill named %q", name)
	}
	clean := path.Clean("/" + strings.ReplaceAll(strings.TrimSpace(rel), "\\", "/"))[1:]
	if clean == "" || !strings.HasSuffix(clean, ".md") {
		return "", errors.New("path must name a .md file inside the skill folder")
	}
	full, err := filepath.EvalSymlinks(filepath.Join(info.dir, filepath.FromSlash(clean)))
	if err != nil {
		return "", fmt.Errorf("no file %s in skill %s", clean, name)
	}
	if r, err := filepath.Rel(info.dir, full); err != nil || r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) {
		return "", errors.New("path must stay inside the skill folder")
	}
	data, err := os.ReadFile(full)
	if err != nil {
		return "", err
	}
	note := ""
	if len(data) > skillFileMaxBytes {
		data, note = data[:skillFileMaxBytes], "\n\n[file cut at 48 KB]"
	}
	_, body := splitFrontMatter(data)
	dir := path.Dir(clean)
	return fmt.Sprintf("Skill %s, file %s. Links in this file are relative to %s/.\n\n%s%s", name, clean, dir, body, note), nil
}

func (l *SkillLibrary) names() []any {
	out := make([]any, len(l.skills))
	for i, s := range l.skills {
		out[i] = s.Name
	}
	return out
}

func (l *SkillLibrary) tools() []Tool {
	return []Tool{
		{
			Name:        skillToolLoad,
			Description: "Load one of your game-mastering or fiction skills. Returns its instructions as text; changes nothing in the game.",
			Parameters: map[string]any{
				"type":                 "object",
				"properties":           map[string]any{"name": map[string]any{"type": "string", "enum": l.names()}},
				"required":             []any{"name"},
				"additionalProperties": false,
			},
		},
		{
			Name:        skillToolRead,
			Description: "Read a Markdown file that a loaded skill links to, such as references/base-workflow.md. Returns text; changes nothing in the game.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"skill": map[string]any{"type": "string", "enum": l.names()},
					"path":  map[string]any{"type": "string", "description": "Path relative to the skill folder."},
				},
				"required":             []any{"skill", "path"},
				"additionalProperties": false,
			},
		},
	}
}

func (l *SkillLibrary) instructions() string {
	var b strings.Builder
	b.WriteString("You have skills for game mastering and writing fiction:\n")
	for _, s := range l.skills {
		fmt.Fprintf(&b, "- %s: %s\n", s.Name, s.Description)
	}
	b.WriteString("\nStart this turn by reading a skill, before you narrate or call any game tool:\n" +
		"1. Reply with only a " + skillToolLoad + " call for the skill that fits the turn best (for running play, run-rpg-sessions).\n" +
		"2. Reply with only a " + skillToolRead + " call for the procedure the skill points to.\n" +
		"3. Then write the turn as the skill describes, calling game tools as usual.\n" +
		"Skill tools only return text and change nothing in the game. A game tool called in the same reply as a skill tool is not run.")
	return b.String()
}

// call answers one skill tool call.
func (l *SkillLibrary) call(tc ToolCall) string {
	str := func(k string) string { s, _ := tc.Arguments[k].(string); return s }
	var out string
	var err error
	switch tc.Name {
	case skillToolLoad:
		out, err = l.load(str("name"))
	case skillToolRead:
		out, err = l.readFile(str("skill"), str("path"))
	}
	if err != nil {
		return "error: " + err.Error()
	}
	return out
}

// SkillProvider gives the wrapped provider two read-only tools, load_skill
// and read_skill_file, and answers them itself: the engine only sees game
// tool calls. The skill exchange of a turn is remembered by the first game
// tool call it led to and put back before that call when the engine sends the
// tool results, so the follow-up request has the same context and does not
// reload the skills.
type SkillProvider struct {
	inner     Provider
	lib       *SkillLibrary
	maxRounds int

	mu    sync.Mutex
	turns map[string][]Message
	order []string
	// known holds skill text already read, in the order it was read; later
	// turns get it in the instructions instead of reading it again.
	known      []knownSkillText
	knownBytes int
}

type knownSkillText struct {
	key, text string
}

// NewSkillProvider wraps inner with the skill tools from lib.
func NewSkillProvider(inner Provider, lib *SkillLibrary) *SkillProvider {
	return &SkillProvider{inner: inner, lib: lib, maxRounds: skillRoundsDefault, turns: map[string][]Message{}}
}

func isSkillTool(name string) bool { return name == skillToolLoad || name == skillToolRead }

// Complete implements Provider.
func (p *SkillProvider) Complete(ctx context.Context, messages []Message, tools []Tool) (*Response, error) {
	base := p.withContext(messages)
	withSkills := append(append([]Tool{}, tools...), p.lib.tools()...)

	var exchange []Message
	for round := 0; ; round++ {
		roundTools := withSkills
		if round == p.maxRounds {
			roundTools = tools // last round: game tools only
		}
		resp, err := p.inner.Complete(ctx, append(append([]Message{}, base...), exchange...), roundTools)
		if err != nil {
			return nil, err
		}
		var skillCalls int
		for _, tc := range resp.ToolCalls {
			if isSkillTool(tc.Name) {
				skillCalls++
			}
		}
		if skillCalls == 0 {
			if len(exchange) > 0 && len(resp.ToolCalls) > 0 {
				p.remember(resp.ToolCalls[0].ID, exchange)
			}
			return resp, nil
		}

		exchange = append(exchange, Message{Role: RoleAssistant, Content: resp.Content, ToolCalls: resp.ToolCalls})
		for _, tc := range resp.ToolCalls {
			result := `{"error":"not run: a game tool called together with a skill tool does not run. Call it again in your next reply."}`
			if isSkillTool(tc.Name) {
				result = p.lib.call(tc)
				p.learn(tc, result)
			}
			exchange = append(exchange, Message{Role: RoleTool, ToolCallID: tc.ID, Content: result})
		}
	}
}

// Stream implements Provider by completing first; skill rounds can't stream.
func (p *SkillProvider) Stream(ctx context.Context, messages []Message, tools []Tool) (<-chan StreamChunk, error) {
	resp, err := p.Complete(ctx, messages, tools)
	if err != nil {
		return nil, err
	}
	ch := make(chan StreamChunk, 2+len(resp.ToolCalls))
	if resp.Content != "" {
		ch <- StreamChunk{ContentDelta: resp.Content}
	}
	for i := range resp.ToolCalls {
		ch <- StreamChunk{ToolCallDelta: &resp.ToolCalls[i]}
	}
	ch <- StreamChunk{Done: true}
	close(ch)
	return ch, nil
}

// withContext adds the skill instructions, and any skill text read in earlier
// turns, after the leading system messages, and puts remembered skill
// exchanges back before the calls they led to.
func (p *SkillProvider) withContext(messages []Message) []Message {
	out := make([]Message, 0, len(messages)+1)
	i := 0
	for i < len(messages) && messages[i].Role == RoleSystem {
		out = append(out, messages[i])
		i++
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	instructions := p.lib.instructions()
	if len(p.known) > 0 {
		var b strings.Builder
		b.WriteString(instructions)
		b.WriteString("\n\nSkill text you have already read is below. When it covers this turn, skip steps 1 and 2 and use it.\n")
		for _, k := range p.known {
			b.WriteString("\n<skill_text>\n" + k.text + "\n</skill_text>\n")
		}
		instructions = b.String()
	}
	out = append(out, Message{Role: RoleSystem, Content: instructions})
	for ; i < len(messages); i++ {
		m := messages[i]
		if m.Role == RoleAssistant && len(m.ToolCalls) > 0 {
			out = append(out, p.turns[m.ToolCalls[0].ID]...)
		}
		out = append(out, m)
	}
	return out
}

// learn keeps a successful skill read for later turns, up to 32 KB in all.
func (p *SkillProvider) learn(tc ToolCall, result string) {
	if strings.HasPrefix(result, "error: ") {
		return
	}
	key := fmt.Sprintf("%s %v %v %v", tc.Name, tc.Arguments["name"], tc.Arguments["skill"], tc.Arguments["path"])
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, k := range p.known {
		if k.key == key {
			return
		}
	}
	if p.knownBytes+len(result) > skillKnownMaxBytes {
		return
	}
	p.known = append(p.known, knownSkillText{key: key, text: result})
	p.knownBytes += len(result)
}

func (p *SkillProvider) remember(firstCallID string, exchange []Message) {
	if firstCallID == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.turns[firstCallID]; !ok {
		p.order = append(p.order, firstCallID)
	}
	p.turns[firstCallID] = exchange
	for len(p.order) > skillMemoryMax {
		delete(p.turns, p.order[0])
		p.order = p.order[1:]
	}
}
