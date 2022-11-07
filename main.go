package main

import (
	"database/sql"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lylecantcode/gym_tui/history"
	"github.com/lylecantcode/gym_tui/workout"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

// create a basic database table if it doesn't exist.
var db *sql.DB

func dbStartUp() *sql.DB {
	database, err :=
		sql.Open("sqlite3", "./gym_routine.db")
	if err != nil {
		log.Fatal("error opening DB: " + err.Error())
	}
	statement, err :=
		database.Prepare("CREATE TABLE IF NOT EXISTS gym_routine (id INTEGER PRIMARY KEY, exercise VARCHAR NOT NULL, weight INTEGER DEFAULT 0, reps INTEGER DEFAULT 0, date TEXT DEFAULT CURRENT_DATE)")
	if err != nil {
		log.Fatal("error with creation statement: " + err.Error())
	}
	_, err = statement.Exec()
	if err != nil {
		log.Fatal("error creating table: " + err.Error())
	}
	return database
}

func main() {
	db = dbStartUp()
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		log.Fatal("Gym TUI encountered the following error: %v", err)
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
		s += fmt.Sprintf("%s %s \n", cursor, m.options[i])
	}

	// The footer
	s += `
Navigate using the arrow keys and use enter to select.
Press Q or Ctrl+C to quit.`
	return s

}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	// Is the input a key press?
	case tea.KeyMsg:

		// What key was pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
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
				return workout.StartWorkout(db, m), cmd
			case 1:
				return history.GetHistory(db, m), cmd
			case 2:
				return history.GetBests(db, m), cmd
			}
		}
	}
	return m, cmd
}