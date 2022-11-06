package workout

import (
	"database/sql"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"regexp"
	"strconv"
)

type model struct {
	choices   []string
	weights   map[string][]string // weights on the to-do list
	cursor    int                 // which list item our cursor is pointing at
	selected  map[int]struct{}    // which weights are selected
	addingNew textinput.Model     // used to add new values
	typing    bool                // used to stop delete or quit triggering accidentally
	mainMenu tea.Model
	db *sql.DB
}

func initialModel(database *sql.DB, mainMenu tea.Model) model {
	exercises := model{
		selected: make(map[int]struct{}),
		mainMenu: mainMenu,
		db: database,
}
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
				return m.mainMenu, cmd
			}

		case "ctrl+c":
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
	// used to drive the addNewSet functionality
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
	// using regex to separate out the 2 values, for the weight and reps
	// this allows any separator to be used for the 2 values
	set := m.regExFiltering()	
	if len(set) == 2 {
		m.setMapping(set)
		// using a go routine here will minimise any delay in usage of GUI
		go m.addToDB(set, m.choices[m.cursor])
	}
	// checks for existing map and if none, initiates map
	// unfocus/deselect the "add new" line
	m.addingNew.Blur()
	// reset the "add new" line back to default values
	m.addingNew.Reset()

	m.typing = false
}

func (m *model) regExFiltering() []string {
	// get value from the user inputted field.
	newSet := m.addingNew.Value()
	// check for groupings of numbers and ignores any other values.
	re := regexp.MustCompile(`\d+`)
	return re.FindAllString(newSet, -1)

}

func (m *model) setMapping(set []string) {
	if m.weights == nil { // may need to keep if allowing for adding new exercises
		m.weights = make(map[string][]string)
	}
	ex := m.choices[m.cursor]
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
		weight, err := strconv.Atoi(set[0])
		if err != nil {
			log.Fatal("something went wrong with regex pattern, invalid value: " + err.Error())
		}
		reps, err := strconv.Atoi(set[1])
		if err != nil {
			log.Fatal("something went wrong with regex pattern, invalid value: " + err.Error())
		}
		statement, err :=
			m.db.Prepare("INSERT INTO gym_routine (exercise, weight, reps) VALUES (?, ?, ?)")
		if err != nil {
			log.Fatal("db insertion statement error: " + err.Error())
		}
		_, err = statement.Exec(exercise[0], weight, reps)
		if err != nil {
			log.Fatal("db insertion error: " + err.Error())
		}
	} else if m.db.QueryRow("SELECT count(*) FROM gym_routine").Scan(&counter); counter == 0 {
		for _, ex := range exercise {

			statement, err :=
				m.db.Prepare("INSERT INTO gym_routine (exercise) VALUES (?)")
			if err != nil {
				log.Fatal("db insertion statement error: " + err.Error())
			}
			_, err = statement.Exec(ex)
			if err != nil {
				log.Fatal("db insertion error: " + err.Error())
			}
		}
	}
	// closes the go routine after process is complete.
	return
}

func StartWorkout(database *sql.DB, mainMenu tea.Model) tea.Model {
	return initialModel(database, mainMenu)
}
