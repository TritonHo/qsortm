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

func TestQsortProd(t *testing.T) {
	array := generateRandomSlice(1000000)
	counters := sliceToCounters(array)

	startTime := time.Now()
	qsortProd(array)

	log.Println("TestQsortProd elapsed time:", time.Since(startTime))
	if isAscSorted(array) == false {
		t.Error("the sorting is buggy, isAscSorted failed")
	}
	if verifySliceCounters(array, counters) == false {
		t.Error("the sorting is buggy, verifySliceCounters failed")
	}
}

func TestQsortProdV2(t *testing.T) {
	array := generateRandomSlice(1000000)
	counters := sliceToCounters(array)

	startTime := time.Now()
	qsortProdV2(array)

	log.Println("TestQsortProdV2 elapsed time:", time.Since(startTime))
	if isAscSorted(array) == false {
		t.Error("the sorting is buggy, isAscSorted failed")
	}
	if verifySliceCounters(array, counters) == false {
		t.Error("the sorting is buggy, verifySliceCounters failed")
	}
}

func BenchmarkQsortProd(b *testing.B) {
	array := generateRandomSlice(1000000)
	qsortProd(array)
}
func BenchmarkQsortProdV2(b *testing.B) {
	array := generateRandomSlice(1000000)
	qsortProdV2(array)
}

func BenchmarkQsortStandard(b *testing.B) {
	array := generateRandomSlice(1000000)
	sort.Ints(array)
}
