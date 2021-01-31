package qsortm

import (
	"runtime"
	"sort"
	"sync"
)

type Interface sort.Interface

func taskWorker(data Interface, inputCh, outputCh chan task, subtaskCh chan subtask, wg, remainingTaskNum *sync.WaitGroup) {
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
		case n >= multiThreadThrehold && n > data.Len()/threadNum*2:
			// FIXME: choose a better pivot choosing algorithm instead of hardcoding
			pivotPos := t.startPos
			finalPivotPos := partitionMultiThread(data, t.startPos, t.endPos, pivotPos, subtaskCh)

			// add the sub-tasks to the queue
			remainingTaskNum.Add(2)
			outputCh <- task{startPos: t.startPos, endPos: finalPivotPos}
			outputCh <- task{startPos: finalPivotPos + 1, endPos: t.endPos}
		case n > threshold:
			// FIXME: choose a better pivot choosing algorithm instead of hardcoding
			pivotPos := t.startPos
			finalPivotPos := partitionSingleThread(data, t.startPos, t.endPos, pivotPos)

			// add the sub-tasks to the queue
			remainingTaskNum.Add(2)
			outputCh <- task{startPos: t.startPos, endPos: finalPivotPos}
			outputCh <- task{startPos: finalPivotPos + 1, endPos: t.endPos}
		case n >= 2:
			// for small n between 2 to threshold, we switch back to standard library
			// FIXME: handle it
			// sort.Ints(input[t.startPos:t.endPos])
		}

		// mark the current task is done
		remainingTaskNum.Done()
	}
}

func Sort(data Interface) {
	taskWg := &sync.WaitGroup{}
	subTaskWg := &sync.WaitGroup{}
	remainingTaskNum := &sync.WaitGroup{}

	threadNum := runtime.NumCPU()
	// ch1 link from inverter --> taskWorker, it should be unbuffered allow FILO behaviour in coordinator
	ch1 := make(chan task, 1)
	// ch2 link from taskWorker --> inverter, it pass the partitioned new task
	ch2 := make(chan task, 10*threadNum)
	// subtaskCh link from partitionMultiThread --> subTaskWorker
	subtaskCh := make(chan subtask, 10*threadNum)

	// init workers
	taskWg.Add(threadNum)
	subTaskWg.Add(threadNum)
	for i := 0; i < threadNum; i++ {
		go taskWorker(data, ch1, ch2, subtaskCh, taskWg, remainingTaskNum)
		go subTaskWorker(data, subtaskCh, subTaskWg)
	}

	// init the invertor
	go channelInverter(ch2, ch1)

	// add the input to channel
	remainingTaskNum.Add(1)
	ch1 <- task{startPos: 0, endPos: data.Len()}
	// wait for all task done
	remainingTaskNum.Wait()

	// let the worker threads die peacefully before exit
	// we must NOT have any zombie worker thread left
	close(ch2)
	close(subtaskCh)
	taskWg.Wait()
	subTaskWg.Wait()
}
