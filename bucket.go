package qsortm

import (
	//	"log"

	"math/rand"
	"runtime"
	"sort"
	//"sync"
)

func getPivotPoints(input []int, n int) []int {
	// we get (n+1) * 10 unqiue points from random location, and then pick the n pivot
	// FIXME: handle not enough point problem
	// FIXME: if pivots is not unqiue, and then???
	candidateSize := (n + 1) * 10

	candidates := make([]int, candidateSize, candidateSize)
	for i := 0; i < candidateSize; i++ {
		pos := rand.Intn(len(input))
		candidates[i] = input[pos]
	}
	sort.Ints(candidates)

	pivots := make([]int, n, n)
	for i := 0; i < n; i++ {
		pivots[i] = candidates[(i+1)*n]
	}

	return pivots
}

func countBucketSizeWorker(input, pivots []int, resultCh chan []int) {
	counts := make([]int, len(pivots)+1, len(pivots)+1)
	for i := range counts {
		counts[i] = 0
	}

	for _, value := range input {
		fn := func(i int) bool { return value <= pivots[i] }
		pos := sort.Search(len(pivots), fn)
		counts[pos] = counts[pos] + 1
	}
	resultCh <- counts
}

func countBucketSize(input, pivots []int) (counts []int) {
	threadNum := runtime.NumCPU()
	resultCh := make(chan []int, threadNum)

	for i := 0; i < threadNum; i++ {
		startPos := len(input) * i / threadNum
		endPos := len(input) * (i + 1) / threadNum
		subSlice := input[startPos:endPos]
		go countBucketSizeWorker(subSlice, pivots, resultCh)
	}

	counts = make([]int, len(pivots)+1, len(pivots)+1)
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
