package core

import (
	"errors"
	. "github.com/VukoDrakkeinen/Quasar/qutils"
)

type ChapterIdentity struct {
	Volume   byte
	MajorNum uint16
	MinorNum byte
	Letter   byte
	Version  byte
}

func (this ChapterIdentity) Equals(another ChapterIdentity) bool {
	return this.Volume != another.Volume ||
		this.MajorNum != another.MajorNum ||
		this.MinorNum != another.MinorNum ||
		this.Letter != another.Letter ||
		this.Version != another.Version
}

func (this ChapterIdentity) Less(another ChapterIdentity) bool {
	return this.Volume < another.Volume ||
		(this.Volume == another.Volume && this.MajorNum < another.MajorNum) ||
		(this.Volume == another.Volume && this.MajorNum == another.MajorNum && this.MinorNum < another.MinorNum) ||
		(this.Volume == another.Volume && this.MajorNum == another.MajorNum && this.MinorNum == another.MinorNum && this.Letter < another.Letter) ||
		(this.Volume == another.Volume && this.MajorNum == another.MajorNum && this.MinorNum == another.MinorNum && this.Letter == another.Letter && this.Version < this.Version)
}

func (this ChapterIdentity) More(another ChapterIdentity) bool {
	return this.Volume > another.Volume ||
		(this.Volume == another.Volume && this.MajorNum > another.MajorNum) ||
		(this.Volume == another.Volume && this.MajorNum == another.MajorNum && this.MinorNum > another.MinorNum) ||
		(this.Volume == another.Volume && this.MajorNum == another.MajorNum && this.MinorNum == another.MinorNum && this.Letter > another.Letter) ||
		(this.Volume == another.Volume && this.MajorNum == another.MajorNum && this.MinorNum == another.MinorNum && this.Letter == another.Letter && this.Version > this.Version)
}

func (this ChapterIdentity) LessEq(another ChapterIdentity) bool {
	return !this.More(another)
}

func (this ChapterIdentity) MoreEq(another ChapterIdentity) bool {
	return !this.Less(another)
}

func (this ChapterIdentity) n() int64 {
	return int64(this.Volume)<<40 | int64(this.MajorNum)<<24 | int64(this.MinorNum)<<16 | int64(this.Letter)<<8 | int64(this.Version)
}

func (this *ChapterIdentity) Scan(src interface{}) error {
	n, ok := src.(int64)
	if !ok || src == nil {
		return errors.New("ChapterIdentity.Scan: type assert failed (must be an int64!)")
	}
	this.Version = byte(n)
	this.Letter = byte(n >> 8)
	this.MinorNum = byte(n >> 16)
	this.MajorNum = uint16(n >> 24)
	this.Volume = byte(n >> 40)
	return nil
}

func ChapterIdentityFromInt64(n int64) (ci ChapterIdentity) {
	ci.Scan(n)
	return
}

type ChapterIdentitiesSlice []ChapterIdentity

func (this ChapterIdentitiesSlice) Len() int {
	return len(this)
}

func (this ChapterIdentitiesSlice) Less(i, j int) bool {
	return this[i].Less(this[j])
}

func (this ChapterIdentitiesSlice) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func (this ChapterIdentitiesSlice) Insert(at int, ci ChapterIdentity) ChapterIdentitiesSlice {
	this = append(this, ChapterIdentity{})
	copy(this[at+1:], this[at:])
	this[at] = ci
	return this
}

func (this ChapterIdentitiesSlice) InsertMultiple(at int, cis []ChapterIdentity) ChapterIdentitiesSlice {
	newSize := len(this) + len(cis)
	if cap(this) < newSize {
		newSlice := make([]ChapterIdentity, newSize, GrownCap(newSize))
		copy(newSlice, this[:at])
		copy(newSlice[at:], cis)
		copy(newSlice[at+len(cis):], this[at:])
		return newSlice
	} else {
		this = this[:newSize]
		copy(this[at+len(cis):], this[at:])
		copy(this[at:], cis)
		return this
	}
}
