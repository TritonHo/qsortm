package qsortm

import "reflect"

var reflectValueOf = reflect.ValueOf
var reflectSwapper = reflect.Swapper

type sliceWrapper struct {
	length int
	less   func(i, j int) bool
	swap   func(i, j int)
}

func (w sliceWrapper) Len() int {
	return w.length
}
func (w sliceWrapper) Swap(i, j int) {
	w.swap(i, j)
}
func (w sliceWrapper) Less(i, j int) bool {
	return w.less(i, j)
}

func Slice(slice interface{}, less func(i, j int) bool) {
	rv := reflectValueOf(slice)

	w := sliceWrapper{
		length: rv.Len(),
		swap:   reflectSwapper(slice),
		less:   less,
	}
	Sort(w)
}
