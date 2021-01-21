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
	lock                     *sync.RWMutex
}

const bufferMaxSize = 100

func bucketWorkerV3(input, finalizedPivotPositions []int, buckets []bucket, qsortWorkerCh chan []int, initBucket int) {
	// first, build the local buffer
	localBuffer := map[int][]int{}
	for i := 0; i < len(finalizedPivotPositions)+1; i++ {
		localBuffer[i] = make([]int, 0, bufferMaxSize)
	}

	bucketIndex := initBucket
	for {
		nextIndex := workOnBucket(input, finalizedPivotPositions, buckets, qsortWorkerCh, localBuffer, bucketIndex)

		// if bucketIndex is -1, then randomly choose a bucket with non-empty buffer
		if nextIndex == -1 {
			for index, buffer := range localBuffer {
				b := buckets[index]
				b.lock.RLock()
				itemsToBeCleaned := b.endPos - b.startPos - b.cleanedPos
				b.lock.RUnlock()
				if itemsToBeCleaned > 0 || len(buffer) > 0 {
					nextIndex = index
					break
				}
			}
		}
		if nextIndex == -1 {
			// there is no further bucket need working, this worker can be returned
			return
		}
		bucketIndex = nextIndex
	}
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

// nextBucketIdx == -1 means the workingBucket has all off-bucket item cleaned
func workOnBucket(input, finalizedPivotPositions []int, buckets []bucket, qsortWorkerCh chan []int, localBuffer map[int][]int, workingIndex int) (nextBucketIdx int) {
	b := &buckets[workingIndex]
	nextBucketIdx = -1

	b.lock.Lock()
	defer b.lock.Unlock()

	completedPosOriginal := b.completedPos
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

		if targetIndex != workingIndex {
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

	// step 4: if the bucket is fully completed, pass it to the successive qsort worker
	if completedPosOriginal != b.completedPos && b.completedPos == len(subSlice) {
		qsortWorkerCh <- subSlice
	}

	return nextBucketIdx
}

func qsortWithBucketV3(input []int) {

	threadNum := runtime.NumCPU()

	// prepare the pivots, and then move the pivots to final location
	pivotPositions := getPivotPositions(input, 10240)
	pivots := countBucketSize(input, pivotPositions)
	mergedPivots := mergePivots(input, pivots, 4096-1)
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

	// add the number of remainingTasks
	remainingTaskNum.Add(pivotCount + 1)

	// build the starting and ending point of each bucket
	temp := append([]int{-1}, finalizedPivotPositions...)
	temp = append(temp, len(input))

	buckets := make([]bucket, pivotCount+1, pivotCount+1)
	for i := 0; i < len(temp)-1; i++ {
		buckets[i] = bucket{
			startPos:     temp[i] + 1,
			endPos:       temp[i+1],
			completedPos: 0,
			cleanedPos:   0,
			lock:         &sync.RWMutex{},
		}
	}

	// start the bucket workers
	for i := 0; i < runtime.NumCPU()*2; i++ {
		go bucketWorkerV3(input, finalizedPivotPositions, buckets, ch2, i)
	}

	// wait for all task done, and the qsort worker thread die peacefully
	remainingTaskNum.Wait()
	close(ch2)
	wg.Wait()
}
