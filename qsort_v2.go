package qsortm

import (
	"runtime"
)

type sliceRange struct {
	start, end int
}
type subtaskResult struct {
	left, right                 sliceRange
	leftFinished, rightFinished int
}
type subtask struct {
	left, right sliceRange
	pivotPos    int
	callbackCh  chan subtaskResult
}

const subTaskBatchSize = 1000

func getNewLeftRange(unprocessedLeftIdx, unprocessedRightIdx *int) sliceRange {
	r := sliceRange{
		start: *unprocessedLeftIdx,
		end:   *unprocessedLeftIdx + subTaskBatchSize,
	}
	if *unprocessedRightIdx < r.end {
		r.end = *unprocessedRightIdx
	}
	*unprocessedLeftIdx = r.end

	return r
}

func getNewRightRange(unprocessedLeftIdx, unprocessedRightIdx *int) sliceRange {
	r := sliceRange{
		start: *unprocessedRightIdx - subTaskBatchSize,
		end:   *unprocessedRightIdx,
	}
	if r.start < *unprocessedLeftIdx {
		r.start = *unprocessedLeftIdx
	}
	*unprocessedRightIdx = r.start

	return r
}

func qsortPartitionMulti(input []int, startPos, endPos int, subtaskCh chan subtask) (finalPivotPos int) {
	// FIXME: fix the pivot selection
	pivotPos := startPos

	// swap the startPos with pivotPos first
	input[startPos], input[pivotPos] = input[pivotPos], input[startPos]

	threadNum := runtime.NumCPU()
	unprocessedLeftIdx := startPos + 1
	unprocessedRightIdx := endPos
	callbackCh := make(chan subtaskResult, threadNum)
	outstandingSubTaskCount := threadNum

	// remarks: we only allows the task with N much larger than subTaskBatchSize * threadNum use this function.
	// this no need for handling for small N
	// FIXME: is this rule hold true?
	for i := 0; i < threadNum; i++ {
		st := subtask{
			left:       getNewLeftRange(&unprocessedLeftIdx, &unprocessedRightIdx),
			right:      getNewRightRange(&unprocessedLeftIdx, &unprocessedRightIdx),
			pivotPos:   pivotPos,
			callbackCh: callbackCh,
		}
		subtaskCh <- st
	}

	unfinishedLefts := []sliceRange{}
	unfinishedRights := []sliceRange{}

	for {
		stResult := <-callbackCh
		unLeft := sliceRange{start: stResult.left.start + leftFinished, end: stResult.left.end}
		unRight := sliceRange{start: stResult.right.start, end: stResult.right.end - rightFinished}

		// the left has unfinished portion
		if unLeft.start != unLeft.end {
			nextSubTask := subtask{
				left:       unLeft,
				pivotPos:   pivotPos,
				callbackCh: callbackCh,
			}

			switch {
			case unprocessedLeftIdx < unprocessedRightIdx:
				nextSubTask.right = getNewRightRange(&unprocessedLeftIdx, &unprocessedRightIdx)
				subtaskCh <- nextSubTask
			case len(unfinishedRights) > 0:
				nextSubTask.right = unfinishedRights[0]
				unfinishedRights = unfinishedRights[1:]
				subtaskCh <- nextSubTask
			default:
				unfinishedLefts = append(unfinishedLefts, unLeft)
				outstandingSubTaskCount--
			}
		}
		// the right has unfinished portion
		if unfinishedRight.start != unfinishedRight.end {
			continue
		}
	}

	// FIXME: implement it
	return
}
