package tabs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	structs "github.com/Parz1val02/cloud-cli/structs"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jedib0t/go-pretty/v6/table"
)

type model struct {
	Tabs       []string
	TabContent []string
	activeTab  int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "right", "l", "n", "tab":
			m.activeTab = min(m.activeTab+1, len(m.Tabs)-1)
			return m, nil
		case "left", "h", "p", "shift+tab":
			m.activeTab = max(m.activeTab-1, 0)
			return m, nil
		}
	}

	return m, nil
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle          = lipgloss.NewStyle().Padding(1).MarginBottom(1)
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(lipgloss.Color("12")).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Copy().Border(activeTabBorder, true)
	windowStyle       = lipgloss.NewStyle().BorderForeground(lipgloss.Color("12")).Padding(1, 0).Align(lipgloss.Center).Border(lipgloss.NormalBorder()).UnsetBorderTop()
)

func (m model) View() string {
	doc := strings.Builder{}

	var renderedTabs []string

	for i, t := range m.Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.Tabs)-1, i == m.activeTab
		if isActive {
			style = activeTabStyle.Copy()
		} else {
			style = inactiveTabStyle.Copy()
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	doc.WriteString(windowStyle.Width((lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize())).Render(m.TabContent[m.activeTab]))
	return docStyle.Render(doc.String())
}

func MainTabs(templateId, token string) {
	serverPort := 4444
	var templateById structs.ListTemplateById
	var jsonresp structs.NormalResponse
	requestURL := fmt.Sprintf("http://localhost:%d/templateservice/templates/%s", serverPort, templateId)
	client := &http.Client{}
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}
	req.Header.Set("X-API-Key", token)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&jsonresp)
		if err != nil {
			fmt.Printf("Error decoding response body: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Unexpected status code: %d, Error: %s\n", resp.StatusCode, jsonresp.Msg)
		os.Exit(1)
	}
	err = json.NewDecoder(resp.Body).Decode(&templateById)
	if err != nil {
		fmt.Printf("Error decoding response body: %v", err)
		os.Exit(1)
	}

	//templateFile, err := os.Open("cloud.templatebyid.json")
	//if err != nil {
	//	fmt.Println("Error opening file: ", err.Error())
	//}
	//defer templateFile.Close()
	//if err = json.NewDecoder(templateFile).Decode(&templateById); err != nil {
	//	fmt.Println("Error parsing json: ", err.Error())
	//}

	if templateById.Result == "success" && templateById.Template.TemplateID == templateId {
		//out, err := json.Marshal(templateById.Template.Topology)
		//if err != nil {
		//	panic(err)
		//}
		style := lipgloss.NewStyle().
			Bold(true).Align(lipgloss.Left)
		tabs := []string{style.Render("\t\tTemplate Info\t\t"), style.Render("\t\tNodes\t\t"), style.Render("\t\tLinks\t\t")}
		info_string := fmt.Sprintf("ID: %s\n\nName: %s\n\nDescription: %s\n\nCreated at: %s\n\nTopology type: %s",
			templateById.Template.TemplateID, templateById.Template.Name, templateById.Template.Description, templateById.Template.CreatedAt.Format("2006-01-02 15:04:05"), templateById.Template.TopologyType)
		nodes := table.NewWriter()
		nodes.AppendHeader(table.Row{"ID", "Name", "Image", "CPU", "Memory", "Storage"})
		for _, v := range templateById.Template.Topology.Nodes {
			nodes.AppendRow(table.Row{v.NodeID, v.Name, v.Image, strconv.Itoa(v.Flavor.CPU), strconv.FormatFloat(float64(v.Flavor.Memory), 'f', 1, 32), strconv.FormatFloat(float64(v.Flavor.Storage), 'f', 1, 32)})
		}
		links := table.NewWriter()
		links.AppendHeader(table.Row{"ID", "Source", "Target"})
		for _, v := range templateById.Template.Topology.Links {
			links.AppendRow(table.Row{v.LinkID, v.Source, v.Target})
		}
		tabContent := []string{style.Render(info_string), style.Render(nodes.Render()), style.Render(links.Render())}
		m := model{Tabs: tabs, TabContent: tabContent}
		if _, err := tea.NewProgram(m).Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	}
}

func SliceInfoTabs(sliceId, token string) {
	serverPort := 4444
	var sliceById structs.ListSliceById
	var jsonresp structs.NormalResponse
	requestURL := fmt.Sprintf("http://localhost:%d/sliceservice/slices/%s", serverPort, sliceId)
	client := &http.Client{}
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}
	req.Header.Set("X-API-Key", token)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&jsonresp)
		if err != nil {
			fmt.Printf("Error decoding response body: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Unexpected status code: %d, Error: %s\n", resp.StatusCode, jsonresp.Msg)
		os.Exit(1)
	}
	err = json.NewDecoder(resp.Body).Decode(&sliceById)
	if err != nil {
		fmt.Printf("Error decoding response body: %v", err)
		os.Exit(1)
	}

	//templateFile, err := os.Open("cloud.templatebyid.json")
	//if err != nil {
	//	fmt.Println("Error opening file: ", err.Error())
	//}
	//defer templateFile.Close()
	//if err = json.NewDecoder(templateFile).Decode(&templateById); err != nil {
	//	fmt.Println("Error parsing json: ", err.Error())
	//}

	if sliceById.Result == "success" && sliceById.Slice.SliceID == sliceId {
		//out, err := json.Marshal(templateById.Template.Topology)
		//if err != nil {
		//	panic(err)
		//}
		style := lipgloss.NewStyle().
			Bold(true).Align(lipgloss.Left)
		tabs := []string{style.Render("\t\tTemplate Info\t\t"), style.Render("\t\tNodes\t\t"), style.Render("\t\tLinks\t\t")}
		info_string := fmt.Sprintf("ID: %s\n\nName: %s\n\nDescription: %s\n\nCreated at: %s\n\nTopology type: %s",
			sliceById.Slice.SliceID, sliceById.Slice.Name, sliceById.Slice.Description, sliceById.Slice.CreatedAt.Format("2006-01-02 15:04:05"), sliceById.Slice.TopologyType)
		nodes := table.NewWriter()
		nodes.AppendHeader(table.Row{"ID", "Name", "Image", "CPU", "Memory", "Storage"})
		for _, v := range sliceById.Slice.Topology.Nodes {
			nodes.AppendRow(table.Row{v.NodeID, v.Name, v.Image, strconv.Itoa(v.Flavor.CPU), strconv.FormatFloat(float64(v.Flavor.Memory), 'f', 1, 32), strconv.FormatFloat(float64(v.Flavor.Storage), 'f', 1, 32)})
		}
		links := table.NewWriter()
		links.AppendHeader(table.Row{"ID", "Source", "Target"})
		for _, v := range sliceById.Slice.Topology.Links {
			links.AppendRow(table.Row{v.LinkID, v.Source, v.Target})
		}
		tabContent := []string{style.Render(info_string), style.Render(nodes.Render()), style.Render(links.Render())}
		m := model{Tabs: tabs, TabContent: tabContent}
		if _, err := tea.NewProgram(m).Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
