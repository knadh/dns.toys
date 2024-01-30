package sudokusolver

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const incorrect_puzzle_string_format = "Incorrect sudoku puzzle string. Puzzle string must be in the row major order and each row separated by a dot. Empty puzzle cells must be replaced with 0 value. The string must consists of only digits [0, 9] and '.'"
const puzzle_not_solvable = "Puzzle could not be solved."

// stores the starting index of each of the 3x3 blocks. for example first 3x3 block's starting row and column are (0,0)
// for the last 3x3 block starting row and column are (6, 6)
var startRowColumn = [][][]int{
	{[]int{0, 0}, []int{0, 3}, []int{0, 6}},
	{[]int{3, 0}, []int{3, 3}, []int{3, 6}},
	{[]int{6, 0}, []int{6, 3}, []int{6, 6}},
}

type SudokuSolver struct{}

// New returns a new instance of SudokuSolver.
func New() *SudokuSolver {
	return &SudokuSolver{}
}

// Query converts a number to words.
func (solver *SudokuSolver) Query(q string) ([]string, error) {
	puzzle, err := stringToPuzzle(q)
	if err != nil {
		return nil, errors.New(incorrect_puzzle_string_format)
	}
	if solve(puzzle) {
		r := fmt.Sprintf(`%s 1 TXT "%s"`, q, puzzleToString(puzzle))
		return []string{r}, nil
	}
	return nil, errors.New(puzzle_not_solvable)
}

// Dump is not implemented in this package.
func (n *SudokuSolver) Dump() ([]byte, error) {
	return nil, nil
}

// converts string to puzzle which is of type [][]int
func stringToPuzzle(puzzleString string) ([][]int, error) {
	rows := strings.Split(puzzleString, ".")
	if len(rows) != 9 {
		return nil, errors.New(incorrect_puzzle_string_format)
	}
	puzzle := make([][]int, 0)
	i := 0
	for _, row := range rows {
		puzzle = append(puzzle, make([]int, 0))
		for _, char := range row {
			k, err := strconv.Atoi(string(char))
			if err != nil {
				return nil, fmt.Errorf(incorrect_puzzle_string_format)
			}
			puzzle[i] = append(puzzle[i], k)
		}
		i += 1
	}
	return puzzle, nil
}

// finds all the valid options for the cell
func getOptions(puzzle [][]int, row, col int) []int {
	var availableOptions = map[int]bool{
		1: true,
		2: true,
		3: true,
		4: true,
		5: true,
		6: true,
		7: true,
		8: true,
		9: true,
	}

	// finding exclusions along the column
	for i := 0; i < 9; i++ {
		temp := puzzle[row][i]
		if temp != 0 {
			availableOptions[temp] = false
		}
	}

	// finding exclusions along the row
	for i := 0; i < 9; i++ {
		temp := puzzle[i][col]
		if temp != 0 {
			availableOptions[temp] = false
		}
	}

	// finding exclusions in the block
	startRow, startCol := startRowColumn[int(row/3)][int(col/3)][0], startRowColumn[int(row/3)][int(col/3)][1]
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			temp := puzzle[startRow+i][startCol+j]
			if temp != 0 {
				availableOptions[temp] = false
			}
		}
	}
	options := make([]int, 0)
	for k, v := range availableOptions {
		if v {
			options = append(options, k)
		}
	}

	return options
}

// solves the sudoku puzzle
func solve(puzzle [][]int) bool {
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			if puzzle[row][col] == 0 {
				options := getOptions(puzzle, row, col)
				for _, opt := range options {
					puzzle[row][col] = opt
					if solve(puzzle) {
						return true
					} else {
						puzzle[row][col] = 0
					}
				}
				return false
			}
		}
	}
	return true
}

// converts puzzle which is of type [][]int back to string for response
func puzzleToString(puzzle [][]int) string {
	rowStrings := make([]string, 0)
	for _, row := range puzzle {
		for _, val := range row {
			rowStrings = append(rowStrings, fmt.Sprintf("%d", val))
		}
		rowStrings = append(rowStrings, ".")
	}
	return strings.Join(rowStrings, "")[:89]
}

// only for visualization purpose
func printPuzzle(puzzle [][]int) {
	rowString := ""
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			rowString += fmt.Sprintf("%d ", puzzle[row][col])
		}
		rowString += "\n"
	}
	fmt.Printf(rowString)
}
