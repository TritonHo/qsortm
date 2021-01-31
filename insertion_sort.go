package qsortm

// copied from sort/sort.go
// Insertion sort
// FIXME: implement the shell sort instead
func insertionSort(data Interface, startPos, endPos int) {
	for i := startPos + 1; i < endPos; i++ {
		for j := i; j > startPos && data.Less(j, j-1); j-- {
			data.Swap(j, j-1)
		}
	}
}
