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

type byLeft []sliceRange

func (a byLeft) Len() int           { return len(a) }
func (a byLeft) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byLeft) Less(i, j int) bool { return a[i].start < a[j].start }

type byRight []sliceRange

func (a byRight) Len() int           { return len(a) }
func (a byRight) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byRight) Less(i, j int) bool { return a[i].end > a[j].end }

func (sr *sliceRange) getNewLeft(leftRemaining int) sliceRange {
	return sliceRange{start: sr.end - leftRemaining, end: sr.end}
}
func (sr *sliceRange) getNewRight(rightRemaining int) sliceRange {
	return sliceRange{start: sr.start, end: sr.start + rightRemaining}
}

func handleFragments(data Interface, unLefts, unRights []sliceRange, unprocessedLeftIdx, unprocessedRightIdx, pivotPos int) (middleLeft, middleRight int) {
	// step 1: sort the left and right, by position
	insertionSort(byLeft(unLefts), 0, len(unLefts))
	insertionSort(byRight(unRights), 0, len(unRights))

	// step 2: find out the leftest left-block, and rightest right-block, and do the swapping
	for len(unLefts) > 0 && len(unRights) > 0 {
		leftRemaining, rightRemaining := swappingOnBlock(data, unLefts[0], unRights[0], pivotPos)

		newLeft := unLefts[0].getNewLeft(leftRemaining)
		newRight := unRights[0].getNewRight(rightRemaining)
		if newLeft.start == newLeft.end {
			unLefts = unLefts[1:]
		} else {
			unLefts[0] = newLeft
		}
		if newRight.start == newRight.end {
			unRights = unRights[1:]
		} else {
			unRights[0] = newRight
		}
	}
	/*
		unMiddle := {start: unprocessedLeftIdx, end: unprocessedRightIdx}
		if len(unLefts) > 0 {
			leftRemaining, rightRemaining := swappingOnBlock(data, unLefts[0], unMiddle, pivotPos)

			newLeft := sliceRange{start: unLefts[0].end - leftRemaining, end: unLefts[0].end}
			newRight := sliceRange{start: unRights[0].start, end: unRights[0].start + rightRemaining}

		}
	*/
	// FIXME: implement it
	return 0, 0
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

	// create initial subtasks, which has 25% of the n
	initBatchSize := (endPos - startPos) / 4 / (2 * threadNum)
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

	//  it should be a value that begin with large number
	// and then slowly decreased to small number
	const subTaskMinBatchSize = 100
	reservedUnprocessed := subTaskMinBatchSize * threadNum
	for {
		if outstandingSubTaskCount == 0 {
			break
		}
		// FIXME: determine better batchSize
		batchSize := (unprocessedRightIdx - unprocessedLeftIdx) / (2 * threadNum)
		if batchSize < subTaskMinBatchSize {
			batchSize = subTaskMinBatchSize
		}

		stResult := <-callbackCh

		unLeft := stResult.left.getNewLeft(stResult.leftRemaining)
		unRight := stResult.right.getNewRight(stResult.rightRemaining)

		// the left has unfinished portion
		if unLeft.start != unLeft.end {
			nextSubTask := subtask{
				left:       unLeft,
				pivotPos:   pivotPos,
				callbackCh: callbackCh,
			}

			if unprocessedRightIdx-unprocessedLeftIdx > reservedUnprocessed {
				nextSubTask.right = getNewRightRange(&unprocessedLeftIdx, &unprocessedRightIdx, batchSize)
				subtaskCh <- nextSubTask
			} else {
				// stop further processing and remember the unprocessed prositions
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

			if unprocessedRightIdx-unprocessedLeftIdx > reservedUnprocessed {
				nextSubTask.left = getNewLeftRange(&unprocessedLeftIdx, &unprocessedRightIdx, batchSize)
				subtaskCh <- nextSubTask
			} else {
				// stop further processing and remember the unprocessed prositions
				unfinishedRights = append(unfinishedRights, unRight)
				outstandingSubTaskCount--
			}
			continue
		}

		// when the it reach this line, the previous subtask is a perfect match and left nothing unfinished
		if unprocessedRightIdx-unprocessedLeftIdx > reservedUnprocessed {
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

	// FIXME: handle the remaining

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
		result := subtaskResult{
			left:  st.left,
			right: st.right,
		}

		result.leftRemaining, result.rightRemaining = swappingOnBlock(data, st.left, st.right, st.pivotPos)

		st.callbackCh <- result
	}
}

func swappingOnBlock(data Interface, left, right sliceRange, pivotPos int) (leftRemaining, rightRemaining int) {

	//st subtask) subtaskResult {
	startIdx := left.start
	endIdx := right.end - 1

	for {
		// scan for the swapping pairs
		for startIdx < left.end && data.Less(startIdx, pivotPos) {
			startIdx++
		}
		for endIdx >= right.start && data.Less(pivotPos, endIdx) {
			endIdx--
		}

		if startIdx == left.end || endIdx < right.start {
			break
		}
		// perform swapping
		data.Swap(startIdx, endIdx)
	}

	leftRemaining = left.end - startIdx
	rightRemaining = endIdx - right.start + 1

	return leftRemaining, rightRemaining
}
