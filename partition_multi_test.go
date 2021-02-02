package qsortm

import (
	"testing"
	//"log"
)

func verifyIntArray(t *testing.T, expectedInput, actualInput []int) {
	for i, v := range expectedInput {
		if actualInput[i] != v {
			t.Error("verifyIntArray not match", i, v, actualInput[i])
		}
	}
}

func getSampleInput() (intSlice, lessSwap) {
	slice := []int{10, 11, 12, 13, 14, 15, 1, 2, 3, 4, 5}

	temp := lessSwap{
		length: len(slice),
		less:   func(i, j int) bool { return slice[i] < slice[j] },
		swap:   func(i, j int) { slice[i], slice[j] = slice[j], slice[i] },
	}

	return slice, temp
}

func TestSwappingOnBlock1(t *testing.T) {
	intArray, data := getSampleInput()

	left := sliceRange{start: 1, end: 6}
	right := sliceRange{start: 6, end: 11}
	pivotPos := 0

	leftRemaining, rightRemaining := swappingOnBlock(data, left, right, pivotPos)
	expected := []int{10, 5, 4, 3, 2, 1, 15, 14, 13, 12, 11}

	verifyIntArray(t, expected, intArray)
	if leftRemaining != 0 || rightRemaining != 0 {
		t.Error("leftRemaining / rightRemaining not match", leftRemaining, rightRemaining)
	}
}

func TestSwappingOnBlock2(t *testing.T) {
	intArray, data := getSampleInput()

	left := sliceRange{start: 1, end: 5}
	right := sliceRange{start: 7, end: 11}
	pivotPos := 0

	leftRemaining, rightRemaining := swappingOnBlock(data, left, right, pivotPos)
	expected := []int{10, 5, 4, 3, 2, 15, 1, 14, 13, 12, 11}

	verifyIntArray(t, expected, intArray)
	if leftRemaining != 0 || rightRemaining != 0 {
		t.Error("leftRemaining / rightRemaining not match", leftRemaining, rightRemaining)
	}
}

func TestSwappingOnBlock3(t *testing.T) {
	intArray, data := getSampleInput()

	left := sliceRange{start: 1, end: 5}
	right := sliceRange{start: 8, end: 11}
	pivotPos := 0

	leftRemaining, rightRemaining := swappingOnBlock(data, left, right, pivotPos)
	expected := []int{10, 5, 4, 3, 14, 15, 1, 2, 13, 12, 11}

	verifyIntArray(t, expected, intArray)
	if leftRemaining != 1 || rightRemaining != 0 {
		t.Error("leftRemaining / rightRemaining not match", leftRemaining, rightRemaining)
	}
}

func TestSwappingOnBlock4(t *testing.T) {
	intArray, data := getSampleInput()

	left := sliceRange{start: 1, end: 4}
	right := sliceRange{start: 7, end: 11}
	pivotPos := 0

	leftRemaining, rightRemaining := swappingOnBlock(data, left, right, pivotPos)
	expected := []int{10, 5, 4, 3, 14, 15, 1, 2, 13, 12, 11}

	verifyIntArray(t, expected, intArray)
	if leftRemaining != 0 || rightRemaining != 1 {
		t.Error("leftRemaining / rightRemaining not match", leftRemaining, rightRemaining)
	}
}
