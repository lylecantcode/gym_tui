package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lylecantcode/gym_tui/history"
	"github.com/lylecantcode/gym_tui/workout"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

// create a basic database table if it doesn't exist.
var db *sql.DB

func dbStartUp() *sql.DB {
	database, _ :=
		sql.Open("sqlite3", "./gym_routine.db")
	statement, _ :=
		database.Prepare("CREATE TABLE IF NOT EXISTS gym_routine (id INTEGER PRIMARY KEY, exercise VARCHAR NOT NULL, weight INTEGER, reps INTEGER, date TEXT)")
	statement.Exec()

	return database
}

func main() {
	db = dbStartUp()
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Gym TUI encountered the following error: %v", err)
		os.Exit(1)
	}
}

type model struct {
	options []string
	cursor  int // which list item our cursor is pointing at
	hidden  bool
}

func initialModel() model {
	menu := model{}
	menu.options = []string{"Begin workout!", "Exercise History", "Personal Bests"}
	return menu
}

// try without this first
func (m model) Init() tea.Cmd {
	return nil
}

func (m model) View() string {
	// The header
	s := "What would you like to access?\n\n"

	for i := 0; i < len(m.options); i++ {

		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Render the rows
		s += fmt.Sprintf("%s %s \n", cursor, m.options[i]) //, checked, m.choices[i], m.weights[m.choices[i]])
	}

	// The footer
	s += "\nNavigate using the arrow keys and use enter to select.\nPress Q or Ctrl+C to quit."
	if m.hidden {
		s = ""
	}
	return s

}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	p := tea.NewProgram(initialModel())
	var cmd tea.Cmd

	switch msg := msg.(type) {

	// Is the input a key press?
	case tea.KeyMsg:

		// What key was pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			p.Kill()
			return m, tea.Quit

		// The "up" key moves the cursor up
		case "up", "shift+tab":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" key moves the cursor down
		case "down", "tab":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}

		// The "enter" key allows for a set to be inputted.
		case "enter":
			switch m.cursor {
			case 0:
				workout.StartWorkout(db)
				p.StartReturningModel()
			case 1:
				history.GetHistory(db)
				p.StartReturningModel()
			case 2:
				fmt.Println("Still a work in progress :)\n\n")
				p.StartReturningModel()
			}

		}
	}
	return m, cmd
}
