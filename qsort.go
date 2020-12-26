package qsortm

import (
	//	"log"

	"sort"

	"runtime"
	"sync"
)

func channelInverterV2(inputCh, outputCh chan []int) {
	taskBuffer := [][]int{}
	for {
		task, ok := <-inputCh
		if !ok {
			// the input channel has already closed
			break
		}
		taskBuffer = append(taskBuffer, task)
		for len(taskBuffer) > 0 {
			select {
			case outputCh <- taskBuffer[len(taskBuffer)-1]:
				taskBuffer = taskBuffer[:len(taskBuffer)-1]
			case newTask := <-inputCh:
				taskBuffer = append(taskBuffer, newTask)
			}
		}
	}

	close(outputCh)
}

func qsortProdWorker(inputTaskCh, subTaskCh chan []int, wg, remainingTaskNum *sync.WaitGroup) {

	// if the size of the task is below threshold, it will use the standard library for sorting
	// too small threshold will cause necessary data exchange between threads and degrade performance
	const threshold = 10000

	defer wg.Done()

	for task := range inputTaskCh {
		n := len(task)
		switch {
		case n > threshold:
			pivotPos := qsortPartition(task)

			// add the sub-tasks to the queue
			remainingTaskNum.Add(2)
			subTaskCh <- task[:pivotPos]
			subTaskCh <- task[pivotPos+1:]
		case n >= 2:
			// for small n between 2 to threshold, we switch back to standard library
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
	ch1 := make(chan []int, threadNum)
	// ch2 link from worker --> inverter, it pass the sub-task
	ch2 := make(chan []int, 100*threadNum)

	wg.Add(threadNum)
	for i := 0; i < threadNum; i++ {
		go qsortProdWorker(ch1, ch2, &wg, &remainingTaskNum)
	}

	go channelInverterV2(ch2, ch1)

	// add the input to channel
	remainingTaskNum.Add(1)
	ch1 <- input
	remainingTaskNum.Wait()

	// wait for all task done, and the worker thread die peacefully
	close(ch2)
	wg.Wait()
}
