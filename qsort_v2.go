package qsortm

import (
//	"log"
/*
	"sort"

	"runtime"
	"sync"
*/
)

type subtaskResult struct {
	leftStart, leftEnd, rightStart, rightEnd int
	leftFinished, rightFinished              int
}
type subtask struct {
	leftStart, leftEnd, rightStart, rightEnd int
	pivotPos                                 int
	callbackCh                               chan subtaskResult
}

/*
const subTaskBatchSize = 1000
func qsortPartitionMulti(input []int, startPos, endPos int, subtaskCh chan subtask) (pivotPos int) {
	// FIXME: fix the pivot selection
	pivot := input[startPos]

	startIdx := startPos + 1
	endIdx := endPos - 1

		threadNum := runtime.NumCPU()
}
*/
