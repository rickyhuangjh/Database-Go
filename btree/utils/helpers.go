package utils


func Insert[T any](slice []T, idx int, val T) []T {
	if idx < 0 || idx > len(slice) {
		panic("Slice insert idx out of bounds")
	}
	if idx == cap(slice) {
		panic("Slice over capacity")
	}
	var dummy T
	slice = append(slice, dummy)
	copy(slice[idx+1:], slice[idx:])
	slice[idx] = val
	return slice
}

func Delete[T any](slice []T, idx int) []T {
	if idx < 0 || idx >= len(slice) {
		panic("Slice delete idx out of bounds")
	}
	return append(slice[:idx], slice[idx+1:]...)
}
