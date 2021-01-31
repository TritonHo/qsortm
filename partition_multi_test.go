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

func TestSwappingOnBlock1(t *testing.T) {
	input := getSampleInput()

	left := sliceRange{start: 1, end: 6}
	right := sliceRange{start: 6, end: 11}
	pivotPos := 0

	leftRemaining, rightRemaining := swappingOnBlock(input, left, right, pivotPos)
	expected := []int{10, 5, 4, 3, 2, 1, 15, 14, 13, 12, 11}

	verifyIntArray(t, expected, input)
	if leftRemaining != 0 || rightRemaining != 0 {
		t.Error("leftRemaining / rightRemaining not match", leftRemaining, rightRemaining)
	}
}

func TestSwappingOnBlock2(t *testing.T) {
	input := getSampleInput()

	left := sliceRange{start: 1, end: 5}
	right := sliceRange{start: 7, end: 11}
	pivotPos := 0

	leftRemaining, rightRemaining := swappingOnBlock(input, left, right, pivotPos)
	expected := []int{10, 5, 4, 3, 2, 15, 1, 14, 13, 12, 11}

	verifyIntArray(t, expected, input)
	if leftRemaining != 0 || rightRemaining != 0 {
		t.Error("leftRemaining / rightRemaining not match", leftRemaining, rightRemaining)
	}
}

func TestSwappingOnBlock3(t *testing.T) {
	input := getSampleInput()

	left := sliceRange{start: 1, end: 5}
	right := sliceRange{start: 8, end: 11}
	pivotPos := 0

	leftRemaining, rightRemaining := swappingOnBlock(input, left, right, pivotPos)
	expected := []int{10, 5, 4, 3, 14, 15, 1, 2, 13, 12, 11}

	verifyIntArray(t, expected, input)
	if leftRemaining != 1 || rightRemaining != 0 {
		t.Error("leftRemaining / rightRemaining not match", leftRemaining, rightRemaining)
	}
}

func TestSwappingOnBlock4(t *testing.T) {
	input := getSampleInput()

	left := sliceRange{start: 1, end: 4}
	right := sliceRange{start: 7, end: 11}
	pivotPos := 0

	leftRemaining, rightRemaining := swappingOnBlock(input, left, right, pivotPos)
	expected := []int{10, 5, 4, 3, 14, 15, 1, 2, 13, 12, 11}

	verifyIntArray(t, expected, input)
	if leftRemaining != 0 || rightRemaining != 1 {
		t.Error("leftRemaining / rightRemaining not match", leftRemaining, rightRemaining)
	}
}
