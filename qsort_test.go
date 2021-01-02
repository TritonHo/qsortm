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

func TestCountBucketSize(t *testing.T) {
	input := generateRandomSlice(100000)

	pivotPositions := getPivotPositions(input, 100)
	pivots := countBucketSize(input, pivotPositions)

	for _, v := range input {
		for i := range pivots {
			if v <= input[pivots[i].pos] {
				pivots[i].count--
				break
			}
		}
	}
	for i, p := range pivots {
		if p.count != 0 {
			t.Error("the count not match", i, p)
		}
	}
	log.Println("TestCountBucketSize finished")
}

func TestRelocatePivots(t *testing.T) {
	input := generateRandomSlice(10000)
	input2 := []int{}
	for _, value := range input {
		input2 = append(input2, value)
	}

	pivotPositions := getPivotPositions(input, 100)
	pivots := countBucketSize(input, pivotPositions)
	mergedPivots := mergePivots(input, pivots, 10)
	finalizedPivotPositions := relocatePivots(input, mergedPivots)

	// get the value of the pivots
	values := []int{}
	for _, pos := range finalizedPivotPositions {
		values = append(values, input[pos])
	}

	sort.Ints(input2)

	// if the finalizedPivotPositions is positioned corrected, it should remain in the same location
	for i, pos := range finalizedPivotPositions {
		if input2[pos] != values[i] {
			t.Error("the pivot position not match", input2[pos], values[i])
		}
	}

	log.Println(`TestRelocatePivots finished`)
}

func TestQsortWithBucket(t *testing.T) {
	array := generateRandomSlice(10000)

	startTime := time.Now()
	qsortWithBucket(array)

	log.Println("elapsed time:", time.Since(startTime))
	if isAscSorted(array) {
		t.Error("the sorting is buggy")
	}
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
