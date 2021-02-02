package qsortm

// perform single thread partitioning
func partitionSingleThread(data lessSwap, startPos, endPos, pivotPos int) (finalPivotPos int) {
	// swap the startPos with pivotPos first
	data.swap(startPos, pivotPos)
	pivotPos = startPos

	startIdx := startPos + 1
	endIdx := endPos - 1

	for {
		// scan for the swapping pairs
		for startIdx <= endIdx && data.less(pivotPos, startIdx) == false {
			startIdx++
		}
		for startIdx <= endIdx && data.less(pivotPos, endIdx) {
			endIdx--
		}

		if startIdx >= endIdx {
			break
		}
		// perform swapping
		data.swap(startIdx, endIdx)
		startIdx++
		endIdx--
	}

	// put back the pivot into correct position
	finalPivotPos = startIdx - 1
	data.swap(startPos, finalPivotPos)

	return finalPivotPos
}
