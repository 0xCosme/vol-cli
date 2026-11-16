package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"volcli/internal/pulse"
)

const (
	barWidth      = 30 // largura da barra de volume, em caracteres
	linesPerEntry = 3  // nome, barra e linha em branco
	chromeLines   = 9  // linhas ocupadas por abas, avisos e rodapé
)

var (
	activeTab = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(lipgloss.Color("12")).Padding(0, 2)
	idleTab   = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Padding(0, 2)
	selStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	dimStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	errStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

// View desenha a tela inteira como texto. É chamado depois de cada Update.
func (m Model) View() string {
	var b strings.Builder

	m.writeTabs(&b)
	if m.err != nil {
		b.WriteString(errStyle.Render("Erro: "+m.err.Error()) + "\n\n")
	}

	rows, hidden := m.rows(m.tab)
	if len(rows) == 0 && m.err == nil {
		b.WriteString(dimStyle.Render("Nenhum dispositivo encontrado") + "\n\n")
	}

	start, end := m.window(len(rows))
	textWidth := 0
	if m.width > 0 {
		textWidth = m.width - 4 // descontando o cursor "> "
	}
	for i := start; i < end; i++ {
		m.writeEntry(&b, rows[i], i == m.cursor[m.tab], textWidth)
	}

	if len(rows) > end-start {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  mostrando %d-%d de %d", start+1, end, len(rows))) + "\n")
	}
	if hidden > 0 {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  %d desconectado(s) oculto(s) — 'a' mostra", hidden)) + "\n")
	}

	help := "tab abas • ↑↓ escolher • ←→ volume • enter usar • m mudo • a desconectados • q sair"
	b.WriteString(dimStyle.Render(trunc(help, m.width)) + "\n")
	return b.String()
}

func (m Model) writeTabs(b *strings.Builder) {
	for _, k := range pulse.Kinds {
		if k == m.tab {
			b.WriteString(activeTab.Render(k.Title()))
		} else {
			b.WriteString(idleTab.Render(k.Title()))
		}
	}
	b.WriteString("\n\n")
}

// window calcula quais itens cabem na tela, mantendo o cursor no meio
// quando a lista é maior que o terminal. Sem isso o topo seria cortado.
func (m Model) window(total int) (start, end int) {
	visible := total
	if m.height > 0 {
		visible = max(1, (m.height-chromeLines)/linesPerEntry)
	}
	if total > visible {
		start = min(max(0, m.cursor[m.tab]-visible/2), total-visible)
	}
	return start, min(total, start+visible)
}

// writeEntry desenha um item: linha do nome e, embaixo, a barra de volume.
func (m Model) writeEntry(b *strings.Builder, e pulse.Entry, selected bool, textWidth int) {
	cursor, mark := "  ", " "
	if selected {
		cursor = "> "
	}
	if e.Current() {
		mark = "●"
	}

	title := mark + " " + e.Desc
	if e.Sub != "" {
		title += " · " + e.Sub
	}
	line := cursor + trunc(title, textWidth)
	switch {
	case selected:
		line = selStyle.Render(line)
	case e.Unavailable():
		line = dimStyle.Render(line)
	}

	// Só a porta ativa tem volume; as demais mostram a dica de uso.
	sub := dimStyle.Render("Enter para usar")
	if e.Controllable() {
		status := fmt.Sprintf("%3d%%", e.Percent)
		if e.Mute {
			status += " (mudo)"
		}
		sub = bar(e.Percent, e.Mute) + " " + status
	}
	b.WriteString(line + "\n      " + sub + "\n\n")
}

// bar desenha a barra de volume, cinza quando está mudo.
func bar(percent int, muted bool) string {
	filled := min(barWidth, max(0, percent*barWidth/100))
	s := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	if muted {
		return dimStyle.Render(s)
	}
	return s
}

// trunc corta o texto com "…" para caber na largura (0 = sem limite).
func trunc(s string, width int) string {
	r := []rune(s)
	if width > 0 && len(r) > width {
		return string(r[:max(0, width-1)]) + "…"
	}
	return s
}
