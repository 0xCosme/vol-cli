// Package ui é a interface de terminal, feita com Bubble Tea.
//
// O Bubble Tea organiza o programa em três partes:
//   - Model:  o estado da tela (este arquivo)
//   - Update: recebe mensagens (teclas, dados novos, relógio) e muda o estado
//   - View:   transforma o estado em texto (view.go)
package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"volcli/internal/pulse"
)

const volumeStep = 5 // porcentagem por toque nas setas

// Model é todo o estado da tela.
type Model struct {
	tab     pulse.Kind     // aba atual: Output ou Input
	entries pulse.Snapshot // últimas listas lidas do pactl
	cursor  [2]int         // item selecionado em cada aba (indexado por Kind)
	showAll bool           // mostrar também os dispositivos desconectados
	err     error
	width   int
	height  int
}

// Run abre a interface e só volta quando o usuário sair.
func Run() error {
	_, err := tea.NewProgram(Model{}).Run()
	return err
}

// Init roda uma vez ao abrir: lê o estado e agenda a primeira atualização.
func (m Model) Init() tea.Cmd { return tea.Batch(loadCmd, tickCmd()) }

// Update é chamado a cada mensagem e devolve o novo estado.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case loadedMsg:
		m.err = msg.err
		if msg.ok {
			m.entries = msg.snap
		}
		m.clampCursors()

	case tickMsg:
		return m, tea.Batch(loadCmd, tickCmd())

	case tea.KeyMsg:
		return m.handleKey(msg.String())
	}
	return m, nil
}

// handleKey trata uma tecla: primeiro a navegação, depois as ações no item selecionado.
func (m Model) handleKey(key string) (tea.Model, tea.Cmd) {
	rows, _ := m.rows(m.tab)
	cur := m.cursor[m.tab]

	switch key {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "tab":
		m.tab = m.tab.Other()
		return m, nil
	case "a":
		m.showAll = !m.showAll
		m.clampCursors()
		return m, nil
	case "up", "k":
		if cur > 0 {
			m.cursor[m.tab]--
		}
		return m, nil
	case "down", "j":
		if cur < len(rows)-1 {
			m.cursor[m.tab]++
		}
		return m, nil
	}

	if len(rows) == 0 {
		return m, nil
	}
	e := rows[cur]

	switch key {
	case "enter":
		return m, activateCmd(m.tab, e)
	case "m":
		if e.Controllable() {
			return m, muteCmd(m.tab, e)
		}
	case "left", "h":
		if e.Controllable() {
			return m, volumeCmd(m.tab, e, -volumeStep)
		}
	case "right", "l":
		if e.Controllable() {
			return m, volumeCmd(m.tab, e, volumeStep)
		}
	}
	return m, nil
}

// rows devolve as entradas visíveis de uma aba e quantas foram ocultadas.
// Desconectadas ficam ocultas, exceto o item em uso ou com a tecla "a" ativa.
func (m Model) rows(k pulse.Kind) (shown []pulse.Entry, hidden int) {
	for _, e := range m.entries[k] {
		if e.Unavailable() && !e.Current() && !m.showAll {
			hidden++
			continue
		}
		shown = append(shown, e)
	}
	return shown, hidden
}

// clampCursors impede que o cursor aponte para além do fim da lista
// (a lista pode encolher quando um dispositivo é desconectado).
func (m *Model) clampCursors() {
	for _, k := range pulse.Kinds {
		rows, _ := m.rows(k)
		m.cursor[k] = min(m.cursor[k], max(0, len(rows)-1))
	}
}
