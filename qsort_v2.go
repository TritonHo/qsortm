package qsortm

import (
	"runtime"
	"sort"
	"sync"
)

func qsortProdWorkerV2(input []int, inputCh, outputCh chan task, subtaskCh chan subtask, wg, remainingTaskNum *sync.WaitGroup) {
	threadNum := runtime.NumCPU()

	// the multithread version of partitioning will be applied only when n is large
	// also, when the input is broken into enough tasks, each task should use single thread partitioning instead
	// FIXME: where is this 50000 comes from?
	const multiThreadThrehold = 50000

	// if the size of the task is below threshold, it will use the standard library for sorting
	// too small threshold will cause unnecessary data exchange between threads and degrade performance
	const threshold = 10000

	defer wg.Done()

	for t := range inputCh {
		n := t.endPos - t.startPos
		switch {
		case n >= multiThreadThrehold && n > len(input)/threadNum*2:
			// FIXME: choose a better pivot choosing algorithm instead of hardcoding
			pivotPos := t.startPos
			finalPivotPos := partitionMultiThread(input, t.startPos, t.endPos, pivotPos, subtaskCh)

			// add the sub-tasks to the queue
			remainingTaskNum.Add(2)
			outputCh <- task{startPos: t.startPos, endPos: finalPivotPos}
			outputCh <- task{startPos: finalPivotPos + 1, endPos: t.endPos}
		case n > threshold:
			// FIXME: choose a better pivot choosing algorithm instead of hardcoding
			pivotPos := t.startPos
			finalPivotPos := partitionSingleThread(input, t.startPos, t.endPos, pivotPos)

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
	ch1 := make(chan task, 1)
	// ch2 link from worker --> inverter, it pass the partitioned new task
	ch2 := make(chan task, 10*threadNum)
	// subtaskCh link from qsortPartitionMultiThread --> subTaskWorker
	subtaskCh := make(chan subtask, 10*threadNum)

	// init workers
	wg.Add(threadNum)
	subTaskWg.Add(threadNum)
	for i := 0; i < threadNum; i++ {
		go qsortProdWorkerV2(input, ch1, ch2, subtaskCh, wg, remainingTaskNum)
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
