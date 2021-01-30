package experiment

import (
	"testing"
	//"log"
)

func getSampleInput() []int {
	return []int{10, 11, 12, 13, 14, 15, 1, 2, 3, 4, 5}
}

func verifyIntArray(t *testing.T, expectedInput, actualInput []int) {
	for i, v := range expectedInput {
		if actualInput[i] != v {
			t.Error("verifyIntArray not match", i, v, actualInput[i])
		}
	}
}

func TestSubTaskInternal1(t *testing.T) {
	input := getSampleInput()

	st := subtask{
		left:     sliceRange{start: 1, end: 6},
		right:    sliceRange{start: 6, end: 11},
		pivotPos: 0,
	}

	result := subTaskInternal(input, st)
	expected := []int{10, 5, 4, 3, 2, 1, 15, 14, 13, 12, 11}

	verifyIntArray(t, expected, input)
	if result.leftRemaining != 0 || result.rightRemaining != 0 {
		t.Error("leftRemaining / rightRemaining not match", result.leftRemaining, result.rightRemaining)
	}
}

func TestSubTaskInternal2(t *testing.T) {
	input := getSampleInput()

	st := subtask{
		left:     sliceRange{start: 1, end: 5},
		right:    sliceRange{start: 7, end: 11},
		pivotPos: 0,
	}

	result := subTaskInternal(input, st)
	expected := []int{10, 5, 4, 3, 2, 15, 1, 14, 13, 12, 11}

	verifyIntArray(t, expected, input)
	if result.leftRemaining != 0 || result.rightRemaining != 0 {
		t.Error("leftRemaining / rightRemaining not match", result.leftRemaining, result.rightRemaining)
	}
}

func TestSubTaskInternal3(t *testing.T) {
	input := getSampleInput()

	st := subtask{
		left:     sliceRange{start: 1, end: 5},
		right:    sliceRange{start: 8, end: 11},
		pivotPos: 0,
	}

	result := subTaskInternal(input, st)
	expected := []int{10, 5, 4, 3, 14, 15, 1, 2, 13, 12, 11}

	verifyIntArray(t, expected, input)
	if result.leftRemaining != 1 || result.rightRemaining != 0 {
		t.Error("leftRemaining / rightRemaining not match", result.leftRemaining, result.rightRemaining)
	}
}

func TestSubTaskInternal4(t *testing.T) {
	input := getSampleInput()

	st := subtask{
		left:     sliceRange{start: 1, end: 4},
		right:    sliceRange{start: 7, end: 11},
		pivotPos: 0,
	}

	result := subTaskInternal(input, st)
	expected := []int{10, 5, 4, 3, 14, 15, 1, 2, 13, 12, 11}

	verifyIntArray(t, expected, input)
	if result.leftRemaining != 0 || result.rightRemaining != 1 {
		t.Error("leftRemaining / rightRemaining not match", result.leftRemaining, result.rightRemaining)
	}
}
