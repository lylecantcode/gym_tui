package main

import (
	"database/sql"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
	// "github.com/charmbracelet/lipgloss" // package for the future, for style changes.
	_ "github.com/mattn/go-sqlite3"
	"os"
	"regexp"
	"strconv"
	"time"
)

var db = dbStartUp()
var TODAY string = time.Now().Format("02-01-2006")

type model struct {
	choices   []string
	weights   map[string][]string // items on the to-do list
	cursor    int                 // which list item our cursor is pointing at
	selected  map[int]struct{}    // which items are selected
	addingNew textinput.Model     // used to add new values
	typing    bool                // used to stop delete or quit triggering accidentally
}

func initialModel() model {
	items := model{selected: make(map[int]struct{})}
	values := getValues()
	insertModel := textinput.New()
	insertModel.Placeholder = "Type Weight (kg) x Reps here"
	insertModel.Prompt = ""    // we do not want an additional prompt for the "add new item"
	insertModel.CharLimit = 20 // character limit will limit potentially problematic entries
	insertModel.Width = 20     // this limits the field of view when typing to X characters long

	if len(values) == 0 {
		items.choices = []string{"Bench Press", "Squats", "Pullups", "Dips", "Tricep Dips", "Bicep Curls", "Overhead Press", "Deadlifts", "Rows"}
	} else {
		items.choices = values
	}
	items.addingNew = insertModel

	// TODO: need to write a way to handle empty weights/reps to provide the original empty list
	var setSlice []string
	for i := 0; i < len(items.choices); i++ {
		items.addItems(setSlice, items.choices[i])
	}

	return items
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			if msg.String() == "ctrl+c" || !m.typing {
				return m, tea.Quit
			} 

		// The "up" key moves the cursor up
		case "up", "shift+tab":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" key moves the cursor down
		case "down", "tab":
			if m.cursor < len(m.choices) { //-1 {
				m.cursor++
			}

		// The "enter" key toggles the selected state for the item that the cursor is pointing at.
		// if pointing at addNew, this then allows a new value to be typed in.
		// de-focus to stop input and update DB selection boolean.
		case "enter":
			if m.cursor < len(m.choices) {
				m.addingNew.Blur()
				// m.checkItem(m.cursor)
				_, ok := m.selected[m.cursor]
				if ok {
					m.addNewItem()
					delete(m.selected, m.cursor)
				} else {
					m.selected[m.cursor] = struct{}{}
					m.addingNew.Focus()
					m.typing = true
				}
			}
		case "delete", "backspace":
			if m.cursor < len(m.choices) && !m.typing {
				delete(m.selected, m.cursor) // otherwise [x] will be retained visually.
				m.deleteItems(m.choices[m.cursor])
				m.choices = deleteChoice(m.choices, m.cursor)
			}
		}
	}
	cmd = m.updateInputs(msg) // used to drive the addNewItem functionality.
	return m, cmd
}

// Change to remove last rep.
func deleteChoice(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

// insert new item to choices, from the addingNew model
// unfocus to stop further input
// reset addingNew back to default state
// delete selection of addNew, reselect newly inserted item
// add new item to DB
func (m *model) addNewItem() {
	newItem := m.addingNew.Value()
	pos := m.choices[m.cursor]
	if m.weights[pos] == nil {
		m.weights = make(map[string][]string)
	}
	m.addingNew.Blur()
	m.addingNew.Reset()
	re := regexp.MustCompile(`\d+`)
	set := re.FindAllString(newItem, -1)
	if len(set) == 2 {
		m.weights[pos] = append(m.weights[pos], "("+set[0]+"kg x "+set[1]+")")
		m.addItems(set, pos)
	}
	m.typing = false
}

// handles typing once focused.
func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.addingNew, cmd = m.addingNew.Update(msg)
	return cmd
}

func (m *model) dupeCheck() bool {
	newItem := m.addingNew.Value()
	if newItem == " " || newItem == "" {
		return true
	}
	for i, v := range m.choices {
		if strings.ToLower(v) == strings.ToLower(newItem) {
			m.selected[i] = struct{}{}
			return true
		}
	}
	return false
}

func (m model) View() string {
	// The header
	s := "What exercise did you do? (navigate with arrow keys and use enter to select)\n\n"

	// Iterate over our choices
	for i := 0; i <= len(m.choices); i++ {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		if i < len(m.choices) {
			s += fmt.Sprintf("%s [%s] %s %s\n", cursor, checked, m.choices[i], m.weights[m.choices[i]])
		} else {
			// Render the addingNew row, both in "adding" state and in default state.
			s += fmt.Sprintf("%s\n", m.addingNew.View())
		}
	}
	// The footer
	s += "\nPress Delete to delete an item. \nPress Q or Ctrl+C to quit.\n"
	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

// create a basic database table if it doesn't exist.
func dbStartUp() *sql.DB {
	database, _ :=
		sql.Open("sqlite3", "./gym_routine.db")
	statement, _ :=
		database.Prepare("CREATE TABLE IF NOT EXISTS gym_routine (id INTEGER PRIMARY KEY, exercise VARCHAR NOT NULL UNIQUE, weight INTEGER, reps INTEGER, date TEXT NOT NULL)")
	statement.Exec()

	return database
}

func (m model) addItems(set []string, exercise string) {
	if len(set) == 2 {
		weight, _ := strconv.Atoi(set[0])
		reps, _ := strconv.Atoi(set[1])
		statement, _ :=
			db.Prepare("INSERT INTO gym_routine (exercise, date, weight, reps) VALUES (?, ?, ?, ?)")
		statement.Exec(exercise, TODAY, weight, reps)
	} else {
		statement, _ :=
			db.Prepare("INSERT INTO gym_routine (exercise, date) VALUES (?, ?)")
		statement.Exec(exercise, TODAY)
	}
}

func (m model) deleteItems(item string) {
	statement, _ :=
		db.Prepare("DELETE FROM gym_routine WHERE exercise = ?")
	statement.Exec(item)
}

// func (m model) checkItem(i int) {
// 	statement, _ :=
// 		db.Prepare("UPDATE gym_routine SET checked = NOT checked WHERE exercise = ?")
// 	statement.Exec(m.choices[i])
// }

// prepopulate the choices if there is an existing database, rather than using default values
// i.e. memory between sessions.
func getValues() []string {
	var item string
	itemArray := []string{}

	rows, _ :=
		db.Query("SELECT exercise FROM gym_routine")
	for rows.Next() {
		rows.Scan(&item)
		itemArray = append(itemArray, item)
	}
	return itemArray
}
