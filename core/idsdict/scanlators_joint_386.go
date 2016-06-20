package idsdict

import (
	"fmt"
	"sort"
	"unsafe"
)

const intSize = 4

func init() {
	if idSize := unsafe.Sizeof(Id(0)); idSize != intSize {
		panic(fmt.Sprintf("Id type has invalid size of %v, expected 4!", idSize))
	}
}

func JoinScanlators(ids []ScanlatorId) JointScanlatorIds {
	sort.Sort(ScanlatorSlice(ids))
	bytes := make([]byte, 0, len(ids)*intSize)
	for _, id := range ids {
		bytes = append(bytes, byte(id.ordinal), byte(id.ordinal>>8), byte(id.ordinal>>16), byte(id.ordinal>>24))
	}
	return JointScanlatorIds(*(*string)(unsafe.Pointer(&bytes)))

}

func (this *JointScanlatorIds) Slice() []ScanlatorId {
	ids := make([]ScanlatorId, 0, len(*this))
	var id Id
	for i, jbyte := range *(*[]byte)(unsafe.Pointer(this)) {
		switch i % intSize {
		case 0:
			id = Id(jbyte)
		case 1, 2:
			id |= Id(jbyte) << (8 * uint(i))
		case 3:
			id |= Id(jbyte) << (8 * uint(i))
			ids = append(ids, ScanlatorId{id})
		}
	}
	return ids
}
