package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/tavp-stack/tavpbox/internal/config"
	"github.com/tavp-stack/tavpbox/internal/lxd"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "First-time setup wizard",
	RunE:  runInit,
}

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).MarginBottom(1)
	docStyle   = lipgloss.NewStyle().Margin(1, 2)
)

func runInit(cmd *cobra.Command, args []string) error {
	if err := checkPrerequisites(); err != nil {
		return err
	}

	m := newInitModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return err
	}

	model := result.(initModel)
	if model.cancelled {
		fmt.Println("Setup cancelled.")
		return nil
	}

	cfg := &config.GlobalConfig{
		DomainSuffix:  model.domain,
		DefaultRAM:    model.ram,
		DefaultCPU:    1,
		DefaultDistro: model.selectedDistro,
		Network: config.NetworkConfig{
			Bridge: "lxdbr0",
			Subnet: "10.0.3.0/24",
			DNS:    "10.0.3.1",
		},
	}

	home := config.HomeDir()
	for _, dir := range []string{"", "boxes", "plugins", "snapshots"} {
		os.MkdirAll(filepath.Join(home, dir), 0755)
	}

	if err := config.SaveGlobal(cfg); err != nil {
		return err
	}

	fmt.Println("\nConfiguring networking...")
	exec.Command("lxd", "init", "--auto").Run()

	fmt.Printf("Downloading %s image (first time)...\n", model.selectedDistro)
	client := lxd.New()
	client.Create("tavpbox-image-temp", model.selectedDistro, "64MB", 1)
	client.Delete("tavpbox-image-temp")

	fmt.Println(`
╔══════════════════════════════════════════════════════╗
║          ✓  TAVPBox initialized successfully         ║
╠══════════════════════════════════════════════════════╣
║                                                      ║
║  Next steps:                                         ║
║    tavpbox create          Create a dev box          ║
║    tavpbox list            List all boxes            ║
║    tavpbox --help          See all commands           ║
║                                                      ║
╚══════════════════════════════════════════════════════╝`)

	return nil
}

func checkPrerequisites() error {
	if _, err := exec.LookPath("lxc"); err != nil {
		if _, err := os.Stat("/snap/bin/lxc"); err != nil {
			return fmt.Errorf("required tool 'lxc' not found. Run: tavpbox setup")
		}
	}
	return nil
}

// ─── TUI Model ───

type initModel struct {
	step           int
	selectedDistro string
	domain         string
	ram            string
	cancelled      bool
	distros        []string
	cursor         int
}

func newInitModel() initModel {
	return initModel{
		step: 0,
		distros: []string{
			"ubuntu/24.04",
			"alpine/3.20",
			"debian/12",
			"fedora/40",
			"archlinux",
		},
		domain: "tavp.local",
		ram:    "512MB",
	}
}

func (m initModel) Init() tea.Cmd { return nil }

func (m initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.cancelled = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.distros)-1 {
				m.cursor++
			}

		case "enter":
			switch m.step {
			case 0:
				m.selectedDistro = m.distros[m.cursor]
				m.step = 1
				return m, nil
			case 1:
				m.step = 2
				return m, nil
			case 2:
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m initModel) View() string {
	s := titleStyle.Render("⚡ TAVPBox — Initial Setup")
	s += "\n\n"

	switch m.step {
	case 0:
		s += "Select base distro for your boxes:\n\n"
		for i, d := range m.distros {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}
			s += fmt.Sprintf("  %s %s\n", cursor, d)
		}
		s += "\n  ↑↓ navigate · enter select"

	case 1:
		s += "Domain suffix for auto-generated URLs:\n"
		s += fmt.Sprintf("  Boxes will be accessible at: <name>.%s\n\n", m.domain)
		s += fmt.Sprintf("  [Enter to keep: %s]\n", m.domain)

	case 2:
		s += "Default RAM limit per box:\n"
		s += "  (can be overridden per box)\n\n"
		s += fmt.Sprintf("  [Enter to keep: %s]\n", m.ram)
		s += "\n  enter to finish setup"
	}

	return docStyle.Render(s)
}
