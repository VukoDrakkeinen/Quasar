package core

import (
	"math/big"
)

func (this ChapterIdentitiesSlice) vestedIndexOf(ci ChapterIdentity) (index int) {
	low := int64(0)
	mid := int64(0)
	high := int64(0)
	if this.Len() > 0 {
		high = int64(this.Len() - 1)
	}
	if high == 0 || ci.n() > this[high].n() {
		return this.Len()
	}
	var diff0, diff1, diff2 big.Int
	for this[low].n() <= ci.n() && this[high].n() >= ci.n() {
		//The following four lines are equivalent to
		//mid = low + ((ci.n()-this[low].n())*(high-low))/(this[high].n()-this[low].n())
		//(except the longer implementation doesn't overflow int64)
		diff0.SetInt64(ci.n() - this[low].n())
		diff1.SetInt64(high - low)
		diff2.SetInt64(this[high].n() - this[low].n())
		mid = low + diff0.Mul(&diff0, &diff1).Div(&diff0, &diff2).Int64()
		if this[mid].n() < ci.n() {
			low = mid + 1
		} else if this[mid].n() > ci.n() {
			high = mid - 1
		} else {
			return int(mid)
		}
	}

	if ci.n() > this[low].n() {
		return int(mid)
	} else {
		return int(low)
	}
}
