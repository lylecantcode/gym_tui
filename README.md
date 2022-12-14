# Gym Terminal User Interface

Used to track workouts from the terminal in an easy and intuitive manner.
Saves the workout in an SQLite3 database.

## Table of Contents

- [Example of use](#example-of-use)
- [Background](#background)
- [Future Plans](#future-plans)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)
- [Known Bugs](#known-bugs)

## Example of use
I just did a set of squats and I want to add my set of 50 kg and 10 reps.
Assuming the TUI is already running:

1) From the main menu, select workouts.
    * Do this via the arrow keys and enter.
2) Navigate using the arrow keys to the exercise (squats).
3) Click enter.
    * This will allow you to enter text into the "add new" line.
4) Enter your weight and reps, (50 x 10).
5) Click enter to submit.
	* This will now display next to the exercise.

```sh
What exercise did you do?
~~
  [ ] Bench Press []
> [ ] Squats [(50kg x 10)]

```

6) This can now be viewed from the history.
    * To access the history:
      * "q" to go back to main menu.
      * navigate to history and select using enter.


## Background
I previously created a [gym tracker](https://github.com/lylecantcode/gym), to record my workouts on.
This did the job but was a bit unwieldy and was not a pleasant user experience.
I came across the package [BubbleTea](https://github.com/charmbracelet/bubbletea) and after a bit of experimentation it occured to me that it would be a great way to update my gym tracker.

My personal usage is via the android app [Termux](https://play.google.com/store/apps/details?id=com.termux&hl=en_GB&gl=US). This allows me to easily track my workouts on the fly but also further exaggerated the awkwardness of use. 

## Future Plans

### Completed
~~I would like to add a start menu, where you can choose between:~~  
~~1) Start your workout~~  
~~2) View past workouts (a list of days you can view, with an array of exercises done as a summary)~~  
~~3) Add remove/exercises~~  
~~4) View personal bests~~  


## Usage

```sh
$ go run main.go
# Starts the program and you're good to go!
```


## Contributing

Feel free to dive in! [Open an issue](https://github.com/lylecantcode/gym_tui/issues/new) or submit PRs.

## Known Bugs