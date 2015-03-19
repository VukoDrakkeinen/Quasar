package redshift

import "fmt"

const (
	_UpdateNotificationMode_name = "OnLaunchAccumulativeDelayedManual"
	_comicType_name              = "InvalidComicMangaManhwaManhuaWesternWebcomicOther"
	_comicStatus_name            = "ComicStatusInvalidComicCompleteComicOngoingComicOnHiatusComicDiscontinued"
)

var (
	_UpdateNotificationMode_index = [...]uint8{8, 20, 27, 33}
	_comicType_index              = [...]uint8{12, 17, 23, 29, 36, 44, 49}
	_comicStatus_index            = [...]uint8{18, 31, 43, 56, 73}
)

func (i UpdateNotificationMode) String() string {
	if i < 0 || i >= UpdateNotificationMode(len(_UpdateNotificationMode_index)) {
		return fmt.Sprintf("UpdateNotificationMode(%d)", i)
	}
	hi := _UpdateNotificationMode_index[i]
	lo := uint8(0)
	if i > 0 {
		lo = _UpdateNotificationMode_index[i-1]
	}
	return _UpdateNotificationMode_name[lo:hi]
}

func UpdateNotificationModeFromString(s string) UpdateNotificationMode {
	for i := 0; i < len(_UpdateNotificationMode_index); i++ {
		u := UpdateNotificationMode(i)
		if u.String() == s {
			return u
		}
	}
	return UpdateNotificationMode(0)
}

func UpdateNotificationModeValueNames() (names []string) {
	for i := 0; i < len(_UpdateNotificationMode_index); i++ {
		u := UpdateNotificationMode(i)
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
