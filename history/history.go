package history

import (
	"database/sql"
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/mattn/go-sqlite3"
	wiki "github.com/trietmn/go-wiki"
	"log"
	"strconv"
	"strings"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table      table.Model
	rows       int
	searchTerm map[string]string
	mainMenu   tea.Model
}

type workouts struct {
	id       int
	exercise string
	weight   int
	rep      int
	date     string
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q":
			return m.mainMenu, cmd
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			c := make(chan string)
			go m.searchWiki(c)
			results := <-c
			return m, tea.Batch(
				tea.Printf("\n%s\n", results),
			)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.rows == 0 {
		return "No previous results found."
	}
	footer := fmt.Sprintf(`
Feel free to scroll (there %d entries total).
Select an exercise to learn more about it.
Ctrl+C to quit or Q to go back to main menu.
`, m.rows)
	return baseStyle.Render(m.table.View()) + footer
}

// if running this as a channel, should it be launched near the started and just run in background?
// this is still causing lag.
func (m *model) searchWiki(c chan string) {
	var results string
	// results, err := wiki.Summary(m.searchTerm[m.table.SelectedRow()[1]], 1, -1, false, true)
	page, err := wiki.GetPage(m.searchTerm[m.table.SelectedRow()[1]], -1, false, true)
	if err != nil {
		results = fmt.Sprintf("%s", err)
	} else {
		content, err := page.GetContent()
		if err != nil {
			results = fmt.Sprintf("%s", err)
		}
		results, _, _ = strings.Cut(content, ".")
		// this is still being cropped by the TUI :(
		// think because of screen width.
	}
	c <- results
}

func createTableTUI(exercises []workouts, mainMenu tea.Model) tea.Model {
	rowCount := len(exercises)

	rows := []table.Row{}
	for i := 0; i < rowCount; i++ {
		data := table.Row{
			strconv.Itoa(i + 1), //exercises[i].id),
			exercises[i].exercise,
			strconv.Itoa(exercises[i].weight),
			strconv.Itoa(exercises[i].rep),
			exercises[i].date,
		}
		rows = append(rows, data)
	}

	columns := []table.Column{
		{Title: "id", Width: 4},
		{Title: "Exercise", Width: 15},
		{Title: "Weight", Width: 6},
		{Title: "Rep", Width: 5},
		{Title: "Date", Width: 12},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	searchTerms := make(map[string]string)

	m := model{t, rowCount, searchTerms, mainMenu}

	m.searchTerm["Squats"] = "Squat (exercise)"
	m.searchTerm["Deadlifts"] = "Deadlift"
	m.searchTerm["Rows"] = "Bent-over row"
	m.searchTerm["Overhead Press"] = "Overhead Press"
	m.searchTerm["Bench Press"] = "Bench Press"
	m.searchTerm["Tricep Dips"] = "Dip (exercise)"
	m.searchTerm["Dips"] = "Dip (exercise)"
	m.searchTerm["Pullups"] = "Pull-up (exercise)"
	m.searchTerm["Bicep Curls"] = "Biceps curl"

	return m
}

func GetBests(db *sql.DB, mainMenu tea.Model) tea.Model {
	var exercises []workouts
	dbRows, err :=
		db.Query("SELECT exercise, MAX(weight), reps, date FROM gym_routine GROUP BY exercise ORDER BY weight,reps DESC LIMIT 50")
	if err != nil {
		log.Fatal("Select Bests query error: " + err.Error())
	}
	for dbRows.Next() {
		workout := workouts{}
		dbRows.Scan(&workout.exercise, &workout.weight, &workout.rep, &workout.date)
		exercises = append(exercises, workout)
	}
	return createTableTUI(exercises, mainMenu)
}

func GetHistory(db *sql.DB, mainMenu tea.Model) tea.Model {
	var exercises []workouts
	dbRows, err :=
		db.Query("SELECT id, exercise, weight, reps, date FROM gym_routine WHERE weight != 0 OR reps != 0 ORDER BY id ASC LIMIT 50")
	if err != nil {
		log.Fatal("Select History error: " + err.Error())
	}
	for dbRows.Next() {
		workout := workouts{}
		dbRows.Scan(&workout.id, &workout.exercise, &workout.weight, &workout.rep, &workout.date)
		exercises = append(exercises, workout)
	}
	return createTableTUI(exercises, mainMenu)
}
