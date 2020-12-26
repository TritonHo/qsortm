package qsortm

import (
	//	"log"

	"math/rand"
	"runtime"
	"sort"
	//"sync"
)

func getPivotPositions(input []int, n int) (pivotPositions []int) {
	pivotPositions = make([]int, n, n)
	m := map[int]bool{}
	for i := 0; i < n; i++ {
		for {
			pos := rand.Intn(len(input))
			if _, ok := m[pos]; !ok {
				m[pos] = true
				pivotPositions[i] = pos
			}
		}
	}

	// sort the pivots according to the value
	lessFn := func(i, j int) bool {
		v1 := input[pivotPositions[i]]
		v2 := input[pivotPositions[j]]
		return v1 < v2
	}
	sort.Slice(pivotPositions, lessFn)

	return pivotPositions
}

func countBucketSizeWorker(input, subSlice, pivotPositions []int, resultCh chan []int) {
	// since Search return n if "not found", we +1 make the code more clean
	counts := make([]int, len(pivotPositions)+1, len(pivotPositions)+1)
	for i := range counts {
		counts[i] = 0
	}

	for _, value := range subSlice {
		fn := func(i int) bool { return value <= input[pivotPositions[i]] }
		bucketIndex := sort.Search(len(pivotPositions), fn)
		counts[bucketIndex] = counts[bucketIndex] + 1
	}
	resultCh <- counts[:len(pivotPositions)]
}

func countBucketSize(input, pivotPositions []int) (counts []int) {
	threadNum := runtime.NumCPU()
	resultCh := make(chan []int, threadNum)

	for i := 0; i < threadNum; i++ {
		startPos := len(input) * i / threadNum
		endPos := len(input) * (i + 1) / threadNum
		subSlice := input[startPos:endPos]
		go countBucketSizeWorker(input, subSlice, pivotPositions, resultCh)
	}

	counts = make([]int, len(pivotPositions), len(pivotPositions))
	for i := range counts {
		counts[i] = 0
	}

	// merge the results from workers
	for i := 0; i < threadNum; i++ {
		subCounts := <-resultCh
		for i, c := range subCounts {
			counts[i] = counts[i] + c
		}
	}

	return counts
}

func mergePivots(input, pivotPositions, counts []int, target int) (mergedPivots, mergedCounts []int) {
	threhold := len(input) / target

	mergedPivots = []int{}
	mergedCounts = []int{}

	total := 0
	for i, pos := range pivotPositions {
		total = total + counts[i]
		if total >= threhold {
			mergedPivots = append(mergedPivots, pos)
			mergedCounts = append(mergedPivots, total)
			total = 0
		}
	}

	return mergedPivots, mergedCounts
}

func relocatePivots(input, pivotPositions, counts []int) (finalizedPivotPositions []int) {
	finalizedPivotPositions = make([]int, len(counts), len(counts))

	for i, originalPos := range pivotPositions {
		newPos := counts[i] - 1
		// swap the content
		input[newPos], input[originalPos] = input[originalPos], input[newPos]

		finalizedPivotPositions[i] = newPos
	}
	return finalizedPivotPositions
}

func bucketWorker(input, subSlice, finalizedPivotPositions []int, exchangeCh []chan int, workerIndex int, qsortWorkerCh chan []int) {
	for i, item := range subSlice {
		fn := func(i int) bool { return item <= input[finalizedPivotPositions[i]] }
		bucketIndex := sort.Search(len(finalizedPivotPositions), fn)

		if bucketIndex != workerIndex {
			// put the item into corresponding bucket channel, and then get back the item from the assigned channel
			exchangeCh[bucketIndex] <- item
			subSlice[i] = <-exchangeCh[workerIndex]
		}
	}

	// after the bucket is finished, pass the subslice to qsort for further processing
	qsortWorkerCh <- subSlice
}

/*
func qsortWithBucket(input []int) {
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
*/
