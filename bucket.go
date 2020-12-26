package qsortm

import (
	//	"log"

	"math/rand"
	"runtime"
	"sort"
	//"sync"
)

func getPivotPoints(input []int, n int) (pivotPositions []int) {
	// we get (n+1) * 10 unqiue points from random location, and then pick the n pivot
	// FIXME: handle not enough point problem
	// FIXME: if pivots is not unqiue, and then???
	candidateSize := (n + 1) * 10

	candidates := make([]int, candidateSize, candidateSize)
	m := map[int]bool{}
	for i := 0; i < candidateSize; i++ {
		for {
			pos := rand.Intn(len(input))
			if _, ok := m[pos]; !ok {
				m[pos] = true
				candidates[i] = pos
			}
		}
	}

	// the the candidates according to the value
	lessFn := func(i, j int) bool {
		v1 := input[pivotPositions[i]]
		v2 := input[pivotPositions[j]]
		return v1 < v2
	}
	sort.Slice(candidates, lessFn)

	pivots := make([]int, n, n)
	for i := 0; i < n; i++ {
		pivots[i] = candidates[(i+1)*n]
	}

	return pivots
}

func countBucketSizeWorker(input, subSlice, pivotPositions []int, resultCh chan []int) {
	counts := make([]int, len(pivotPositions)+1, len(pivotPositions)+1)
	for i := range counts {
		counts[i] = 0
	}

	for _, value := range subSlice {
		fn := func(i int) bool { return value <= input[pivotPositions[i]] }
		bucketIndex := sort.Search(len(pivotPositions), fn)
		counts[bucketIndex] = counts[bucketIndex] + 1
	}
	resultCh <- counts
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

	counts = make([]int, len(pivotPositions)+1, len(pivotPositions)+1)
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
