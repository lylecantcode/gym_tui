package main

import (
	"database/sql"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
	weights   map[string][]string // weights on the to-do list
	cursor    int                 // which list item our cursor is pointing at
	selected  map[int]struct{}    // which weights are selected
	addingNew textinput.Model     // used to add new values
	typing    bool                // used to stop delete or quit triggering accidentally
}

func initialModel() model {
	items := model{selected: make(map[int]struct{})}
	values := getValuesFromDB()
	insertModel := textinput.New()
	insertModel.Placeholder = "After selecting, type Weight (kg) x Reps here"
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
		items.addToDB(setSlice, items.choices[i])
	}

	return items
}

func (m model) Init() tea.Cmd {
	return nil
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
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The "enter" key allows for a set to be inputted.
		case "enter":
			if m.cursor < len(m.choices) {
				// unfocuses if still focused.
				m.addingNew.Blur()
				// check if the weight is currently selected, only one should be selected at a time.
				_, ok := m.selected[m.cursor]
				if ok {
					// if it is already selected then submit the value.
					m.addNewSet()
					// then delete the selection
					delete(m.selected, m.cursor)
				} else {
					// otherwise, select the item and focus on the text field.
					// use of struct{}{} minimises memory cost.
					m.selected[m.cursor] = struct{}{}
					m.addingNew.Focus()
					m.typing = true
				}
			}
		// allows the deletion of an exercise, feature will likely be removed or changed.
		case "delete", "backspace":
			// stops the backspace key from deleting an exercise if fixing a typo!
			if m.cursor < len(m.choices) && !m.typing {
				m.deleteExerciseFromDB(m.choices[m.cursor])
				m.choices = deleteChoice(m.choices, m.cursor)
			}
		}
	}
	// used to drive the addNewSet functionality.
	cmd = m.updateInputs(msg)
	return m, cmd
}

// handles typing once focused.
func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.addingNew, cmd = m.addingNew.Update(msg)
	return cmd
}

func (m model) View() string {
	// The header
	s := "What exercise did you do?\n\n"

	// Iterate over our choices.
	for i := 0; i < len(m.choices); i++ {

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

		// Render the rows
		s += fmt.Sprintf("%s [%s] %s %s\n", cursor, checked, m.choices[i], m.weights[m.choices[i]])
	}
	// Render the addingNew row, both in "adding" state and in default state.
	s += fmt.Sprintf("%s\n\n", m.addingNew.View())

	// The footer
	s += `Navigate using the arrow keys and use enter to select.
Press Delete to delete an exercise. 
Press Q or Ctrl+C to quit.`
	return s
}

func (m *model) addNewSet() {
	// obtains the value from the "add new" line
	newItem := m.addingNew.Value()
	pos := m.choices[m.cursor]
	// checks for existing map and if none, initiates map
	if m.weights[pos] == nil {
		m.weights = make(map[string][]string)
	}
	// unfocus/deselect the "add new" line
	m.addingNew.Blur()
	// reset the "add new" line back to default values
	m.addingNew.Reset()
	// check for groupings of numbers and ignores any other values.
	re := regexp.MustCompile(`\d+`)
	set := re.FindAllString(newItem, -1)
	// if 2 groups of numbers, i.e. weights and sets
	if len(set) == 2 {
		// add new set to exercise in fixed format
		m.weights[pos] = append(m.weights[pos], "("+set[0]+"kg x "+set[1]+")")
		// add new set to DB
		m.addToDB(set, pos)
	}
	m.typing = false
}

// delete an exercise from the list by creating a new slice of all other items.
func deleteChoice(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
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

func (m model) addToDB(set []string, exercise string) {
	if len(set) == 2 {
		// based on the requested format:
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

func (m model) deleteExerciseFromDB(exercise string) {
	statement, _ :=
		db.Prepare("DELETE FROM gym_routine WHERE exercise = ?")
	statement.Exec(exercise)
}

func getValuesFromDB() []string {
	var exercise string
	exerciseArray := []string{}

	rows, _ :=
		db.Query("SELECT exercise FROM gym_routine")
	for rows.Next() {
		rows.Scan(&exercise)
		exerciseArray = append(exerciseArray, exercise)
	}
	return exerciseArray
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Gym TUI encountered the following error: %v", err)
		os.Exit(1)
	}
}