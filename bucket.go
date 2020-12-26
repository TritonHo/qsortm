package qsortm

import (
	//	"log"

	"math/rand"
	"runtime"
	"sort"
	"sync"
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
