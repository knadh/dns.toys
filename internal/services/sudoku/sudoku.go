package sudoku

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrBadInput = errors.New("invalid puzzle string. See dig help ..")

// startRowColumns stores the starting index of each of the 3x3 blocks. for example first 3x3 block's starting row and column are (0,0)
// for the last 3x3 block starting row and column are (6, 6)
var startRowColumns = [][][]int{
	{[]int{0, 0}, []int{0, 3}, []int{0, 6}},
	{[]int{3, 0}, []int{3, 3}, []int{3, 6}},
	{[]int{6, 0}, []int{6, 3}, []int{6, 6}},
}

// TTL is set to 900 seconds (15 minutes).
const TTL = 900

type Sudoku struct{}

// New returns a new instance of Sudoku.
func New() *Sudoku {
	return &Sudoku{}
}

// Query converts a number to words.
func (s *Sudoku) Query(q string) ([]string, error) {
	puzzle, err := s.parsePuzzleString(q)
	if err != nil {
		return nil, ErrBadInput
	}

	if s.solvePuzzle(puzzle) {
		return []string{fmt.Sprintf(`%s %d TXT "%s"`, q, TTL, s.puzzleToString(puzzle))}, nil
	}

	return nil, errors.New("puzzle could not be solved.")
}

// Dump is not implemented in this package.
func (s *Sudoku) Dump() ([]byte, error) {
	return nil, nil
}

// parsePuzzleString parses the string representation into a [][]int grid.
func (s *Sudoku) parsePuzzleString(p string) ([][]int, error) {
	rows := strings.Split(p, ".")
	if len(rows) != 9 {
		return nil, ErrBadInput
	}

	out := make([][]int, 0, len(rows))
	for i, row := range rows {
		if len(row) != 9 {
			return nil, ErrBadInput
		}

		out = append(out, make([]int, 0, len(row)))
		for _, char := range row {
			k, err := strconv.Atoi(string(char))
			if err != nil {
				return nil, ErrBadInput
			}
			out[i] = append(out[i], k)
		}
	}
	return out, nil
}

// getValidValues returns the possible options for a cell.
func (s *Sudoku) getValidValues(puzzle [][]int, row, col int) []int {
	available := make(map[int]bool, 9)
	for n := 1; n <= 9; n++ {
		available[n] = true
	}

	// Find exclusions along the column.
	for i := 0; i < 9; i++ {
		v := puzzle[row][i]
		if v != 0 {
			available[v] = false
		}
	}

	// Find exclusions along the row.
	for i := 0; i < 9; i++ {
		v := puzzle[i][col]
		if v != 0 {
			available[v] = false
		}
	}

	// Find exclusions in the block.
	startRow, startCol := startRowColumns[int(row/3)][int(col/3)][0], startRowColumns[int(row/3)][int(col/3)][1]
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			v := puzzle[startRow+i][startCol+j]
			if v != 0 {
				available[v] = false
			}
		}
	}

	out := make([]int, 0, len(available))
	for k, v := range available {
		if v {
			out = append(out, k)
		}
	}

	return out
}

// solvePuzzle recursively solves the given puzzle.
func (s *Sudoku) solvePuzzle(puzzle [][]int) bool {
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			if puzzle[row][col] == 0 {
				options := s.getValidValues(puzzle, row, col)

				for _, opt := range options {
					puzzle[row][col] = opt
					if s.solvePuzzle(puzzle) {
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

// puzzleToString converts the given int representation to string.
func (s *Sudoku) puzzleToString(puzzle [][]int) string {
	var (
		out = make([]string, 0, len(puzzle))
		val = strings.Builder{}
	)
	for _, row := range puzzle {
		for _, col := range row {
			val.WriteString(fmt.Sprintf("%d", col))
		}
		out = append(out, val.String())
		val.Reset()
	}

	return strings.Join(out, ".")
}
