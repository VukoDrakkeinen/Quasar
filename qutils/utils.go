package qutils

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/VukoDrakkeinen/Quasar/qregexp"
	"github.com/VukoDrakkeinen/Quasar/qutils/math"
)

func GrownCap(newSize int) int {
	return (newSize*3 + 1) / 2
}

func Contains(list interface{}, elem interface{}) bool {
	//indexableTypeAssert(list, elem, "Contains")
	slice := reflect.ValueOf(list)
	sliceLen := slice.Len()
	for i := 0; i < sliceLen; i++ {
		if slice.Index(i).Interface() == elem {
			return true
		}
	}
	return false
}

func IndexOf(list interface{}, elem interface{}) (int, error) {
	//indexableTypeAssert(list, elem, "IndexOf")
	slice := reflect.ValueOf(list)
	sliceLen := slice.Len()
	for i := 0; i < sliceLen; i++ {
		if slice.Index(i).Interface() == elem {
			return i, nil
		}
	}
	return -1, errors.New("IndexOf: element not found")
}

func indexableTypeAssert(list interface{}, elem interface{}, funcName string) {
	listType := reflect.TypeOf(list)
	switch listType.Kind() {
	case reflect.Array, reflect.Slice: //continue
	default:
		panic(funcName + ": list type is not indexable!")
	}
	if listType.Elem().Kind() != reflect.TypeOf(elem).Kind() {
		panic(funcName + ": types do not match!")
	}
}

func AppendUnique(list interface{}, elems ...interface{}) (newList interface{}) {
	slice := reflect.ValueOf(list)
	for _, elem := range elems {
		if elemsSlice := reflect.ValueOf(elem); elemsSlice.Kind() == reflect.Slice {
			for slen, i := elemsSlice.Len(), 0; i < slen; i++ {
				elem := elemsSlice.Index(i)
				if !Contains(list, elem.Interface()) {
					slice = reflect.Append(slice, elem)
				}
			}
		} else if !Contains(list, elem) {
			slice = reflect.Append(slice, reflect.ValueOf(elem))
		}
	}
	return slice.Interface()
}

func ByteSlicesToStrings(bss [][]byte) []string {
	ret := make([]string, 0, len(bss))
	for _, bs := range bss {
		ret = append(ret, string(bs))
	}
	return ret
}

func BoolsToBitfield(table []bool) (bitfield uint64) {
	if len(table) > 64 {
		panic("BoolsToBitfield: provided bool table is too long!")
	}
	elvisOp := map[bool]uint64{false: 0, true: 1}
	for i, b := range table {
		bitfield |= (elvisOp[b] << uint64(i))
	}
	return
}

func BitfieldToBools(bitfield uint64, expectedLength int) (table []bool) {
	elvisOp := map[uint64]bool{0: false, 1: true}
	bitLength := BitLen(bitfield)
	table = make([]bool, math.Max(expectedLength, bitLength))
	for i := 0; i < bitLength; i++ {
		table[i] = elvisOp[(bitfield>>uint64(i))&1]
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

var stackRegexp = qregexp.MustCompile(`(?<=runtime/panic.go:\d+ \(0x.{6}\)\n.+\n)(?s:.+)(?=\n.+runtime/asm)`)

func Stack() string {
	stack := debug.Stack()
	return string(stackRegexp.Find(stack))
}

func Reverse(data sliceWrapper) {
	for min, max := 0, data.Len()-1; min < max; min, max = min+1, max-1 {
		data.Swap(min, max)
	}
}

type sliceWrapper interface {
	Len() int
	Swap(i, j int)
}

const bufLen = 512

func BackgroundCopy(r io.Reader, w io.Writer) (copiedChan <-chan int, errChan <-chan error) {
	copied_ := make(chan int, 10)
	err_ := make(chan error, 1)

	go func() {
		defer func() {
			close(err_)
		}()

		var copied int

		if wb, ok := w.(*bytes.Buffer); ok {
			buffer := wb.Bytes()

			cycleStart := time.Now()
			for {
				n, err := r.Read(buffer[copied:math.Min(copied+bufLen, cap(buffer))])
				if n == 0 {
					if err != io.EOF {
						err_ <- err
					}
					return
				}

				copied += n
				if time.Now().Sub(cycleStart) > (32 * time.Millisecond) { //TODO: rethink this
					select {
					case copied_ <- copied:
					default:
					}
					cycleStart = time.Now()
				}

				if copied == cap(buffer) {
					select {
					case copied_ <- copied:
					default:
					}
					return
				}
			}
		} else {
			buffer := make([]byte, bufLen)

			cycleStart := time.Now()
			for {
				n, err := r.Read(buffer)
				if n == 0 {
					if err != io.EOF {
						err_ <- err
					}
					return
				}

				_, err = w.Write(buffer[:n])
				if err != nil {
					err_ <- err
					return
				}

				copied += n
				if time.Now().Sub(cycleStart) > (32 * time.Millisecond) { //TODO: rethink this
					select {
					case copied_ <- copied:
					default:
					}
					cycleStart = time.Now()
				}
			}
		}
	}()

	return copied_, err_
}
