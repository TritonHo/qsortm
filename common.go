package qsortm

// FIXME: fix the pivot selection
// take input[startPos] as pivot, and then run one iteration
func qsortPartition(input []int, startPos, endPos int) (pivotPos int) {
	pivot := input[startPos]

	startIdx := startPos + 1
	endIdx := endPos - 1

	for {
		// scan for the swapping pairs
		for startIdx <= endIdx && input[startIdx] <= pivot {
			startIdx++
		}
		for startIdx <= endIdx && input[endIdx] > pivot {
			endIdx--
		}

		if startIdx >= endIdx {
			break
		}
		// perform swapping
		input[startIdx], input[endIdx] = input[endIdx], input[startIdx]
	}

	// put back the pivot into correct position
	pivotPos = startIdx - 1
	input[startPos], input[pivotPos] = input[pivotPos], pivot
	return pivotPos
}
