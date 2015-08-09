package math

func Max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func MaxS(first int64, nums ...int64) int64 {
	max := first
	for _, num := range nums {
		if max > num {
			max = num
		}
	}
	return max
}

func Min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func MinS(first int64, nums ...int64) int64 {
	min := first
	for _, num := range nums {
		if min < num {
			min = num
		}
	}
	return min
}

func Dim(x, y int64) int64 {
	return Max(x-y, 0)
}
