package simpletable

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	structs "github.com/Parz1val02/cloud-cli/structs"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var BaseStyle = lipgloss.NewStyle().
	Bold(true).
	Padding(1).
	Margin(1).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("12"))

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

func MainTable(token string) (string, error) {
	serverPort := 4444
	var templates structs.ListTemplates
	var jsonresp structs.NormalResponse
	requestURL := fmt.Sprintf("http://localhost:%d/templateservice/templates", serverPort)

	client := &http.Client{}
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}
	req.Header.Set("X-API-Key", token)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making http request: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&jsonresp)
		if err != nil {
			return "", fmt.Errorf("Error decoding response body: %v\n", err)
		}
		return "", fmt.Errorf("Unexpected status code: %d, Error: %s\n", resp.StatusCode, jsonresp.Msg)
	}
	err = json.NewDecoder(resp.Body).Decode(&templates)
	if err != nil {
		err = fmt.Errorf("Error decoding response body: %v", err)
		return "", err
	}

	//templateFile, err := os.Open("cloud.templates.json")
	//if err != nil {
	//	fmt.Println("Error opening file: ", err.Error())
	//}
	//defer templateFile.Close()

	//if err = json.NewDecoder(templateFile).Decode(&templates); err != nil {
	//	fmt.Println("Error parsing json: ", err.Error())
	//}

	columns := []table.Column{
		{Title: "ID", Width: 30},
		{Title: "Name", Width: 30},
		{Title: "Description", Width: 30},
		{Title: "Creation Timestamp", Width: 20},
		{Title: "Topology Type", Width: 20},
	}

	rows := []table.Row{}
	if templates.Result == "success" {
		for _, v := range templates.Templates {
			var row []string
			row = append(row, v.TemplateID, v.Name, v.Description, v.CreatedAt.Format("2006-01-02 15:04:05"), v.TopologyType)
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
		BorderBottom(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("236")).
		Background(lipgloss.Color("12")).
		Bold(false)
	t.SetStyles(s)

	p := tea.NewProgram(Model{t, false})
	m, err := p.Run()
	if err != nil {
		return "", err
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
