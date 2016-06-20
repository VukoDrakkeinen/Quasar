package core

import "github.com/VukoDrakkeinen/Quasar/qutils"

func (this ChapterIdentitiesSlice) fittingIndexOf(ci ChapterIdentity) (index int) {
	low := int64(0)
	mid := int64(0)
	high := int64(0)
	if this.Len() > 0 {
		high = int64(this.Len() - 1)
	}
	if high == 0 || ci.n() > this[high].n() {
		return this.Len()
	}
	for this[low].n() <= ci.n() && this[high].n() >= ci.n() {
		//The following is equivalent to
		//mid = low + ((ci.n()-this[low].n())*(high-low))/(this[high].n()-this[low].n())
		//except it doesn't overflow int64
		mid = low + qutils.MultiplyThenDivide(ci.n()-this[low].n(), high-low, this[high].n()-this[low].n())
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
