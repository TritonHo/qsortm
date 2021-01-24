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

func qsortPartitionMultiThread(input []int, startPos, endPos, pivotPos int, subtaskCh chan subtask) (finalPivotPos int) {
	// swap the startPos with pivotPos first
	input[startPos], input[pivotPos] = input[pivotPos], input[startPos]
	pivotPos = startPos

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
		unLeft := sliceRange{start: stResult.left.start + stResult.leftFinished, end: stResult.left.end}
		unRight := sliceRange{start: stResult.right.start, end: stResult.right.end - stResult.rightFinished}

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
				nextSubTask.left = getNewLeftRange(&unprocessedLeftIdx, &unprocessedRightIdx)
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
				left:       getNewLeftRange(&unprocessedLeftIdx, &unprocessedRightIdx),
				right:      getNewRightRange(&unprocessedLeftIdx, &unprocessedRightIdx),
				pivotPos:   pivotPos,
				callbackCh: callbackCh,
			}
			subtaskCh <- st
		} else {
			outstandingSubTaskCount--
		}

		if outstandingSubTaskCount == 0 {
			break
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
	input[pivotPos], input[middleStart-1] = input[middleStart-1], input[pivotPos]
	pivotPos = middleStart - 1

	// run the simple single thread qsort partitioning
	finalPivotPos = qsortPartition(input, middleStart-1, middleEnd, pivotPos)

	return finalPivotPos
}

func subTaskWorker(input []int, subtaskCh chan subtask, subTaskWg *sync.WaitGroup) {
	defer subTaskWg.Done()

	for st := range subtaskCh {
		result := subTaskInternal(input, st)
		st.callbackCh <- result
	}
}

func subTaskInternal(input []int, st subtask) subtaskResult {
	pivot := input[st.pivotPos]

	startIdx := st.left.start
	endIdx := st.right.end - 1

	for {
		// scan for the swapping pairs
		for startIdx < st.left.end && input[startIdx] <= pivot {
			startIdx++
		}
		for endIdx >= st.right.start && input[endIdx] > pivot {
			endIdx--
		}

		if startIdx == st.left.end || endIdx < st.right.start {
			break
		}
		// perform swapping
		input[startIdx], input[endIdx] = input[endIdx], input[startIdx]
	}

	result := subtaskResult{
		left:          st.left,
		right:         st.right,
		leftFinished:  st.left.end - startIdx,
		rightFinished: endIdx - st.right.start + 1,
	}
	return result
}

func qsortProdWorkerV2(input []int, inputCh, outputCh chan task, subtaskCh chan subtask, wg, remainingTaskNum *sync.WaitGroup) {

	// we use multithread version of partitioning, only when the n is large enough
	// also, when the input is broken into enough tasks, each task should use single thread partitioning instead
	threadNum := runtime.NumCPU()
	multiThreadThrehold := len(input) / threadNum
	if multiThreadThrehold < subTaskBatchSize*threadNum {
		multiThreadThrehold = subTaskBatchSize * threadNum
	}

	// if the size of the task is below threshold, it will use the standard library for sorting
	// too small threshold will cause necessary data exchange between threads and degrade performance
	const threshold = 10000

	defer wg.Done()

	for t := range inputCh {
		n := t.endPos - t.startPos
		switch {
		case n >= multiThreadThrehold:
			// FIXME: choose a better pivot choosing algorithm instead of hardcoding
			pivotPos := t.startPos
			finalPivotPos := qsortPartitionMultiThread(input, t.startPos, t.endPos, pivotPos, subtaskCh)

			// add the sub-tasks to the queue
			remainingTaskNum.Add(2)
			outputCh <- task{startPos: t.startPos, endPos: finalPivotPos}
			outputCh <- task{startPos: finalPivotPos + 1, endPos: t.endPos}
		case n > threshold:
			// FIXME: choose a better pivot choosing algorithm instead of hardcoding
			pivotPos := t.startPos
			finalPivotPos := qsortPartition(input, t.startPos, t.endPos, pivotPos)

			// add the sub-tasks to the queue
			remainingTaskNum.Add(2)
			outputCh <- task{startPos: t.startPos, endPos: finalPivotPos}
			outputCh <- task{startPos: finalPivotPos + 1, endPos: t.endPos}
		case n >= 2:
			// for small n between 2 to threshold, we switch back to standard library
			sort.Ints(input[t.startPos:t.endPos])
		}

		// mark the current task is done
		remainingTaskNum.Done()
	}
}

func qsortProdV2(input []int) {
	wg := &sync.WaitGroup{}
	subTaskWg := &sync.WaitGroup{}
	remainingTaskNum := &sync.WaitGroup{}

	threadNum := runtime.NumCPU()
	// ch1 link from inverter --> worker, it should be unbuffered allow FILO behaviour in coordinator
	ch1 := make(chan task, threadNum)
	// ch2 link from worker --> inverter, it pass the partitioned new task
	ch2 := make(chan task, 10*threadNum)
	// subtaskCh link from qsortPartitionMultiThread --> subTaskWorker
	subtaskCh := make(chan subtask, 10*threadNum)

	// init workers
	wg.Add(threadNum)
	subTaskWg.Add(threadNum)
	for i := 0; i < threadNum; i++ {
		go qsortProdWorker(input, ch1, ch2, wg, remainingTaskNum)
		go subTaskWorker(input, subtaskCh, subTaskWg)
	}

	// init the invertor
	go channelInverter(ch2, ch1)

	// add the input to channel
	remainingTaskNum.Add(1)
	ch1 <- task{startPos: 0, endPos: len(input)}
	// wait for all task done
	remainingTaskNum.Wait()

	// let the worker threads die peacefully
	close(ch2)
	close(subtaskCh)
	wg.Wait()
	subTaskWg.Wait()
}
