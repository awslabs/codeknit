// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package tui provides an interactive terminal UI for configuring codeknit.
package tui

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"codeknit/internal/config"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/common-nighthawk/go-figure"
)

// screen identifies which TUI screen is active.
type screen int

const (
	screenCommandSelect screen = iota
	screenParseOptions
	screenGraphOptions
	screenGraphAnalyzeOptions
	screenGraphHotspotOptions
	screenFingerprintOptions
)

// command represents a selectable CLI command in the TUI.
type command struct {
	Name string
	Desc string
}

var commands = []command{
	{Name: "parse", Desc: "Parse source code and extract structural information"},
	{Name: "graph show", Desc: "Generate an interactive HTML graph of the codebase"},
	{Name: "graph analyze", Desc: "Run graph analysis algorithms for code quality"},
	{Name: "graph hotspots", Desc: "Rank change hotspots using Git history and structure"},
	{Name: "fingerprint", Desc: "Detect duplicate and near-duplicate code via fuzzy hashing"},
}

// field identifies which input field is currently focused.
type field int

const (
	fieldInputPath field = iota
	fieldOutputMode
	fieldOutputFormat
	fieldOutputDir
	fieldMaxLines
	fieldCollectTest
	fieldMinify
	fieldEdges
	fieldClean
	fieldWorkers
	fieldConfirm
	// Graph analyze specific fields.
	fieldFanThreshold
	fieldGodThreshold
	fieldMaxInheritanceDepth
	fieldTopN
	fieldBetweennessThreshold
	fieldPropagationCutoff
	// Graph hotspot specific fields.
	fieldHotspotFormat
	fieldHotspotSince
	fieldHotspotMaxCommits
	fieldHotspotMaxFilesPerCommit
	fieldHotspotMinCoChanges
	fieldHotspotTopN
	fieldHotspotIncludeMerges
	// Fingerprint specific fields.
	fieldFingerprintMinSim
	fieldFingerprintMaxSim
	fieldFingerprintRerank
	fieldFingerprintModel
	fieldFingerprintShowAll
)

const totalFields = 11

var (
	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A78BFA"))

	activeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C4B5FD")).
			Bold(true)

	quitStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	confirmStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true)

	suggestionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Italic(true)

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))
)

// bannerGradient defines colors from bright (top) to deep (bottom) for depth.
var bannerGradient = []string{
	"#E0CFFC", // highlight / light edge
	"#C4B5FD", // bright
	"#A78BFA", // mid-bright
	"#7C3AED", // core purple
	"#6D28D9", // deeper
	"#5B21B6", // shadow
}

func banner() string {
	fig := figure.NewFigure("codeknit", "shadow", true)
	lines := strings.Split(strings.TrimRight(fig.String(), "\n"), "\n")
	var b strings.Builder
	for i, line := range lines {
		ci := i
		if ci >= len(bannerGradient) {
			ci = len(bannerGradient) - 1
		}
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(bannerGradient[ci])).
			Bold(true)
		b.WriteString(style.Render(line))
		b.WriteByte('\n')
	}
	return b.String()
}

// Model is the bubbletea model for the interactive TUI.
type Model struct {
	InputPath            string
	OutputMode           config.OutputMode
	OutputFormat         config.OutputFormat
	OutputDir            string
	MaxLines             string
	Workers              string
	GraphOutput          string
	err                  string
	suggestion           string
	AnalysisOutput       string
	FanThreshold         string
	GodThreshold         string
	MaxInheritanceDepth  string
	TopN                 string
	BetweennessThreshold string
	PropagationCutoff    string
	HotspotOutput        string
	HotspotFormat        config.OutputFormat
	HotspotSince         string
	HotspotMaxCommits    string
	HotspotMaxFiles      string
	HotspotMinCoChanges  string
	HotspotTopN          string
	FingerprintOutput    string
	FingerprintMinSim    string
	FingerprintMaxSim    string
	FingerprintModel     string
	screen               screen
	focus                field
	cmdIndex             int // selected command index on the command-select screen
	suggCycle            int
	FingerprintShowAll   bool
	FingerprintRerank    bool
	HotspotIncludeMerges bool
	CollectTest          bool
	Minify               bool
	Edges                bool
	Clean                bool
	confirmed            bool
}

const maxSuggestions = 5

// completeDirs returns directory names matching the given path prefix.
func completeDirs(input string) []string {
	if input == "" {
		input = "."
	}

	dir := input
	prefix := ""

	info, err := os.Stat(input)
	if err != nil || !info.IsDir() {
		dir = filepath.Dir(input)
		prefix = filepath.Base(input)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var matches []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if prefix != "" && !strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix)) {
			continue
		}
		full := filepath.Join(dir, name)
		matches = append(matches, full)
	}
	sort.Strings(matches)
	if len(matches) > maxSuggestions {
		matches = matches[:maxSuggestions]
	}
	return matches
}

// refreshSuggestion updates the inline suggestion for the current directory field.
func (m *Model) refreshSuggestion() {
	var input string
	switch m.focus {
	case fieldInputPath:
		input = m.InputPath
	case fieldOutputDir:
		input = m.OutputDir
	default:
		m.suggestion = ""
		m.suggCycle = 0
		return
	}
	matches := completeDirs(input)
	if len(matches) == 0 {
		m.suggestion = ""
		m.suggCycle = 0
		return
	}
	idx := m.suggCycle % len(matches)
	m.suggestion = matches[idx]
}

// acceptSuggestion fills the current directory field with the suggestion.
func (m *Model) acceptSuggestion() {
	if m.suggestion == "" {
		return
	}
	val := m.suggestion + string(filepath.Separator)
	switch m.focus {
	case fieldInputPath:
		m.InputPath = val
	case fieldOutputDir:
		m.OutputDir = val
	default:
		return
	}
	m.suggCycle = 0
	m.refreshSuggestion()
}

// NewModel returns a Model with sensible defaults.
func NewModel() Model {
	return Model{
		screen:               screenCommandSelect,
		OutputMode:           config.DefaultParseOutputMode,
		OutputFormat:         config.DefaultParseOutputFormat,
		OutputDir:            config.DefaultParseOutputDir,
		MaxLines:             strconv.Itoa(config.DefaultParseMaxLines),
		Workers:              strconv.Itoa(config.DefaultWorkers),
		GraphOutput:          config.DefaultGraphOutput,
		AnalysisOutput:       config.DefaultAnalyzeOutput,
		FanThreshold:         strconv.Itoa(config.DefaultAnalyzeFanThreshold),
		GodThreshold:         strconv.Itoa(config.DefaultAnalyzeGodThreshold),
		MaxInheritanceDepth:  strconv.Itoa(config.DefaultAnalyzeMaxInheritanceDepth),
		TopN:                 strconv.Itoa(config.DefaultAnalyzeTopN),
		BetweennessThreshold: strconv.FormatFloat(config.DefaultAnalyzeBetweennessThreshold, 'g', -1, 64),
		PropagationCutoff:    strconv.FormatFloat(config.DefaultAnalyzePropagationCutoff, 'g', -1, 64),
		HotspotOutput:        config.DefaultHotspotOutput,
		HotspotFormat:        config.DefaultHotspotFormat,
		HotspotSince:         config.DefaultHotspotSince,
		HotspotMaxCommits:    strconv.Itoa(config.DefaultHotspotMaxCommits),
		HotspotMaxFiles:      strconv.Itoa(config.DefaultHotspotMaxFilesPerCommit),
		HotspotMinCoChanges:  strconv.Itoa(config.DefaultHotspotMinCoChanges),
		HotspotTopN:          strconv.Itoa(config.DefaultHotspotTopN),
		HotspotIncludeMerges: config.DefaultHotspotIncludeMerges,
		FingerprintOutput:    config.DefaultFingerprintOutput,
		FingerprintMinSim:    strconv.Itoa(config.DefaultFingerprintMinSimilarity),
		FingerprintMaxSim:    strconv.Itoa(config.DefaultFingerprintMaxSimilarity),
		FingerprintModel:     config.DefaultFingerprintModelOverride,
		FingerprintShowAll:   config.DefaultFingerprintShowAll,
		FingerprintRerank:    config.DefaultFingerprintRerank,
		CollectTest:          config.DefaultCollectTest,
		Minify:               config.DefaultParseMinify,
		Edges:                config.DefaultParseEdges,
		Clean:                config.DefaultParseClean,
	}
}

// SelectedCommandKind identifies which subcommand the user picked in the TUI.
type SelectedCommandKind int

// Selected command kinds.
const (
	CmdParse SelectedCommandKind = iota
	CmdGraphShow
	CmdGraphAnalyze
	CmdGraphHotspots
	CmdFingerprint
)

// SelectedCommand returns the kind of the command chosen on the selection screen.
func (m *Model) SelectedCommand() SelectedCommandKind {
	if m.cmdIndex >= 0 && m.cmdIndex < len(commands) {
		switch commands[m.cmdIndex].Name {
		case "graph show":
			return CmdGraphShow
		case "graph analyze":
			return CmdGraphAnalyze
		case "graph hotspots":
			return CmdGraphHotspots
		case "fingerprint":
			return CmdFingerprint
		}
	}
	return CmdParse
}

// common builds the shared fields used by every per-command config.
func (m *Model) common() config.Common {
	w, _ := strconv.Atoi(m.Workers)
	return config.Common{
		InputPath:   m.InputPath,
		Workers:     w,
		CollectTest: m.CollectTest,
		Verbose:     config.DefaultVerbose,
	}
}

// ToParseConfig converts the TUI model state into a ParseConfig.
func (m *Model) ToParseConfig() config.ParseConfig {
	ml, _ := strconv.Atoi(m.MaxLines)
	return config.ParseConfig{
		Common:       m.common(),
		OutputDir:    m.OutputDir,
		OutputMode:   m.OutputMode,
		OutputFormat: m.OutputFormat,
		MaxLines:     ml,
		Minify:       m.Minify,
		Edges:        m.Edges,
		Clean:        m.Clean,
	}
}

// ToGraphConfig converts the TUI model state into a GraphConfig.
func (m *Model) ToGraphConfig() config.GraphConfig {
	return config.GraphConfig{
		Common: m.common(),
		Output: m.GraphOutput,
	}
}

// ToAnalyzeConfig converts the TUI model state into an AnalyzeConfig.
func (m *Model) ToAnalyzeConfig() config.AnalyzeConfig {
	ft, _ := strconv.Atoi(m.FanThreshold)
	gt, _ := strconv.Atoi(m.GodThreshold)
	mid, _ := strconv.Atoi(m.MaxInheritanceDepth)
	tn, _ := strconv.Atoi(m.TopN)
	bt, _ := strconv.ParseFloat(m.BetweennessThreshold, 64)
	pc, _ := strconv.ParseFloat(m.PropagationCutoff, 64)
	return config.AnalyzeConfig{
		Common:               m.common(),
		Output:               m.AnalysisOutput,
		FanThreshold:         ft,
		GodThreshold:         gt,
		MaxInheritanceDepth:  mid,
		TopN:                 tn,
		BetweennessThreshold: bt,
		PropagationCutoff:    pc,
	}
}

// ToHotspotConfig converts the TUI model state into a HotspotConfig.
func (m *Model) ToHotspotConfig() config.HotspotConfig {
	maxCommits, _ := strconv.Atoi(m.HotspotMaxCommits)
	maxFiles, _ := strconv.Atoi(m.HotspotMaxFiles)
	minCoChanges, _ := strconv.Atoi(m.HotspotMinCoChanges)
	topN, _ := strconv.Atoi(m.HotspotTopN)
	return config.HotspotConfig{
		Common:            m.common(),
		Output:            m.HotspotOutput,
		Format:            m.HotspotFormat,
		Since:             m.HotspotSince,
		MaxCommits:        maxCommits,
		MaxFilesPerCommit: maxFiles,
		MinCoChanges:      minCoChanges,
		TopN:              topN,
		IncludeMerges:     m.HotspotIncludeMerges,
	}
}

// ToFingerprintConfig converts the TUI model state into a FingerprintConfig.
func (m *Model) ToFingerprintConfig() config.FingerprintConfig {
	minSim, _ := strconv.Atoi(m.FingerprintMinSim)
	maxSim, _ := strconv.Atoi(m.FingerprintMaxSim)
	return config.FingerprintConfig{
		Common:     m.common(),
		Output:     m.FingerprintOutput,
		EmbedModel: config.ResolveFingerprintEmbedModel(m.FingerprintRerank, m.FingerprintModel),
		MinSim:     minSim,
		MaxSim:     maxSim,
		ShowAll:    m.FingerprintShowAll,
	}
}

// Confirmed reports whether the user pressed enter on the confirm button.
func (m *Model) Confirmed() bool { return m.confirmed }

// Init implements tea.Model.
//
//nolint:gocritic // hugeParam: value receiver required by tea.Model interface.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
//
//nolint:gocritic // hugeParam: value receiver required by tea.Model interface.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch m.screen {
		case screenCommandSelect:
			return m.handleCommandSelectKey(msg)
		case screenParseOptions:
			return m.handleParseOptionsKey(msg)
		case screenGraphOptions:
			return m.handleGraphOptionsKey(msg)
		case screenGraphAnalyzeOptions:
			return m.handleGraphAnalyzeOptionsKey(msg)
		case screenGraphHotspotOptions:
			return m.handleGraphHotspotOptionsKey(msg)
		case screenFingerprintOptions:
			return m.handleFingerprintOptionsKey(msg)
		}
	}
	return m, nil
}

// --- Command selection screen ---

//nolint:gocritic // hugeParam: called from value-receiver tea.Model methods.
func (m Model) handleCommandSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cmdIndex > 0 {
			m.cmdIndex--
		}
		return m, nil
	case "down", "j":
		if m.cmdIndex < len(commands) { // len(commands) is the quit entry
			m.cmdIndex++
		}
		return m, nil
	case "enter":
		if m.cmdIndex == len(commands) { // quit entry selected
			return m, tea.Quit
		}
		switch commands[m.cmdIndex].Name {
		case "graph show":
			m.screen = screenGraphOptions
			m.focus = fieldInputPath
		case "graph analyze":
			m.screen = screenGraphAnalyzeOptions
			m.focus = fieldInputPath
		case "graph hotspots":
			m.screen = screenGraphHotspotOptions
			m.focus = fieldInputPath
		case "fingerprint":
			m.screen = screenFingerprintOptions
			m.focus = fieldInputPath
		default:
			m.screen = screenParseOptions
			m.focus = fieldInputPath
		}
		return m, nil
	}
	return m, nil
}

// --- Parse options screen (existing behavior) ---

// nextField returns the next field, skipping OutputDir when inline mode is active.
func (m *Model) nextField() field {
	next := (m.focus + 1) % totalFields
	if next == fieldOutputDir && m.OutputMode == config.OutputInline {
		next = (next + 1) % totalFields
	}
	return next
}

// prevField returns the previous field, skipping OutputDir when inline mode is active.
func (m *Model) prevField() field {
	prev := (m.focus - 1 + totalFields) % totalFields
	if prev == fieldOutputDir && m.OutputMode == config.OutputInline {
		prev = (prev - 1 + totalFields) % totalFields
	}
	return prev
}

// cycleOutputMode advances the OutputMode to the next valid value.
func (m *Model) cycleOutputMode() {
	modes := config.ValidOutputModes()
	for i, mode := range modes {
		if mode == m.OutputMode {
			m.OutputMode = modes[(i+1)%len(modes)]
			return
		}
	}
	// Fallback if current mode is somehow invalid.
	m.OutputMode = config.OutputDirectoryFlat
}

// cycleOutputFormat advances the OutputFormat to the next valid value.
func (m *Model) cycleOutputFormat() {
	formats := config.ValidOutputFormats()
	for i, format := range formats {
		if format == m.OutputFormat {
			m.OutputFormat = formats[(i+1)%len(formats)]
			return
		}
	}
	// Fallback if current format is somehow invalid.
	m.OutputFormat = config.OutputFormatSKT
}

//nolint:gocritic // hugeParam: called from value-receiver tea.Model methods.
func (m Model) handleParseOptionsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		if m.focus != fieldInputPath && m.focus != fieldOutputDir && m.focus != fieldWorkers && m.focus != fieldMaxLines {
			return m, tea.Quit
		}
		// 'q' while editing a text field — fall through to typing
	case "esc":
		// Go back to command selection.
		m.screen = screenCommandSelect
		m.err = ""
		return m, nil
	}

	switch key {
	case "tab":
		// On directory fields: accept suggestion, or cycle if already accepted, or move on.
		if m.focus == fieldInputPath || m.focus == fieldOutputDir {
			if m.suggestion != "" {
				// If current value already matches the suggestion, cycle to next match.
				current := m.InputPath
				if m.focus == fieldOutputDir {
					current = m.OutputDir
				}
				if strings.TrimRight(current, string(filepath.Separator)) == m.suggestion {
					m.suggCycle++
					m.refreshSuggestion()
					if m.suggestion != "" {
						m.acceptSuggestion()
						return m, nil
					}
				} else {
					m.acceptSuggestion()
					return m, nil
				}
			}
		}
		m.focus = m.nextField()
		m.suggCycle = 0
		m.refreshSuggestion()
		return m, nil
	case "down":
		m.focus = m.nextField()
		m.suggCycle = 0
		m.refreshSuggestion()
		return m, nil
	case "shift+tab", "up":
		m.focus = m.prevField()
		m.suggCycle = 0
		m.refreshSuggestion()
		return m, nil
	}

	// Toggle / cycle fields
	switch m.focus {
	case fieldOutputMode:
		if isActivationKey(key) {
			m.cycleOutputMode()
		}
		return m, nil
	case fieldOutputFormat:
		if isActivationKey(key) {
			m.cycleOutputFormat()
		}
		return m, nil
	case fieldCollectTest:
		if isActivationKey(key) {
			m.CollectTest = !m.CollectTest
		}
		return m, nil
	case fieldMinify:
		if isActivationKey(key) {
			m.Minify = !m.Minify
		}
		return m, nil
	case fieldEdges:
		if isActivationKey(key) {
			m.Edges = !m.Edges
		}
		return m, nil
	case fieldClean:
		if isActivationKey(key) {
			m.Clean = !m.Clean
		}
		return m, nil
	case fieldConfirm:
		if key == "enter" {
			if err := m.validate(); err != "" {
				m.err = err
				return m, nil
			}
			m.confirmed = true
			return m, tea.Quit
		}
		return m, nil
	}

	// Text input fields
	switch m.focus {
	case fieldInputPath:
		m.InputPath = editText(m.InputPath, msg)
		m.suggCycle = 0
		m.refreshSuggestion()
	case fieldOutputDir:
		m.OutputDir = editText(m.OutputDir, msg)
		m.suggCycle = 0
		m.refreshSuggestion()
	case fieldMaxLines:
		m.MaxLines = editNumeric(m.MaxLines, msg)
	case fieldWorkers:
		m.Workers = editNumeric(m.Workers, msg)
	}
	return m, nil
}

func editText(val string, msg tea.KeyMsg) string {
	key := msg.String()
	switch key {
	case "backspace":
		if val != "" {
			return val[:len(val)-1]
		}
		return val
	case "enter", "tab", "shift+tab", "up", "down":
		return val
	default:
		if len(key) == 1 {
			return val + key
		}
		return val
	}
}

// editBackspace handles the shared backspace logic for numeric/float editors.
func editBackspace(val string) string {
	if val != "" {
		return val[:len(val)-1]
	}
	return val
}

func editNumeric(val string, msg tea.KeyMsg) string {
	key := msg.String()
	if key == "backspace" {
		return editBackspace(val)
	}
	if len(key) == 1 && key[0] >= '0' && key[0] <= '9' {
		return val + key
	}
	return val
}

func editFloat(val string, msg tea.KeyMsg) string {
	key := msg.String()
	if key == "backspace" {
		return editBackspace(val)
	}
	if len(key) == 1 && ((key[0] >= '0' && key[0] <= '9') || (key[0] == '.' && !strings.Contains(val, "."))) {
		return val + key
	}
	return val
}

func isActivationKey(key string) bool {
	return key == "enter" || key == "space" || key == " "
}

//nolint:gocritic // hugeParam: called from value-receiver tea.Model methods.
func (m Model) validate() string {
	if m.InputPath == "" {
		return "input path is required"
	}
	return ""
}

// --- View ---

// View implements tea.Model.
//
//nolint:gocritic // hugeParam: value receiver required by tea.Model interface.
func (m Model) View() tea.View {
	switch m.screen {
	case screenCommandSelect:
		return m.viewCommandSelect()
	case screenParseOptions:
		return m.viewParseOptions()
	case screenGraphOptions:
		return m.viewGraphOptions()
	case screenGraphAnalyzeOptions:
		return m.viewGraphAnalyzeOptions()
	case screenGraphHotspotOptions:
		return m.viewGraphHotspotOptions()
	case screenFingerprintOptions:
		return m.viewFingerprintOptions()
	}
	return tea.NewView("")
}

//nolint:gocritic // hugeParam: called from value-receiver View().
func (m Model) viewCommandSelect() tea.View {
	s := banner() + "\n"
	s += labelStyle.Render("  Select a command:") + "\n\n"

	for i, cmd := range commands {
		if i == m.cmdIndex {
			s += activeStyle.Render("> "+cmd.Name) + "  " + descStyle.Render(cmd.Desc) + "\n"
		} else {
			s += "  " + labelStyle.Render(cmd.Name) + "  " + descStyle.Render(cmd.Desc) + "\n"
		}
	}

	s += "\n"
	if m.cmdIndex == len(commands) {
		s += quitStyle.Render("> [ quit ]") + "\n"
	} else {
		s += "  " + quitStyle.Render("[ quit ]") + "\n"
	}

	s += "\n" + hintStyle.Render("↑/↓ navigate • enter select • q quit") + "\n"
	return tea.NewView(s)
}

//nolint:gocritic // hugeParam: called from value-receiver View().
func (m Model) viewParseOptions() tea.View {
	s := banner() + "\n"
	s += labelStyle.Render("  Command: ") + activeStyle.Render(commands[m.cmdIndex].Name) + "\n\n"

	s += m.textField("Input path:       ", m.InputPath, fieldInputPath)
	s += m.selectionField("Output mode:      ", string(m.OutputMode), fieldOutputMode)
	s += m.selectionField("Output format:    ", string(m.OutputFormat), fieldOutputFormat)
	if m.OutputMode != config.OutputInline {
		s += m.textField("Output directory: ", m.OutputDir, fieldOutputDir)
	}
	s += m.textField("Max lines:        ", m.MaxLines, fieldMaxLines)
	s += m.toggleField("Collect test files", m.CollectTest, fieldCollectTest)
	s += m.toggleField("Minify output", m.Minify, fieldMinify)
	s += m.toggleField("Include edges", m.Edges, fieldEdges)
	s += m.toggleField("Clean output dir", m.Clean, fieldClean)
	s += m.textField("Workers (0=auto): ", m.Workers, fieldWorkers)

	if m.focus == fieldConfirm {
		s += "\n  " + confirmStyle.Render("> [ Confirm ]") + "\n"
	} else {
		s += "\n    " + confirmStyle.Render("[ Confirm ]") + "\n"
	}

	if m.err != "" {
		s += "\n  " + errorStyle.Render("error: "+m.err) + "\n"
	}

	s += "\n" + hintStyle.Render("tab complete/next • shift+tab prev • enter/space toggle • esc back • ctrl+c quit") + "\n"

	return tea.NewView(s)
}

// --- Graph options screen ---

// graph fields reuse fieldInputPath, fieldCollectTest, fieldWorkers, fieldConfirm
// plus a new fieldGraphOutput mapped to fieldOutputDir slot.

//nolint:gocritic // hugeParam: called from value-receiver tea.Model methods.
func (m Model) handleGraphOptionsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		if m.focus != fieldInputPath && m.focus != fieldOutputDir && m.focus != fieldWorkers {
			return m, tea.Quit
		}
	case "esc":
		m.screen = screenCommandSelect
		m.err = ""
		return m, nil
	}

	switch key {
	case "tab":
		if m.focus == fieldInputPath || m.focus == fieldOutputDir {
			if m.suggestion != "" {
				current := m.InputPath
				if m.focus == fieldOutputDir {
					current = m.GraphOutput
				}
				if strings.TrimRight(current, string(filepath.Separator)) == m.suggestion {
					m.suggCycle++
					m.refreshSuggestion()
					if m.suggestion != "" {
						m.acceptSuggestion()
						return m, nil
					}
				} else {
					m.acceptSuggestion()
					return m, nil
				}
			}
		}
		switch m.focus {
		case fieldInputPath:
			m.focus = fieldOutputDir
		case fieldOutputDir:
			m.focus = fieldCollectTest
		case fieldCollectTest:
			m.focus = fieldWorkers
		case fieldWorkers:
			m.focus = fieldConfirm
		default:
			m.focus = fieldInputPath
		}
		m.suggCycle = 0
		m.refreshSuggestion()
		return m, nil
	case "down":
		switch m.focus {
		case fieldInputPath:
			m.focus = fieldOutputDir
		case fieldOutputDir:
			m.focus = fieldCollectTest
		case fieldCollectTest:
			m.focus = fieldWorkers
		case fieldWorkers:
			m.focus = fieldConfirm
		default:
			m.focus = fieldInputPath
		}
		m.suggCycle = 0
		m.refreshSuggestion()
		return m, nil
	case "shift+tab", "up":
		switch m.focus {
		case fieldOutputDir:
			m.focus = fieldInputPath
		case fieldCollectTest:
			m.focus = fieldOutputDir
		case fieldWorkers:
			m.focus = fieldCollectTest
		case fieldConfirm:
			m.focus = fieldWorkers
		default:
			m.focus = fieldConfirm
		}
		m.suggCycle = 0
		m.refreshSuggestion()
		return m, nil
	}

	switch m.focus {
	case fieldCollectTest:
		if isActivationKey(key) {
			m.CollectTest = !m.CollectTest
		}
		return m, nil
	case fieldConfirm:
		if key == "enter" {
			if errMsg := m.validate(); errMsg != "" {
				m.err = errMsg
				return m, nil
			}
			m.confirmed = true
			return m, tea.Quit
		}
		return m, nil
	}

	switch m.focus {
	case fieldInputPath:
		m.InputPath = editText(m.InputPath, msg)
		m.suggCycle = 0
		m.refreshSuggestion()
	case fieldOutputDir:
		m.GraphOutput = editText(m.GraphOutput, msg)
		m.suggCycle = 0
		m.refreshSuggestion()
	case fieldWorkers:
		m.Workers = editNumeric(m.Workers, msg)
	}

	return m, nil
}

//nolint:gocritic // hugeParam: called from value-receiver View().
func (m Model) viewGraphOptions() tea.View {
	s := banner() + "\n"
	s += labelStyle.Render("  Command: ") + activeStyle.Render("graph show") + "\n\n"

	s += m.textField("Input path:       ", m.InputPath, fieldInputPath)
	s += m.textField("Output file:      ", m.GraphOutput, fieldOutputDir)
	s += m.toggleField("Collect test files", m.CollectTest, fieldCollectTest)
	s += m.textField("Workers (0=auto): ", m.Workers, fieldWorkers)

	if m.focus == fieldConfirm {
		s += "\n  " + confirmStyle.Render("> [ Confirm ]") + "\n"
	} else {
		s += "\n    " + confirmStyle.Render("[ Confirm ]") + "\n"
	}

	if m.err != "" {
		s += "\n  " + errorStyle.Render("error: "+m.err) + "\n"
	}

	s += "\n" + hintStyle.Render("tab complete/next • shift+tab prev • space toggle • esc back • ctrl+c quit") + "\n"

	return tea.NewView(s)
}

// --- Graph analyze options screen ---

// graphAnalyzeFields defines the field navigation order for the analyze screen.
var graphAnalyzeFields = []field{
	fieldInputPath,
	fieldOutputDir,
	fieldCollectTest,
	fieldWorkers,
	fieldFanThreshold,
	fieldGodThreshold,
	fieldMaxInheritanceDepth,
	fieldTopN,
	fieldBetweennessThreshold,
	fieldPropagationCutoff,
	fieldConfirm,
}

// nextInFields returns the next field in the given ordered slice, wrapping around.
func nextInFields(fields []field, current field) field {
	for i, f := range fields {
		if f == current {
			return fields[(i+1)%len(fields)]
		}
	}
	return fields[0]
}

// prevInFields returns the previous field in the given ordered slice, wrapping around.
func prevInFields(fields []field, current field) field {
	for i, f := range fields {
		if f == current {
			return fields[(i-1+len(fields))%len(fields)]
		}
	}
	return fields[len(fields)-1]
}

func graphAnalyzeNextField(current field) field {
	return nextInFields(graphAnalyzeFields, current)
}

func graphAnalyzePrevField(current field) field {
	return prevInFields(graphAnalyzeFields, current)
}

func fingerprintNextField(current field) field {
	return nextInFields(fingerprintFields, current)
}

func fingerprintPrevField(current field) field {
	return prevInFields(fingerprintFields, current)
}

//nolint:gocritic // hugeParam: called from value-receiver tea.Model methods.
func (m Model) handleGraphAnalyzeOptionsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Text-input fields where 'q' should be typed, not quit.
	isTextField := m.focus == fieldInputPath || m.focus == fieldOutputDir ||
		m.focus == fieldWorkers || m.focus == fieldFanThreshold ||
		m.focus == fieldGodThreshold || m.focus == fieldMaxInheritanceDepth ||
		m.focus == fieldTopN || m.focus == fieldBetweennessThreshold ||
		m.focus == fieldPropagationCutoff

	switch key {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		if !isTextField {
			return m, tea.Quit
		}
	case "esc":
		m.screen = screenCommandSelect
		m.err = ""
		return m, nil
	}

	switch key {
	case "tab":
		if m.focus == fieldInputPath {
			if m.suggestion != "" {
				current := m.InputPath
				if strings.TrimRight(current, string(filepath.Separator)) == m.suggestion {
					m.suggCycle++
					m.refreshSuggestion()
					if m.suggestion != "" {
						m.acceptSuggestion()
						return m, nil
					}
				} else {
					m.acceptSuggestion()
					return m, nil
				}
			}
		}
		m.focus = graphAnalyzeNextField(m.focus)
		m.suggCycle = 0
		m.refreshSuggestion()
		return m, nil
	case "down":
		m.focus = graphAnalyzeNextField(m.focus)
		m.suggCycle = 0
		m.refreshSuggestion()
		return m, nil
	case "shift+tab", "up":
		m.focus = graphAnalyzePrevField(m.focus)
		m.suggCycle = 0
		m.refreshSuggestion()
		return m, nil
	}

	switch m.focus {
	case fieldCollectTest:
		if isActivationKey(key) {
			m.CollectTest = !m.CollectTest
		}
		return m, nil
	case fieldConfirm:
		if key == "enter" {
			if errMsg := m.validate(); errMsg != "" {
				m.err = errMsg
				return m, nil
			}
			m.confirmed = true
			return m, tea.Quit
		}
		return m, nil
	}

	switch m.focus {
	case fieldInputPath:
		m.InputPath = editText(m.InputPath, msg)
		m.suggCycle = 0
		m.refreshSuggestion()
	case fieldOutputDir:
		m.AnalysisOutput = editText(m.AnalysisOutput, msg)
	case fieldWorkers:
		m.Workers = editNumeric(m.Workers, msg)
	case fieldFanThreshold:
		m.FanThreshold = editNumeric(m.FanThreshold, msg)
	case fieldGodThreshold:
		m.GodThreshold = editNumeric(m.GodThreshold, msg)
	case fieldMaxInheritanceDepth:
		m.MaxInheritanceDepth = editNumeric(m.MaxInheritanceDepth, msg)
	case fieldTopN:
		m.TopN = editNumeric(m.TopN, msg)
	case fieldBetweennessThreshold:
		m.BetweennessThreshold = editFloat(m.BetweennessThreshold, msg)
	case fieldPropagationCutoff:
		m.PropagationCutoff = editFloat(m.PropagationCutoff, msg)
	}

	return m, nil
}

//nolint:gocritic // hugeParam: called from value-receiver View().
func (m Model) viewGraphAnalyzeOptions() tea.View {
	s := banner() + "\n"
	s += labelStyle.Render("  Command: ") + activeStyle.Render("graph analyze") + "\n\n"

	s += m.textField("Input path:         ", m.InputPath, fieldInputPath)
	s += m.textField("Output file:        ", m.AnalysisOutput, fieldOutputDir)
	s += m.toggleField("Collect test files", m.CollectTest, fieldCollectTest)
	s += m.textField("Workers (0=auto):   ", m.Workers, fieldWorkers)
	s += m.textField("Fan threshold:      ", m.FanThreshold, fieldFanThreshold)
	s += m.textField("God threshold:      ", m.GodThreshold, fieldGodThreshold)
	s += m.textField("Max inherit depth:  ", m.MaxInheritanceDepth, fieldMaxInheritanceDepth)
	s += m.textField("Top N:              ", m.TopN, fieldTopN)
	s += m.textField("Betweenness thresh: ", m.BetweennessThreshold, fieldBetweennessThreshold)
	s += m.textField("Propagation cutoff: ", m.PropagationCutoff, fieldPropagationCutoff)

	if m.focus == fieldConfirm {
		s += "\n  " + confirmStyle.Render("> [ Confirm ]") + "\n"
	} else {
		s += "\n    " + confirmStyle.Render("[ Confirm ]") + "\n"
	}

	if m.err != "" {
		s += "\n  " + errorStyle.Render("error: "+m.err) + "\n"
	}

	s += "\n" + hintStyle.Render("tab next • shift+tab prev • space toggle • esc back • ctrl+c quit") + "\n"

	return tea.NewView(s)
}

// --- Graph hotspot options screen ---

var graphHotspotFields = []field{
	fieldInputPath,
	fieldOutputDir,
	fieldHotspotFormat,
	fieldHotspotSince,
	fieldHotspotMaxCommits,
	fieldHotspotMaxFilesPerCommit,
	fieldHotspotMinCoChanges,
	fieldHotspotTopN,
	fieldHotspotIncludeMerges,
	fieldCollectTest,
	fieldWorkers,
	fieldConfirm,
}

func graphHotspotNextField(current field) field {
	return nextInFields(graphHotspotFields, current)
}

func graphHotspotPrevField(current field) field {
	return prevInFields(graphHotspotFields, current)
}

func (m *Model) cycleHotspotFormat() {
	formats := config.ValidOutputFormats()
	for i, format := range formats {
		if format == m.HotspotFormat {
			m.HotspotFormat = formats[(i+1)%len(formats)]
			return
		}
	}
	m.HotspotFormat = config.OutputFormatSKT
}

//nolint:gocritic // hugeParam: called from value-receiver tea.Model methods.
func (m Model) handleGraphHotspotOptionsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	isTextField := m.focus == fieldInputPath || m.focus == fieldOutputDir ||
		m.focus == fieldHotspotSince || m.focus == fieldHotspotMaxCommits ||
		m.focus == fieldHotspotMaxFilesPerCommit || m.focus == fieldHotspotMinCoChanges ||
		m.focus == fieldHotspotTopN || m.focus == fieldWorkers

	switch key {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		if !isTextField {
			return m, tea.Quit
		}
	case "esc":
		m.screen = screenCommandSelect
		m.err = ""
		return m, nil
	}

	switch key {
	case "tab":
		if m.focus == fieldInputPath && m.suggestion != "" {
			current := m.InputPath
			if strings.TrimRight(current, string(filepath.Separator)) == m.suggestion {
				m.suggCycle++
				m.refreshSuggestion()
				if m.suggestion != "" {
					m.acceptSuggestion()
					return m, nil
				}
			} else {
				m.acceptSuggestion()
				return m, nil
			}
		}
		m.focus = graphHotspotNextField(m.focus)
		m.suggCycle = 0
		m.refreshSuggestion()
		return m, nil
	case "down":
		m.focus = graphHotspotNextField(m.focus)
		m.suggCycle = 0
		m.refreshSuggestion()
		return m, nil
	case "shift+tab", "up":
		m.focus = graphHotspotPrevField(m.focus)
		m.suggCycle = 0
		m.refreshSuggestion()
		return m, nil
	}

	switch m.focus {
	case fieldHotspotFormat:
		if isActivationKey(key) {
			m.cycleHotspotFormat()
		}
		return m, nil
	case fieldHotspotIncludeMerges:
		if isActivationKey(key) {
			m.HotspotIncludeMerges = !m.HotspotIncludeMerges
		}
		return m, nil
	case fieldCollectTest:
		if isActivationKey(key) {
			m.CollectTest = !m.CollectTest
		}
		return m, nil
	case fieldConfirm:
		if key == "enter" {
			if errMsg := m.validate(); errMsg != "" {
				m.err = errMsg
				return m, nil
			}
			m.confirmed = true
			return m, tea.Quit
		}
		return m, nil
	}

	switch m.focus {
	case fieldInputPath:
		m.InputPath = editText(m.InputPath, msg)
		m.suggCycle = 0
		m.refreshSuggestion()
	case fieldOutputDir:
		m.HotspotOutput = editText(m.HotspotOutput, msg)
	case fieldHotspotSince:
		m.HotspotSince = editText(m.HotspotSince, msg)
	case fieldHotspotMaxCommits:
		m.HotspotMaxCommits = editNumeric(m.HotspotMaxCommits, msg)
	case fieldHotspotMaxFilesPerCommit:
		m.HotspotMaxFiles = editNumeric(m.HotspotMaxFiles, msg)
	case fieldHotspotMinCoChanges:
		m.HotspotMinCoChanges = editNumeric(m.HotspotMinCoChanges, msg)
	case fieldHotspotTopN:
		m.HotspotTopN = editNumeric(m.HotspotTopN, msg)
	case fieldWorkers:
		m.Workers = editNumeric(m.Workers, msg)
	}
	return m, nil
}

//nolint:gocritic // hugeParam: called from value-receiver View().
func (m Model) viewGraphHotspotOptions() tea.View {
	s := banner() + "\n"
	s += labelStyle.Render("  Command: ") + activeStyle.Render("graph hotspots") + "\n\n"

	s += m.textField("Input path:          ", m.InputPath, fieldInputPath)
	s += m.textField("Output file:         ", m.HotspotOutput, fieldOutputDir)
	s += m.selectionField("Output format:       ", string(m.HotspotFormat), fieldHotspotFormat)
	s += m.textField("History window:      ", m.HotspotSince, fieldHotspotSince)
	s += m.textField("Max commits:         ", m.HotspotMaxCommits, fieldHotspotMaxCommits)
	s += m.textField("Max files/commit:    ", m.HotspotMaxFiles, fieldHotspotMaxFilesPerCommit)
	s += m.textField("Min cochanges:       ", m.HotspotMinCoChanges, fieldHotspotMinCoChanges)
	s += m.textField("Top N:               ", m.HotspotTopN, fieldHotspotTopN)
	s += m.toggleField("Include merge commits", m.HotspotIncludeMerges, fieldHotspotIncludeMerges)
	s += m.toggleField("Collect test files", m.CollectTest, fieldCollectTest)
	s += m.textField("Workers (0=auto):    ", m.Workers, fieldWorkers)

	if m.focus == fieldConfirm {
		s += "\n  " + confirmStyle.Render("> [ Confirm ]") + "\n"
	} else {
		s += "\n    " + confirmStyle.Render("[ Confirm ]") + "\n"
	}
	if m.err != "" {
		s += "\n  " + errorStyle.Render("error: "+m.err) + "\n"
	}
	s += "\n" + hintStyle.Render("tab next • shift+tab prev • space toggle • esc back • ctrl+c quit") + "\n"
	return tea.NewView(s)
}

// --- Fingerprint options screen ---

// fingerprintFields defines the field navigation order for the fingerprint screen.
var fingerprintFields = []field{
	fieldInputPath,
	fieldOutputDir,
	fieldFingerprintMinSim,
	fieldFingerprintMaxSim,
	fieldFingerprintRerank,
	fieldFingerprintModel,
	fieldFingerprintShowAll,
	fieldCollectTest,
	fieldWorkers,
	fieldConfirm,
}

//nolint:gocritic // hugeParam: called from value-receiver tea.Model methods.
func (m Model) handleFingerprintOptionsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	isTextField := m.focus == fieldInputPath || m.focus == fieldOutputDir ||
		m.focus == fieldWorkers || m.focus == fieldFingerprintMinSim ||
		m.focus == fieldFingerprintMaxSim || m.focus == fieldFingerprintModel

	switch key {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		if !isTextField {
			return m, tea.Quit
		}
	case "esc":
		m.screen = screenCommandSelect
		m.err = ""
		return m, nil
	}

	switch key {
	case "tab":
		if m.focus == fieldInputPath {
			if m.suggestion != "" {
				current := m.InputPath
				if strings.TrimRight(current, string(filepath.Separator)) == m.suggestion {
					m.suggCycle++
					m.refreshSuggestion()
					if m.suggestion != "" {
						m.acceptSuggestion()
						return m, nil
					}
				} else {
					m.acceptSuggestion()
					return m, nil
				}
			}
		}
		m.focus = fingerprintNextField(m.focus)
		m.suggCycle = 0
		m.refreshSuggestion()
		return m, nil
	case "down":
		m.focus = fingerprintNextField(m.focus)
		m.suggCycle = 0
		m.refreshSuggestion()
		return m, nil
	case "shift+tab", "up":
		m.focus = fingerprintPrevField(m.focus)
		m.suggCycle = 0
		m.refreshSuggestion()
		return m, nil
	}

	switch m.focus {
	case fieldCollectTest:
		if isActivationKey(key) {
			m.CollectTest = !m.CollectTest
		}
		return m, nil
	case fieldFingerprintRerank:
		if isActivationKey(key) {
			m.FingerprintRerank = !m.FingerprintRerank
		}
		return m, nil
	case fieldFingerprintShowAll:
		if isActivationKey(key) {
			m.FingerprintShowAll = !m.FingerprintShowAll
		}
		return m, nil
	case fieldConfirm:
		if key == "enter" {
			if errMsg := m.validate(); errMsg != "" {
				m.err = errMsg
				return m, nil
			}
			m.confirmed = true
			return m, tea.Quit
		}
		return m, nil
	}

	switch m.focus {
	case fieldInputPath:
		m.InputPath = editText(m.InputPath, msg)
		m.suggCycle = 0
		m.refreshSuggestion()
	case fieldOutputDir:
		m.FingerprintOutput = editText(m.FingerprintOutput, msg)
	case fieldFingerprintMinSim:
		m.FingerprintMinSim = editNumeric(m.FingerprintMinSim, msg)
	case fieldFingerprintMaxSim:
		m.FingerprintMaxSim = editNumeric(m.FingerprintMaxSim, msg)
	case fieldFingerprintModel:
		m.FingerprintModel = editText(m.FingerprintModel, msg)
	case fieldWorkers:
		m.Workers = editNumeric(m.Workers, msg)
	}

	return m, nil
}

//nolint:gocritic // hugeParam: called from value-receiver View().
func (m Model) viewFingerprintOptions() tea.View {
	s := banner() + "\n"
	s += labelStyle.Render("  Command: ") + activeStyle.Render("fingerprint") + "\n\n"

	s += m.textField("Input path:         ", m.InputPath, fieldInputPath)
	s += m.textField("Output file:        ", m.FingerprintOutput, fieldOutputDir)
	s += m.textField("Min similarity (%): ", m.FingerprintMinSim, fieldFingerprintMinSim)
	s += m.textField("Max similarity (%): ", m.FingerprintMaxSim, fieldFingerprintMaxSim)
	s += m.toggleField("Semantic reranking  ", m.FingerprintRerank, fieldFingerprintRerank)
	s += m.textField("Model override:     ", m.FingerprintModel, fieldFingerprintModel)
	s += m.toggleField("Show all fingerprints", m.FingerprintShowAll, fieldFingerprintShowAll)
	s += m.toggleField("Collect test files", m.CollectTest, fieldCollectTest)
	s += m.textField("Workers (0=auto):   ", m.Workers, fieldWorkers)

	if m.focus == fieldConfirm {
		s += "\n  " + confirmStyle.Render("> [ Confirm ]") + "\n"
	} else {
		s += "\n    " + confirmStyle.Render("[ Confirm ]") + "\n"
	}

	if m.err != "" {
		s += "\n  " + errorStyle.Render("error: "+m.err) + "\n"
	}

	s += "\n" + hintStyle.Render("tab complete/next • shift+tab prev • space toggle • esc back • ctrl+c quit") + "\n"

	return tea.NewView(s)
}

//nolint:gocritic // hugeParam: called from value-receiver View().
func (m Model) textField(label, value string, f field) string {
	if m.focus == f {
		ghost := ""
		if (f == fieldInputPath || f == fieldOutputDir) && m.suggestion != "" {
			if strings.HasPrefix(m.suggestion, value) && len(m.suggestion) > len(value) {
				ghost = suggestionStyle.Render(m.suggestion[len(value):])
			}
		}
		return activeStyle.Render("> "+label) + value + ghost + "\n"
	}
	return "  " + labelStyle.Render(label) + value + "\n"
}

//nolint:gocritic // hugeParam: called from value-receiver View().
func (m Model) selectionField(label, value string, f field) string {
	if m.focus == f {
		return activeStyle.Render("> "+label) + "< " + value + " >" + "\n"
	}
	return "  " + labelStyle.Render(label) + value + "\n"
}

//nolint:gocritic // hugeParam: called from value-receiver View().
func (m Model) toggleField(label string, on bool, f field) string {
	check := "[ ]"
	if on {
		check = "[x]"
	}
	if m.focus == f {
		return activeStyle.Render("> "+check+" "+label) + "\n"
	}
	return "  " + labelStyle.Render(check+" "+label) + "\n"
}
