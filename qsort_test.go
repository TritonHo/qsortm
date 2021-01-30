package qsortm

import (
	"math/rand"
	"sort"
	"testing"

	"log"
	"time"
)

func generateRandomSlice(n int) []int {
	slice := make([]int, n, n)
	for i := 0; i < n; i++ {
		slice[i] = rand.Int()
	}
	return slice
}

func isAscSorted(slice []int) bool {
	for i := 1; i < len(slice); i++ {
		if slice[i-1] > slice[i] {
			return false
		}
	}
	return true
}

func sliceToCounters(slice []int) map[int]int {
	counters := map[int]int{}
	for _, item := range slice {
		counters[item] = counters[item] + 1
	}
	return counters
}

func verifySliceCounters(slice []int, counters map[int]int) bool {
	for _, item := range slice {
		counters[item] = counters[item] - 1
	}
	for _, v := range counters {
		if v != 0 {
			return false
		}
	}
	return true
}

func TestSort(t *testing.T) {
	array := generateRandomSlice(1000000)
	counters := sliceToCounters(array)

	startTime := time.Now()
	Sort(array)

	log.Println("TestQsortProdV2 elapsed time:", time.Since(startTime))
	if isAscSorted(array) == false {
		t.Error("the sorting is buggy, isAscSorted failed")
	}
	if verifySliceCounters(array, counters) == false {
		t.Error("the sorting is buggy, verifySliceCounters failed")
	}
}

func BenchmarkQsortm(b *testing.B) {
	array := generateRandomSlice(1000000)
	Sort(array)
}

func BenchmarkStandard(b *testing.B) {
	array := generateRandomSlice(1000000)
	sort.Ints(array)
}
