package qsortm

import (
	"math/rand"
	"sort"
	"testing"

	"log"
	"time"
)

type intSlice []int

func (s intSlice) Len() int           { return len(s) }
func (s intSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s intSlice) Less(i, j int) bool { return s[i] < s[j] }

func generateRandomSlice(n int) intSlice {
	slice := make([]int, n, n)
	for i := 0; i < n; i++ {
		slice[i] = rand.Int()
	}
	return slice
}

func isAscSorted(slice intSlice) bool {
	for i := 1; i < len(slice); i++ {
		if slice[i-1] > slice[i] {
			return false
		}
	}
	return true
}

func sliceToCounters(slice intSlice) map[int]int {
	counters := map[int]int{}
	for _, item := range slice {
		counters[item] = counters[item] + 1
	}
	return counters
}

func verifySliceCounters(slice intSlice, counters map[int]int) bool {
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

	log.Println("TestSort elapsed time:", time.Since(startTime))
	if isAscSorted(array) == false {
		t.Error("the sorting is buggy, isAscSorted failed")
	}
	if verifySliceCounters(array, counters) == false {
		t.Error("the sorting is buggy, verifySliceCounters failed")
	}
}

func TestSlice(t *testing.T) {
	array := generateRandomSlice(1000000)
	counters := sliceToCounters(array)

	startTime := time.Now()
	lessFn := func(i, j int) bool { return array[i] < array[j] }
	Slice(array, lessFn)

	log.Println("TestSlice elapsed time:", time.Since(startTime))
	if isAscSorted(array) == false {
		t.Error("the sorting is buggy, isAscSorted failed")
	}
	if verifySliceCounters(array, counters) == false {
		t.Error("the sorting is buggy, verifySliceCounters failed")
	}
}

func BenchmarkQsortmSort(b *testing.B) {
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		array := generateRandomSlice(1000000)
		b.StartTimer()

		Sort(array)
	}
}
func BenchmarkQsortmSlice(b *testing.B) {
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		array := generateRandomSlice(1000000)
		b.StartTimer()

		lessFn := func(i, j int) bool { return array[i] < array[j] }
		Slice(array, lessFn)
	}
}

func BenchmarkStandardSort(b *testing.B) {
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		array := generateRandomSlice(1000000)
		b.StartTimer()

		sort.Sort(array)
	}
}
func BenchmarkStandardSlice(b *testing.B) {

	for n := 0; n < b.N; n++ {
		b.StopTimer()
		array := generateRandomSlice(1000000)
		b.StartTimer()

		lessFn := func(i, j int) bool { return array[i] < array[j] }
		sort.Slice(array, lessFn)
	}
}
