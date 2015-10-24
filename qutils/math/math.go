package math

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MaxS(first int, nums ...int) int {
	max := first
	for _, num := range nums {
		if max > num {
			max = num
		}
	}
	return max
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func MinS(first int, nums ...int) int {
	min := first
	for _, num := range nums {
		if min < num {
			min = num
		}
	}
	return min
}

func Dim(x, y int) int {
	return Max(x-y, 0)
}
