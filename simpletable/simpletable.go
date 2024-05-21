package simpletable

import (
	"encoding/json"
	"fmt"
	"os"

	structs "github.com/Parz1val02/cloud-cli/structs"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var BaseStyle = lipgloss.NewStyle().
	Bold(true).
	Padding(2).
	Margin(2).
	Align(lipgloss.Center).
	BorderStyle(lipgloss.RoundedBorder())

type Model struct {
	Table table.Model
	Quit  bool
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.Table.Focused() {
				m.Table.Blur()
			} else {
				m.Table.Focus()
			}
		case "q", "ctrl+c":
			m.Quit = true
			return m, tea.Quit
		case "enter":
			return m, tea.Quit
			//return m, tea.Batch(
			//	tea.Printf("Let's go to %s!", m.Table.SelectedRow()[0]),
			//)
		}
	}
	m.Table, cmd = m.Table.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return BaseStyle.Render(m.Table.View()) + "\n"
}

func MainTable() (string, error) {
	templateFile, err := os.Open("cloud.templates.json")
	if err != nil {
		fmt.Println("Error opening file: ", err.Error())
	}
	defer templateFile.Close()

	var templates structs.ListTemplates
	if err = json.NewDecoder(templateFile).Decode(&templates); err != nil {
		fmt.Println("Error parsing json: ", err.Error())
	}
	columns := []table.Column{
		{Title: "ID", Width: 30},
		{Title: "Name", Width: 30},
		{Title: "Description", Width: 30},
		{Title: "Creation Timestamp", Width: 20},
	}

	rows := []table.Row{
		{"1", "Tokyo", "Japan", "37,274,000"},
		{"2", "Delhi", "India", "32,065,760"},
		{"3", "Shanghai", "China", "28,516,904"},
		{"4", "Dhaka", "Bangladesh", "22,478,116"},
		{"5", "São Paulo", "Brazil", "22,429,800"},
		{"6", "Mexico City", "Mexico", "22,085,140"},
		{"7", "Cairo", "Egypt", "21,750,020"},
		{"8", "Beijing", "China", "21,333,332"},
		{"9", "Mumbai", "India", "20,961,472"},
		{"10", "Osaka", "Japan", "19,059,856"},
	}
	if templates.Result == "success" {
		for _, v := range templates.Templates {
			var row []string
			row = append(row, v.TemplateID, v.Name, v.Description, v.CreatedAt.Format("2006-01-02 15:04:05"))
			rows = append(rows, row)
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Bold(true).
		Align(lipgloss.Center).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("236")).
		Background(lipgloss.Color("12")).
		Bold(false)
	t.SetStyles(s)

	p := tea.NewProgram(Model{t, false})
	m, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	if m, ok := m.(Model); ok && m.Table.SelectedRow()[0] != "" {
		if m.Quit {
			return "", fmt.Errorf("\n---\nQuitting!\n")
		} else {
			return m.Table.SelectedRow()[0], nil
		}
	} else {
		return "", fmt.Errorf("Error runing program")
	}
}
