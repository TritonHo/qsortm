package qsortm

import (
	"log"

	"math/rand"
	"runtime"
	"sort"
	"sync"
)

// remarks: the pivotPositions is sorted based on the referencing value
func getPivotPositions(input []int, n int) (pivotPositions []int) {
	pivotPositions = make([]int, n, n)
	m := map[int]bool{}
	for i := 0; i < n; i++ {
		for {
			pos := rand.Intn(len(input))
			if _, ok := m[pos]; !ok {
				m[pos] = true
				pivotPositions[i] = pos
				break
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

// FIXME give better naming
type pivotWithCount struct {
	pos   int
	count int
}

// remarks: the returned pivots is still sorted based on referencing value
func countBucketSize(input, pivotPositions []int) (pivots []pivotWithCount) {
	threadNum := runtime.NumCPU()
	resultCh := make(chan []int, threadNum)

	for i := 0; i < threadNum; i++ {
		startPos := len(input) * i / threadNum
		endPos := len(input) * (i + 1) / threadNum
		subSlice := input[startPos:endPos]
		go countBucketSizeWorker(input, subSlice, pivotPositions, resultCh)
	}

	pivots = make([]pivotWithCount, len(pivotPositions), len(pivotPositions))
	for i := range pivots {
		pivots[i] = pivotWithCount{
			pos:   pivotPositions[i],
			count: 0,
		}
	}

	// merge the results from workers
	for i := 0; i < threadNum; i++ {
		subCounts := <-resultCh
		for i, c := range subCounts {
			pivots[i].count += c
		}
	}

	return pivots
}

// remarks: mergedPivots is sorted based on referencing value
func mergePivots(input []int, pivots []pivotWithCount, target int) (mergedPivots []pivotWithCount) {
	threhold := len(input) / target

	// merge the pivots
	mergedPivots = []pivotWithCount{}
	total := 0
	for _, obj := range pivots {
		total = total + obj.count
		if total >= threhold {
			p := pivotWithCount{pos: obj.pos, count: total}
			mergedPivots = append(mergedPivots, p)
			total = 0
		}
	}

	return mergedPivots
}

func relocatePivots(input []int, mergedPivots []pivotWithCount) (finalizedPivotPositions []int) {

	finalizedPivotPositions = make([]int, len(mergedPivots), len(mergedPivots))

	// example:
	// the content of pos[1,3,5,7,9] need to relocate to pos[3,6,9,12,15]
	// algorithm
	// 		1. backup the content of pos[3,6,9,12,15]
	// 		2. backup the content of pos[1,3,5,7,9]
	// 		3. using the backup in step 2, move the content of pos[1,3,5,7,9] to pos[3,6,9,12,15]
	// 		4. for backup in 1 and 2, remove the pos that exists in both backup
	//		5. fullback the pos[1,5,7] from the backup of pos[6, 12, 15] from step 4

	// storing the (newPos, newValue) pairs
	newKeyValues := map[int]int{}
	// storing the (oldPos, oldValue) pairs
	oldKeyValues := map[int]int{}

	// step 1+2
	total := 0
	for i, pivot := range mergedPivots {
		total += pivot.count

		oldPos, newPos := pivot.pos, total-1

		newKeyValues[newPos] = input[newPos]
		oldKeyValues[oldPos] = input[oldPos]

		finalizedPivotPositions[i] = newPos
	}
	// step 3
	total = 0
	for _, pivot := range mergedPivots {
		total += pivot.count

		oldPos, newPos := pivot.pos, total-1
		input[newPos] = oldKeyValues[oldPos]
	}
	// step 4
	newValues := []int{}
	for newPos, newValue := range newKeyValues {
		if _, ok := oldKeyValues[newPos]; ok {
			delete(oldKeyValues, newPos)
		} else {
			newValues = append(newValues, newValue)
		}
	}
	// step 5
	i := 0
	for oldPos := range oldKeyValues {
		input[oldPos] = newValues[i]
		i++
	}

	return finalizedPivotPositions
}

func bucketWorker(input, finalizedPivotPositions []int, exchangeChannels []chan int, startPos, endPos, workerIndex int, qsortWorkerCh chan []int) {
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

	log.Println(`workerIndex done`, workerIndex)

	// after the bucket is finished, pass the subslice to qsort for further processing
	qsortWorkerCh <- subSlice
}

func qsortWithBucket(input []int) {

	threadNum := runtime.NumCPU()

	// prepare the pivots, and then move the pivots to final location
	pivotPositions := getPivotPositions(input, threadNum*10)
	pivots := countBucketSize(input, pivotPositions)
	mergedPivots := mergePivots(input, pivots, threadNum*2)
	finalizedPivotPositions := relocatePivots(input, mergedPivots)
	pivotCount := len(finalizedPivotPositions)

	// create the exchange channels
	exchangeChannels := make([]chan int, pivotCount+1, pivotCount+1)
	for i := 0; i < len(exchangeChannels); i++ {
		exchangeChannels[i] = make(chan int, 100)
	}

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
