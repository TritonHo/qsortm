package qsortm

import (
	//	"log"

	"sort"

	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

func channelInverter(inputCh, outputCh chan []int) {
	// we use atomic int as a end singal
	var noMoreInput int32 = 0

	taskBufferLock := sync.Mutex{}
	taskBuffer := [][]int{}

	go func() {
		for task := range inputCh {
			taskBufferLock.Lock()
			taskBuffer = append(taskBuffer, task)
			taskBufferLock.Unlock()
		}
		atomic.StoreInt32(&noMoreInput, 1)
	}()

	for {
		// test if the input
		for len(taskBuffer) > 0 {
			// get the last task in the buffer
			taskBufferLock.Lock()
			task := taskBuffer[len(taskBuffer)-1]
			taskBuffer = taskBuffer[:len(taskBuffer)-1]
			taskBufferLock.Unlock()

			outputCh <- task
		}
		// if no more input, then we can end the loop
		if atomic.LoadInt32(&noMoreInput) == 1 {
			break
		}
		// sleep some tiny seconds to avoid CPU-time waste in for(1) loop
		time.Sleep(10 * time.Microsecond)
	}
	close(outputCh)
}

func qsortProdWorker(inputTaskCh, subTaskCh chan []int, wg, remainingTaskNum *sync.WaitGroup) {
	defer wg.Done()

	for task := range inputTaskCh {
		n := len(task)
		switch {
		case n > 100:
			pivotPos := qsortPartition(task)

			// add the sub-tasks to the queue
			remainingTaskNum.Add(2)
			subTaskCh <- task[:pivotPos]
			subTaskCh <- task[pivotPos+1:]
		case n >= 2:
			// for small n between 2 to 100, we switch back to standard library
			sort.Ints(task)
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
	ch1 := make(chan []int)
	// ch2 link from worker --> inverter, it pass the sub-task
	ch2 := make(chan []int, 10)

	wg.Add(threadNum)
	for i := 0; i < threadNum; i++ {
		go qsortProdWorker(ch1, ch2, &wg, &remainingTaskNum)
	}

	go channelInverter(ch2, ch1)

	// add the input to channel
	remainingTaskNum.Add(1)
	ch1 <- input
	remainingTaskNum.Wait()

	// wait for all task done, and the worker thread die peacefully
	close(ch2)
	wg.Wait()
}
