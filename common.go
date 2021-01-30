package qsortm

// perform single thread partitioning
func partitionSingle(input []int, startPos, endPos, pivotPos int) (finalPivotPos int) {
	// swap the startPos with pivotPos first
	input[startPos], input[pivotPos] = input[pivotPos], input[startPos]

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
	finalPivotPos = startIdx - 1
	input[startPos], input[finalPivotPos] = input[finalPivotPos], pivot
	return finalPivotPos
}
