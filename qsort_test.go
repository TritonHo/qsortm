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

func TestQsortProd(t *testing.T) {
	array := generateRandomSlice(100000000)

	startTime := time.Now()
	qsortProd(array)

	log.Println("elapsed time:", time.Since(startTime))
	if isAscSorted(array) {
		t.Error("the sorting is buggy", array)
	}
}

func BenchmarkQsortProd(b *testing.B) {
	array := generateRandomSlice(1000000)
	qsortProd(array)
}

func BenchmarkQsortStandard(b *testing.B) {
	array := generateRandomSlice(1000000)
	sort.Ints(array)
}
