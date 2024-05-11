package configs

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	Choices  []string
	Cursor   int
	Selected map[int]struct{}
	Quit     bool
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q", "esc":
			m.Quit = true
			return m, tea.Quit
		// The "up" and "k" keys move the Cursor up
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}

		// The "down" and "j" keys move the Cursor down
		case "down", "j":
			if m.Cursor < len(m.Choices)-1 {
				m.Cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the Selected state for the item that the Cursor is pointing at.
		case "enter", " ":
			_, ok := m.Selected[m.Cursor]
			if ok {
				delete(m.Selected, m.Cursor)
			} else {
				m.Selected[m.Cursor] = struct{}{}
				return m, tea.Quit
			}
		}
	}
	// Return the updated Model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m Model) View() string {
	// The header
	s := ">Available CRUD operations\n\n"

	// Iterate over our Choices
	for i, choice := range m.Choices {

		// Is the Cursor pointing at this choice?
		Cursor := " " // no Cursor
		if m.Cursor == i {
			Cursor = ">" // Cursor!
		}

		// Is this choice Selected?
		checked := " " // not Selected
		if _, ok := m.Selected[i]; ok {
			checked = "•" // Selected!
		}

		// Render the row
		s += fmt.Sprintf("%s (%s) %s\n", Cursor, checked, choice)
	}

	// The footer
	s += "\nPress q to Quit.\n"

	// Send the UI for rendering
	return s
}
