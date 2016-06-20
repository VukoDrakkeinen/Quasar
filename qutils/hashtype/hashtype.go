package hashtype

import (
	"fmt"
	"hash/fnv"
	"reflect"
	"unsafe"
)

func Struct(s interface{}) uint64 {
	t := reflect.TypeOf(s)
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("HashStructType: Given type %v is not of struct kind", t))
	}
	hasher := fnv.New64a()
	fields := typeFields(t)
	numField := len(fields)
	hasher.Write(unsafeBytes(&numField))
	for i, field := range fields {
		fieldOffs := field.Offset
		fieldType := field.Type
		fieldKind := fieldType.Kind()
		fieldSize := fieldType.Size()
		hasher.Write(unsafeBytes(&i))
		hasher.Write(unsafeBytes(&fieldOffs))
		hasher.Write(unsafeBytes(&fieldKind))
		hasher.Write(unsafeBytes(&fieldSize))
	}
	hash := hasher.Sum(nil)
	return *(*uint64)(unsafe.Pointer(&hash[0]))
}

func unsafeBytes(ptr interface{}) []byte {
	v := reflect.ValueOf(ptr)
	b := reflect.SliceHeader{
		Data: v.Pointer(),
		Len:  int(v.Type().Size()),
	}
	b.Cap = b.Len
	return *(*[]byte)(unsafe.Pointer(&b))
}

func typeFields(t reflect.Type) (fields []reflect.StructField) {
	type fieldScan struct {
		t   reflect.Type
		idx []int
	}
	currentLvl := []fieldScan{}    //Holy hell, thanks, Go Team, for teaching me this straightforward way
	nextLvl := []fieldScan{{t: t}} //to transform certain cases of recursive code into iterative! -Vuko

	names := make(map[string]struct{})
	if t.Kind() == reflect.Struct {
		fields = make([]reflect.StructField, 0, t.NumField()*3/2)
	}

	for len(nextLvl) > 0 {
		currentLvl, nextLvl = nextLvl, currentLvl[:0]
		for _, scan := range currentLvl {
			t := scan.t
			if t.Kind() != reflect.Struct {
				continue
			}

			numField := t.NumField()
			for i := 0; i < numField; i++ {
				f := t.Field(i)
				if _, shadowed := names[f.Name]; shadowed {
					continue
				}
				names[f.Name] = struct{}{}

				idx := append(append(make([]int, 0, len(scan.idx)+1), scan.idx...), i)
				f.Index = idx
				fields = append(fields, f)
				if f.Anonymous {
					nextLvl = append(nextLvl, fieldScan{t: f.Type, idx: idx})
				}
			}
		}
	}
	return fields
}
