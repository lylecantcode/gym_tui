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
)

var db *sql.DB
var Quitting bool

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
	// values := exercises.getValuesFromDB()
	insertModel := textinput.New()
	insertModel.Placeholder = "After selecting, type Weight (kg) x Reps here"
	insertModel.Prompt = ""    // we do not want an additional prompt for the "add new item"
	insertModel.CharLimit = 20 // character limit will limit potentially problematic entries
	insertModel.Width = 20     // this limits the field of view when typing to X characters long

	exercises.choices = []string{"Bench Press", "Squats", "Pullups", "Dips", "Tricep Dips", "Bicep Curls", "Overhead Press", "Deadlifts", "Rows"}
	exercises.addToDB([]string{}, exercises.choices...)

	exercises.addingNew = insertModel

	return exercises
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) View() string {
	// The header
	s := "\n\nWhat exercise did you do?\n\n"

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
Press Q to go back to main menu or Ctrl+C to quit..`
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
		case "q":
			if !m.typing {
				return m, tea.Quit
			}

		case "ctrl+c":
			Quitting = true
			return m, tea.Quit

		// The "up" key moves the cursor up
		case "up", "shift+tab":
			if m.cursor > 0 && !m.typing {
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
	if m.weights /*[ex]*/ == nil { // may need to keep if allowing for adding new exercises
		m.weights = make(map[string][]string)
	}
	set := <-c
	m.weights[ex] = append(m.weights[ex], "("+set[0]+"kg x "+set[1]+")")
}

// delete an exercise from the list by creating a new slice of all other items.
func deleteChoice(s []string, index int) []string {
	// elipses separates the array into separate values to append them to the new slice
	return append(s[:index], s[index+1:]...)
}

func (m *model) addToDB(set []string, exercise ...string) {
	var counter int
	if len(set) == 2 {
		// based on the requested format:
		weight, _ := strconv.Atoi(set[0])
		reps, _ := strconv.Atoi(set[1])
		statement, _ :=
			db.Prepare("INSERT INTO gym_routine (exercise, weight, reps) VALUES (?, ?, ?)")
		statement.Exec(exercise[0], weight, reps)
	} else if db.QueryRow("SELECT count(*) FROM gym_routine").Scan(&counter); counter == 0 {
		for _, ex := range exercise {
			statement, _ :=
				db.Prepare("INSERT INTO gym_routine (exercise) VALUES (?)")
			statement.Exec(ex)
		}
	}
}

func StartWorkout(database *sql.DB) bool {
	db = database
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Gym TUI encountered the following error: %v", err)
		os.Exit(1)
	}
	p.Kill()	// doesn't seem to help :(
	if Quitting {
		return true
	}
	return false
}
