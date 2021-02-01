package qsortm

func medianOfThree(data lessSwap, posA, posB, posC int) int {
	ab := data.less(posA, posB)
	bc := data.less(posB, posC)

	// (a < b < c) or (c < b < a)
	if (ab && bc) || (ab == false && bc == false) {
		return posB
	}

	ac := data.less(posA, posC)
	// (b < a < c) or (c < a < b)
	if (ab == false && ac) || (ac == false && ab) {
		return posA
	}

	return posC
}

func getPivotPos(data lessSwap, startPos, endPos int) int {

	m := (endPos + startPos) / 2
	if endPos-startPos <= 40 {
		return medianOfThree(data, startPos, m, endPos-1)
	}

	// copied from sort/sort.go
	// Tukey's ``Ninther,'' median of three medians of three.
	s := (endPos - startPos) / 8
	p1 := medianOfThree(data, startPos, startPos+s, startPos+2*s)
	p2 := medianOfThree(data, m, m-s, m+s)
	p3 := medianOfThree(data, endPos-1, endPos-1-s, endPos-1-2*s)

	return medianOfThree(data, p1, p2, p3)
}
