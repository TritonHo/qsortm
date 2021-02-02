package qsortm

import "reflect"

var reflectValueOf = reflect.ValueOf
var reflectSwapper = reflect.Swapper

type lessSwap struct {
	length int
	less   func(i, j int) bool
	swap   func(i, j int)
}

func Slice(slice interface{}, less func(i, j int) bool) {
	rv := reflectValueOf(slice)

	temp := lessSwap{
		length: rv.Len(),
		swap:   reflectSwapper(slice),
		less:   less,
	}
	quicksort(temp)
}
