package core

import "fmt"

const (
	_NotificationMode_name = "ImmediateAccumulativeDelayed"
	_comicType_name        = "InvalidComicMangaManhwaManhuaWesternWebcomicOther"
	_comicStatus_name      = "ComicStatusInvalidComicCompleteComicOngoingComicOnHiatusComicDiscontinued"
)

var (
	_NotificationMode_index = [...]uint8{9, 21, 28}
	_comicType_index        = [...]uint8{12, 17, 23, 29, 36, 44, 49}
	_comicStatus_index      = [...]uint8{18, 31, 43, 56, 73}
)

func (i NotificationMode) String() string {
	if i < 0 || i >= NotificationMode(len(_NotificationMode_index)) {
		return fmt.Sprintf("NotificationMode(%d)", i)
	}
	hi := _NotificationMode_index[i]
	lo := uint8(0)
	if i > 0 {
		lo = _NotificationMode_index[i-1]
	}
	return _NotificationMode_name[lo:hi]
}

func NotificationModeFromString(s string) NotificationMode {
	for i := 0; i < len(_NotificationMode_index); i++ {
		u := NotificationMode(i)
		if u.String() == s {
			return u
		}
	}
	return NotificationMode(0)
}

func NotificationModeValueNames() (names []string) {
	for i := 0; i < len(_NotificationMode_index); i++ {
		u := NotificationMode(i)
		names = append(names, u.String())
	}
	return
}

func (i comicType) String() string {
	if i < 0 || i >= comicType(len(_comicType_index)) {
		return fmt.Sprintf("comicType(%d)", i)
	}
	hi := _comicType_index[i]
	lo := uint8(0)
	if i > 0 {
		lo = _comicType_index[i-1]
	}
	return _comicType_name[lo:hi]
}

func (i comicStatus) String() string {
	if i < 0 || i >= comicStatus(len(_comicStatus_index)) {
		return fmt.Sprintf("comicStatus(%d)", i)
	}
	hi := _comicStatus_index[i]
	lo := uint8(0)
	if i > 0 {
		lo = _comicStatus_index[i-1]
	}
	return _comicStatus_name[lo:hi]
}
