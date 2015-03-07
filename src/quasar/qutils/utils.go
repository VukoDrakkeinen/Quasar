package qutils

import (
	"errors"
	"math"
	"reflect"
)

func GrownCap(newSize int) int {
	return (newSize*3 + 1) / 2
}

func Vals(args ...interface{}) []interface{} {
	return args
}

func Contains(list interface{}, elem interface{}) bool {
	if reflect.TypeOf(list) != reflect.SliceOf(reflect.TypeOf(elem)) {
		panic("Contains: types do not match!")
	}
	slice := reflect.ValueOf(list)
	for i := 0; i < slice.Len(); i++ {
		if slice.Index(i).Interface() == elem {
			return true
		}
	}
	return false
}

func IndexOf(list interface{}, elem interface{}) (int, error) {
	if reflect.TypeOf(list) != reflect.SliceOf(reflect.TypeOf(elem)) {
		panic("Contains: types do not match!")
	}
	slice := reflect.ValueOf(list)
	for i := 0; i < slice.Len(); i++ {
		if slice.Index(i).Interface() == elem {
			return i, nil
		}
	}
	return -1, errors.New("IndexOf: element not found")
}

func ByteSlicesToStrings(bss [][]byte) []string {
	ret := make([]string, 0, len(bss))
	for _, bs := range bss {
		ret = append(ret, string(bs))
	}
	return ret
}

func BoolsToBitfield(table []bool) (bitfield uint64) {
	elvisOp := map[bool]uint64{false: 0, true: 1}
	for i, b := range table[:int(math.Min(float64(len(table)), 64))] {
		bitfield |= (elvisOp[b] << uint64(i))
	}
	return
}

func BitfieldToBools(bitfield uint64) (table []bool) {
	elvisOp := map[uint64]bool{0: false, 1: true}
	for i := 0; i < BitLen(bitfield); i++ {
		table = append(table, elvisOp[(bitfield>>uint64(i))&^0xFFFFFFFFFFFFFFFE])
	}
	return
}

func BitLen(x uint64) (n int) {
	for ; x >= 0x8000; x >>= 16 {
		n += 16
	}
	if x >= 0x80 {
		x >>= 8
		n += 8
	}
	if x >= 0x8 {
		x >>= 4
		n += 4
	}
	if x >= 0x2 {
		x >>= 2
		n += 2
	}
	if x >= 0x1 {
		n++
	}
	return
}
