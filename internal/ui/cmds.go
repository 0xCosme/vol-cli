package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"volcli/internal/pulse"
)

// No Bubble Tea, trabalho que demora (chamar o pactl) roda num tea.Cmd e o
// resultado volta para o Update como uma mensagem. Assim a tela nunca trava.

const refreshEvery = 2 * time.Second

// loadedMsg traz o resultado de uma leitura do estado do áudio.
type loadedMsg struct {
	snap pulse.Snapshot
	ok   bool  // true se snap é válido (a leitura funcionou)
	err  error // erro da leitura ou da ação que a precedeu
}

// tickMsg dispara a atualização periódica da lista.
type tickMsg struct{}

// refresh lê o estado do áudio. actionErr é o erro da ação que acabou de rodar
// (ou nil): ele é repassado à tela, junto com a lista já atualizada.
func refresh(actionErr error) tea.Msg {
	snap, err := pulse.Load()
	if err != nil {
		return loadedMsg{err: err}
	}
	return loadedMsg{snap: snap, ok: true, err: actionErr}
}

func loadCmd() tea.Msg { return refresh(nil) }

func tickCmd() tea.Cmd {
	return tea.Tick(refreshEvery, func(time.Time) tea.Msg { return tickMsg{} })
}

// As ações abaixo executam a mudança e depois recarregam a lista.

func activateCmd(k pulse.Kind, e pulse.Entry) tea.Cmd {
	return func() tea.Msg { return refresh(pulse.Activate(k, e)) }
}

func volumeCmd(k pulse.Kind, e pulse.Entry, delta int) tea.Cmd {
	return func() tea.Msg { return refresh(pulse.SetVolume(k, e.Device, e.Percent+delta)) }
}

func muteCmd(k pulse.Kind, e pulse.Entry) tea.Cmd {
	return func() tea.Msg { return refresh(pulse.ToggleMute(k, e.Device)) }
}
