package utils

import "regexp"

type Matrix [][]string

func ParseStdin(stdin, columnSeq, lineSeq string) ([]string, *Matrix) {
	lineSpliter := regexp.MustCompile(lineSeq)
	columnSpilter := regexp.MustCompile(columnSeq)

	lines := lineSpliter.Split(stdin, -1)

	matrixPtr := new(Matrix)

	for i := 0; i < len(lines); i += 1 {
		*matrixPtr = append(*matrixPtr, columnSpilter.Split(lines[i], -1))
	}

	return lines, matrixPtr
}
