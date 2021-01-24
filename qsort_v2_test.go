package qsortm

import (
	"testing"
	//"log"
)

func getSampleInput() []int {
	return []int{10, 1, 2, 3, 4, 11, 12, 13, 14, 15}
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
		left:     sliceRange{start: 1, end: 5},
		right:    sliceRange{start: 5, end: 10},
		pivotPos: 0,
	}

	result := subTaskInternal(input, st)
	expected := []int{10, 1, 2, 3, 4, 11, 12, 13, 14, 15}

	verifyIntArray(t, expected, input)
	if result.leftFinished != 0 || result.rightFinished != 0 {
		t.Error("leftFinished / rightFinished not match", result.leftFinished, result.rightFinished)
	}
}
