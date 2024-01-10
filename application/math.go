package application

import "cmp"

func Clamp[T cmp.Ordered](a T, minValue T, maxValue T) T {
	return min(max(a, minValue), maxValue)
}
