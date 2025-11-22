package ui

import (
	"fmt"

	"github.com/bastiblast/boiler-deploy/internal/config"
	"github.com/bastiblast/boiler-deploy/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
)

// ConfigSelector allows user to select an environment to configure
type ConfigSelector struct {
	environments []string
	cursor       int
	configMgr    *config.Manager
}

// NewConfigSelector creates a new configuration selector
func NewConfigSelector() ConfigSelector {
	stor := storage.NewStorage(".")
	envList, _ := stor.ListEnvironments()
	configMgr := config.NewManager("inventory")
	
	return ConfigSelector{
		environments: envList,
		cursor:       0,
		configMgr:    configMgr,
	}
}

func (s ConfigSelector) Init() tea.Cmd {
	return nil
}

func (s ConfigSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return s, tea.Quit
			
		case "esc":
			return NewMainMenu(), nil
			
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
			
		case "down", "j":
			if s.cursor < len(s.environments)-1 {
				s.cursor++
			}
			
		case "enter":
			if len(s.environments) == 0 {
				return NewMainMenu(), nil
			}
			// Open configuration form for selected environment
			form, err := NewConfigForm(s.environments[s.cursor], s.configMgr)
			if err != nil {
				// TODO: Handle error
				return NewMainMenu(), nil
			}
			return form, nil
		}
	}
	
	return s, nil
}

func (s ConfigSelector) View() string {
	var view string
	view += titleStyle.Render("⚙️  Configuration Options") + "\n\n"
	
	if len(s.environments) == 0 {
		view += errorStyle.Render("No environments found. Create one first.") + "\n\n"
		view += helpStyle.Render("Press esc to go back")
		return view
	}
	
	view += infoStyle.Render("Select environment to configure:") + "\n\n"
	
	for i, env := range s.environments {
		cursor := "  "
		style := normalItemStyle
		if s.cursor == i {
			cursor = "▶ "
			style = selectedItemStyle
		}
		
		// Load and display current config
		cfg, err := s.configMgr.Load(env)
		if err != nil {
			cfg = config.DefaultConfig()
		}
		
		envInfo := fmt.Sprintf("%s (Strategy: %s, Tags: %d prov/%d deploy)",
			env,
			cfg.ProvisioningStrategy,
			len(cfg.ProvisioningTags),
			len(cfg.DeploymentTags),
		)
		
		view += style.Render(cursor+envInfo) + "\n"
	}
	
	view += "\n" + helpStyle.Render("[↑↓/jk] Navigate • [Enter] Configure • [esc] Back • [q] Quit")
	
	return view
}
