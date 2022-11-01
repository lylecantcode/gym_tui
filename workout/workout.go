package workout

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
	exercises := model{selected: make(map[int]struct{})}
	values := exercises.getValuesFromDB()
	insertModel := textinput.New()
	insertModel.Placeholder = "After selecting, type Weight (kg) x Reps here"
	insertModel.Prompt = ""    // we do not want an additional prompt for the "add new item"
	insertModel.CharLimit = 20 // character limit will limit potentially problematic entries
	insertModel.Width = 20     // this limits the field of view when typing to X characters long

	if len(values) == 0 {
		exercises.choices = []string{"Bench Press", "Squats", "Pullups", "Dips", "Tricep Dips", "Bicep Curls", "Overhead Press", "Deadlifts", "Rows"}
	} else {
		exercises.choices = values
	}
	exercises.addingNew = insertModel

	// TODO: need to write a way to handle empty weights/reps to provide the original empty list
	var setSlice []string
	for i := 0; i < len(exercises.choices); i++ {
		exercises.addToDB(setSlice, exercises.choices[i])
	}

	return exercises
}

func (m model) Init() tea.Cmd {
	return nil
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
		case "up", "shift+tab" && !m.typing:
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" key moves the cursor down
		case "down", "tab":
			if m.cursor < len(m.choices)-1 && !m.typing {
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
				deleteExerciseFromDB(m.choices[m.cursor])
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

func (m *model) addNewSet() {
	// obtains the value from the "add new" line
	ex := m.choices[m.cursor]
	// create channel, c
	// rec := make(chan []string)
	c := make(chan []string)
	go m.regExFiltering(c)
	go m.setMapping(ex, c)
	set := <-c

	if len(set) == 2 {
		c <- set
		m.addToDB(set, ex)
	}
	// checks for existing map and if none, initiates map
	// unfocus/deselect the "add new" line
	m.addingNew.Blur()
	// reset the "add new" line back to default values
	m.addingNew.Reset()

	m.typing = false
}

func (m *model) regExFiltering(c chan []string) {
	// get value from the user inputted field.
	newSet := m.addingNew.Value()
	// check for groupings of numbers and ignores any other values.
	re := regexp.MustCompile(`\d+`)
	set := re.FindAllString(newSet, -1)
	c <- set
}

func (m *model) setMapping(ex string, c chan []string) {
	if m.weights/*[ex]*/ == nil { // may need to keep if allowing for adding new exercises
		m.weights = make(map[string][]string)
	}
	set := <-c
	m.weights[ex] = append(m.weights[ex], "("+set[0]+"kg x "+set[1]+")")
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
		database.Prepare("CREATE TABLE IF NOT EXISTS gym_routine (id INTEGER PRIMARY KEY, exercise VARCHAR NOT NULL, weight INTEGER, reps INTEGER, date TEXT)")
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
			db.Prepare("INSERT INTO gym_routine (exercise, weight, reps) VALUES (?, ?, ?)")
		statement.Exec(exercise, 0, 0)
	}
}

func deleteExerciseFromDB(exercise string) {
	statement, _ :=
		db.Prepare("DELETE FROM gym_routine WHERE exercise = ?")
	statement.Exec(exercise)
}

func (m model) getValuesFromDB() []string {
	var exercise string
	var weight, rep string
	exerciseArray := []string{}

	rows, _ :=
		db.Query("SELECT exercise, weight, reps FROM gym_routine WHERE id IN (SELECT DISTINCT exercise FROM gym_routine ORDER BY id DESC)")
	for rows.Next() {
		rows.Scan(&exercise, &weight, &rep)
		exerciseArray = append(exerciseArray, exercise)
		if m.weights[exercise] == nil {
			m.weights = make(map[string][]string)
		}
		m.weights[exercise] = append(m.weights[exercise], "("+weight+"kg x "+rep+")")

	}
	return exerciseArray
}

func StartWorkout() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Gym TUI encountered the following error: %v", err)
		os.Exit(1)
	}
}
