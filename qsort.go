package qsortm

import (
	//	"log"

	"sort"

	"runtime"
	"sync"
)

func qsortProdWorker(input []int, inputCh, outputCh chan task, wg, remainingTaskNum *sync.WaitGroup) {

	// if the size of the task is below threshold, it will use the standard library for sorting
	// too small threshold will cause necessary data exchange between threads and degrade performance
	const threshold = 10000

	defer wg.Done()

	for t := range inputCh {
		n := t.endPos - t.startPos
		switch {
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

func qsortProd(input []int) {
	wg := sync.WaitGroup{}
	remainingTaskNum := sync.WaitGroup{}

	threadNum := runtime.NumCPU()
	// ch1 link from inverter --> worker, it should be unbuffered allow FILO behaviour in coordinator
	ch1 := make(chan task, threadNum)
	// ch2 link from worker --> inverter, it pass the sub-task
	ch2 := make(chan task, 100*threadNum)

	wg.Add(threadNum)
	for i := 0; i < threadNum; i++ {
		go qsortProdWorker(input, ch1, ch2, &wg, &remainingTaskNum)
	}

	go channelInverter(ch2, ch1)

	// add the input to channel
	remainingTaskNum.Add(1)
	ch1 <- task{startPos: 0, endPos: len(input)}
	remainingTaskNum.Wait()

	// wait for all task done, and the worker thread die peacefully
	close(ch2)
	wg.Wait()
}
