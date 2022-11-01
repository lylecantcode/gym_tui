package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lylecantcode/gym_tui/workout"
)

func main() {
	workout.StartWorkout() }
// 	p := tea.NewProgram(initialModel())
// 	if err := p.Start(); err != nil {
// 		fmt.Printf("Gym TUI encountered the following error: %v", err)
// 		os.Exit(1)
// 	}
// }



// func (m model) View() string {
// 	options := []string{"workout", "history", "personal bests"}
// 	// The header
// 	s := "What would you like to access?\n\n"

// 	// Iterate over our choices.
// 	for i := 0; i < len(m.choices); i++ {

// 		// Is the cursor pointing at this choice?
// 		cursor := " " // no cursor
// 		if m.cursor == i {
// 			cursor = ">" // cursor!
// 		}

// 		// Is this choice selected?
// 		checked := " " // not selected
// 		if _, ok := m.selected[i]; ok {
// 			checked = "x" // selected!
// 		}

// 		// Render the rows
// 		s += fmt.Sprintf("%s %s", cursor, options[i])//, checked, m.choices[i], m.weights[m.choices[i]])
// 	}
// 	// Render the addingNew row, both in "adding" state and in default state.
// 	s += fmt.Sprintf("%s\n\n", m.addingNew.View())

// 	// The footer
// 	s += `Navigate using the arrow keys and use enter to select.
// Press Delete to delete an exercise. 
// Press Q or Ctrl+C to quit.`
// 	return s
// }

// func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	var cmd tea.Cmd

// 	switch msg := msg.(type) {

// 	// Is the input a key press?
// 	case tea.KeyMsg:

// 		// What key was pressed?
// 		switch msg.String() {

// 		// These keys should exit the program.
// 		case "ctrl+c", "q":
// 			if msg.String() == "ctrl+c" || !m.typing {
// 				return m, tea.Quit
// 			}

// 		// The "up" key moves the cursor up
// 		case "up", "shift+tab":
// 			if m.cursor > 0 {
// 				m.cursor--
// 			}

// 		// The "down" key moves the cursor down
// 		case "down", "tab":
// 			if m.cursor < len(m.choices)-1 {
// 				m.cursor++
// 			}

// 		// The "enter" key allows for a set to be inputted.
// 		case "enter":
// 			if m.cursor < len(m.choices) {
// 				// unfocuses if still focused.
// 				m.addingNew.Blur()
// 				// check if the weight is currently selected, only one should be selected at a time.
// 				_, ok := m.selected[m.cursor]
// 				if ok {
// 					// if it is already selected then submit the value.
// 					m.addNewSet()
// 					// then delete the selection
// 					delete(m.selected, m.cursor)
// 				} else {
// 					// otherwise, select the item and focus on the text field.
// 					// use of struct{}{} minimises memory cost.
// 					m.selected[m.cursor] = struct{}{}
// 					m.addingNew.Focus()
// 					m.typing = true
// 				}
// 			}
// 		// allows the deletion of an exercise, feature will likely be removed or changed.
// 		case "delete", "backspace":
// 			// stops the backspace key from deleting an exercise if fixing a typo!
// 			if m.cursor < len(m.choices) && !m.typing {
// 				deleteExerciseFromDB(m.choices[m.cursor])
// 				m.choices = deleteChoice(m.choices, m.cursor)
// 			}
// 		}
// 	}
// 	// used to drive the addNewSet functionality.
// 	cmd = m.updateInputs(msg)
// 	return m, cmd
// }
