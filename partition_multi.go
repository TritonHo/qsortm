package qsortm

import (
	"runtime"
	"sync"
)

type sliceRange struct {
	start, end int
}
type subtaskResult struct {
	left, right                   sliceRange
	leftRemaining, rightRemaining int
}
type subtask struct {
	left, right sliceRange
	pivotPos    int
	callbackCh  chan subtaskResult
}

func getNewLeftRange(unprocessedLeftIdx, unprocessedRightIdx *int, batchSize int) sliceRange {
	r := sliceRange{
		start: *unprocessedLeftIdx,
		end:   *unprocessedLeftIdx + batchSize,
	}
	if *unprocessedRightIdx < r.end {
		r.end = *unprocessedRightIdx
	}
	*unprocessedLeftIdx = r.end

	return r
}

func getNewRightRange(unprocessedLeftIdx, unprocessedRightIdx *int, batchSize int) sliceRange {
	r := sliceRange{
		start: *unprocessedRightIdx - batchSize,
		end:   *unprocessedRightIdx,
	}
	if r.start < *unprocessedLeftIdx {
		r.start = *unprocessedLeftIdx
	}
	*unprocessedRightIdx = r.start

	return r
}

func partitionMultiThread(data Interface, startPos, endPos, pivotPos int, subtaskCh chan subtask) (finalPivotPos int) {
	// swap the startPos with pivotPos first
	data.Swap(startPos, pivotPos)
	pivotPos = startPos

	threadNum := runtime.NumCPU()
	unprocessedLeftIdx := startPos + 1
	unprocessedRightIdx := endPos
	callbackCh := make(chan subtaskResult, threadNum)
	outstandingSubTaskCount := threadNum

	// create initial subtasks, which has 10% of the n
	initBatchSize := (endPos - startPos) / 10 / (2 * threadNum)
	for i := 0; i < threadNum; i++ {
		st := subtask{
			left:       getNewLeftRange(&unprocessedLeftIdx, &unprocessedRightIdx, initBatchSize),
			right:      getNewRightRange(&unprocessedLeftIdx, &unprocessedRightIdx, initBatchSize),
			pivotPos:   pivotPos,
			callbackCh: callbackCh,
		}
		subtaskCh <- st
	}

	unfinishedLefts := []sliceRange{}
	unfinishedRights := []sliceRange{}

	// FIXME: it should be a value that begin with large number
	// and then slowly decreased to small number
	const subTaskMinBatchSize = 100
	for {
		if outstandingSubTaskCount == 0 {
			break
		}
		// FIXME: determine better batchSize
		batchSize := (unprocessedRightIdx - unprocessedLeftIdx) / (4 * threadNum)
		if batchSize < subTaskMinBatchSize {
			batchSize = subTaskMinBatchSize
		}

		stResult := <-callbackCh

		unLeft := sliceRange{start: stResult.left.end - stResult.leftRemaining, end: stResult.left.end}
		unRight := sliceRange{start: stResult.right.start, end: stResult.right.start + stResult.rightRemaining}

		// the left has unfinished portion
		if unLeft.start != unLeft.end {
			nextSubTask := subtask{
				left:       unLeft,
				pivotPos:   pivotPos,
				callbackCh: callbackCh,
			}

			switch {
			case unprocessedLeftIdx < unprocessedRightIdx:
				nextSubTask.right = getNewRightRange(&unprocessedLeftIdx, &unprocessedRightIdx, batchSize)
				subtaskCh <- nextSubTask
			case len(unfinishedRights) > 0:
				nextSubTask.right = unfinishedRights[0]
				unfinishedRights = unfinishedRights[1:]
				subtaskCh <- nextSubTask
			default:
				// no further right tasks
				unfinishedLefts = append(unfinishedLefts, unLeft)
				outstandingSubTaskCount--
			}
			continue
		}
		// the right has unfinished portion
		if unRight.start != unRight.end {
			nextSubTask := subtask{
				right:      unRight,
				pivotPos:   pivotPos,
				callbackCh: callbackCh,
			}

			switch {
			case unprocessedLeftIdx < unprocessedRightIdx:
				nextSubTask.left = getNewLeftRange(&unprocessedLeftIdx, &unprocessedRightIdx, batchSize)
				subtaskCh <- nextSubTask
			case len(unfinishedLefts) > 0:
				nextSubTask.left = unfinishedLefts[0]
				unfinishedLefts = unfinishedLefts[1:]
				subtaskCh <- nextSubTask
			default:
				// no further left tasks
				unfinishedRights = append(unfinishedRights, unRight)
				outstandingSubTaskCount--
			}

			continue
		}

		// when the it reach this line, the previous subtask is a perfect match and left nothing unfinished
		if unprocessedLeftIdx < unprocessedRightIdx {
			// generate a new subtask
			st := subtask{
				left:       getNewLeftRange(&unprocessedLeftIdx, &unprocessedRightIdx, batchSize),
				right:      getNewRightRange(&unprocessedLeftIdx, &unprocessedRightIdx, batchSize),
				pivotPos:   pivotPos,
				callbackCh: callbackCh,
			}
			subtaskCh <- st
		} else {
			outstandingSubTaskCount--
		}
	}

	// find out the "middle" portion that need to perform qsort partition once again
	middleStart, middleEnd := unprocessedLeftIdx, unprocessedRightIdx
	for _, unLeft := range unfinishedLefts {
		if unLeft.start < middleStart {
			middleStart = unLeft.start
		}
	}
	for _, unRight := range unfinishedRights {
		if unRight.end > middleEnd {
			middleEnd = unRight.end
		}
	}

	// now we knows the middle portion that need partitioning
	// relocate the pivot to middleStart - 1
	data.Swap(pivotPos, middleStart-1)
	pivotPos = middleStart - 1

	// run the simple single thread qsort partitioning
	finalPivotPos = partitionSingleThread(data, middleStart-1, middleEnd, pivotPos)

	return finalPivotPos
}

func subTaskWorker(data Interface, subtaskCh chan subtask, subTaskWg *sync.WaitGroup) {
	defer subTaskWg.Done()

	for st := range subtaskCh {
		result := swappingOnBlock(data, st)
		st.callbackCh <- result
	}
}

func swappingOnBlock(data Interface, st subtask) subtaskResult {
	startIdx := st.left.start
	endIdx := st.right.end - 1

	for {
		// scan for the swapping pairs
		for startIdx < st.left.end && data.Less(startIdx, st.pivotPos) {
			startIdx++
		}
		for endIdx >= st.right.start && data.Less(st.pivotPos, endIdx) {
			endIdx--
		}

		if startIdx == st.left.end || endIdx < st.right.start {
			break
		}
		// perform swapping
		data.Swap(startIdx, endIdx)
	}

	result := subtaskResult{
		left:           st.left,
		right:          st.right,
		leftRemaining:  st.left.end - startIdx,
		rightRemaining: endIdx - st.right.start + 1,
	}
	return result
}
