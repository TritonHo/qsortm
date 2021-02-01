package qsortm

import (
	"runtime"
	"sort"
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
	sort.Sort(byLeft(unLefts))
	sort.Sort(byRight(unRights))

	// step 2: do the swapping, until one side exhausted
	isMiddleUsed := false
	unMiddle := sliceRange{start: unprocessedLeftIdx, end: unprocessedRightIdx}
	for len(unLefts) > 0 && len(unRights) > 0 {
		leftRemaining, rightRemaining := swappingOnBlock(data, unLefts[0], unRights[0], pivotPos)

		unLefts[0] = unLefts[0].getNewLeft(leftRemaining)
		unRights[0] = unRights[0].getNewRight(rightRemaining)
		if unLefts[0].start == unLefts[0].end {
			unLefts = unLefts[1:]
		}
		if unRights[0].start == unRights[0].end {
			unRights = unRights[1:]
		}

		// if one side exhaused, add the middle and continue
		if len(unLefts) == 0 && isMiddleUsed == false {
			isMiddleUsed = true
			unLefts = append(unLefts, unMiddle)
			unprocessedLeftIdx = unprocessedRightIdx
		}
		if len(unRights) == 0 && isMiddleUsed == false {
			isMiddleUsed = true
			unRights = append(unRights, unMiddle)
			unprocessedRightIdx = unprocessedLeftIdx
		}
	}

	// find out the "middle" portion that need to perform qsort partition once again
	middleStart, middleEnd := unprocessedLeftIdx, unprocessedRightIdx
	for _, unLeft := range unLefts {
		if unLeft.start < middleStart {
			middleStart = unLeft.start
		}
	}
	for _, unRight := range unRights {
		if unRight.end > middleEnd {
			middleEnd = unRight.end
		}
	}

	return middleStart, middleEnd
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

	// it should be a value that begin with large number
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

	// complete the unprocessed fragments in previous steps
	// also, find out the "middle" portion that need to perform qsort partition once again
	middleStart, middleEnd := handleFragments(data, unfinishedLefts, unfinishedRights, unprocessedLeftIdx, unprocessedRightIdx, pivotPos)

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
		for startIdx < left.end && data.Less(pivotPos, startIdx) == false {
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
		startIdx++
		endIdx--
	}

	leftRemaining = left.end - startIdx
	rightRemaining = endIdx - right.start + 1

	return leftRemaining, rightRemaining
}
