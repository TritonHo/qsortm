package qsortm

import (
	// "log"

	//    "math/rand"
	"runtime"
	"sort"
	"sync"
)

type bucket struct {
	startPos, endPos         int
	completedPos, cleanedPos int
	lock                     *sync.Mutex
}

const bufferMaxSize = 100

func bucketWorkerV3(input, finalizedPivotPositions []int, buckets []bucket, qsortWorkerCh chan []int, initBucket int) {
	localBuffer := map[int][]int{}
	for i := 0; i < len(finalizedPivotPositions)+1; i++ {
		localBuffer[i] = make([]int, 0, bufferMaxSize)
	}
	/*
		subSlice := input[startPos:endPos]
		for i, item := range subSlice {
			fn := func(i int) bool { return item <= input[finalizedPivotPositions[i]] }
			bucketIndex := sort.Search(len(finalizedPivotPositions), fn)

			if bucketIndex != workerIndex {
				// put the item into corresponding bucket channel, and then get back the item from the assigned channel
				exchangeChannels[bucketIndex] <- item
				subSlice[i] = <-exchangeChannels[workerIndex]
			}
		}

		// after the bucket is finished, pass the subslice to qsort for further processing
		qsortWorkerCh <- subSlice
	*/
}

func bufferToBusket(slice []int, buff *[]int, b *bucket) {
	workingBuffer := *buff
	for len(workingBuffer) > 0 && b.completedPos < b.cleanedPos {
		slice[b.completedPos] = workingBuffer[len(workingBuffer)-1]
		b.completedPos++
		workingBuffer = workingBuffer[:len(workingBuffer)-1]
	}
	buff = &workingBuffer
}

func workOnBucket(input, finalizedPivotPositions []int, buckets []bucket, localBuffer map[int][]int, workingIndex int) (nextBucketIdx int) {
	b := &buckets[workingIndex]
	nextBucketIdx = -1

	b.lock.Lock()
	defer b.lock.Unlock()

	subSlice := input[b.startPos:b.endPos]
	workingBuffer := localBuffer[workingIndex]

	// step 1: put the data from buffer back to the slice
	bufferToBusket(subSlice, &workingBuffer, b)

	// step 2:
	for b.cleanedPos < len(subSlice) {
		item := subSlice[b.cleanedPos]
		b.cleanedPos++

		fn := func(i int) bool { return item <= input[finalizedPivotPositions[i]] }
		targetIndex := sort.Search(len(finalizedPivotPositions), fn)

		if targetIndex != bucketIdx {
			localBuffer[targetIndex] = append(localBuffer[targetIndex], item)
			if len(localBuffer[targetIndex]) >= bufferMaxSize {
				nextBucketIdx = targetIndex
				break
			}
		} else {
			subSlice[b.completedPos] = item
			b.completedPos++
		}
	}

	// step 3: once again, put the data from the buffer back to the slice
	bufferToBusket(subSlice, &workingBuffer, b)

	localBuffer[workingIndex] = workingBuffer

	return nextBucketIdx
}

func qsortWithBucketV3(input []int) {

	threadNum := runtime.NumCPU()

	// prepare the pivots, and then move the pivots to final location
	pivotPositions := getPivotPositions(input, 10240)
	pivots := countBucketSize(input, pivotPositions)
	mergedPivots := mergePivots(input, pivots, 1000)
	finalizedPivotPositions := relocatePivots(input, mergedPivots)
	pivotCount := len(finalizedPivotPositions)

	wg := sync.WaitGroup{}
	remainingTaskNum := sync.WaitGroup{}

	// ch1 link from inverter --> worker, it should be unbuffered allow FILO behaviour in coordinator
	ch1 := make(chan []int, threadNum)
	// ch2 link from worker --> inverter, it pass the sub-task
	ch2 := make(chan []int, 100*threadNum)

	// start the qsort worker
	wg.Add(threadNum)
	for i := 0; i < threadNum; i++ {
		go qsortProdWorker(ch1, ch2, &wg, &remainingTaskNum)
	}
	// start the qsort channel inverter
	go channelInverterV2(ch2, ch1)

	// start the bucket workers
	remainingTaskNum.Add(pivotCount + 1)

	// add the starting and ending point
	temp := []int{-1}
	temp = append(temp, finalizedPivotPositions...)
	temp = append(temp, len(input))

	for i := 0; i < len(temp)-1; i++ {
		startPos := temp[i] + 1
		endPos := temp[i+1]
		go bucketWorker(input, finalizedPivotPositions, exchangeChannels, startPos, endPos, i, ch1)
	}

	// wait for all task done, and the qsort worker thread die peacefully
	remainingTaskNum.Wait()
	close(ch2)
	wg.Wait()
}
