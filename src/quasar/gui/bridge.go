package gui

import "C"

import (
	"quasar/redshift/idsdict"
	"quasar/redshift/qdb"
)

//export authorNameById
func authorNameById(id int) *C.char {
	var gid idsdict.AuthorId
	gid.Scan(int64(id + 1))
	return C.CString(idsdict.Authors.NameOf(gid))
}

//export artistNameById
func artistNameById(id int) *C.char {
	var gid idsdict.ArtistId
	gid.Scan(int64(id + 1))
	return C.CString(idsdict.Artists.NameOf(gid))
}

//export genreNameById
func genreNameById(id int) *C.char {
	var gid idsdict.ComicGenreId
	gid.Scan(int64(id + 1))
	return C.CString(idsdict.ComicGenres.NameOf(gid))
}

//export categoryNameById
func categoryNameById(id int) *C.char {
	var gid idsdict.ComicTagId
	gid.Scan(int64(id + 1))
	return C.CString(idsdict.ComicTags.NameOf(gid))
}

//export getThumbnailPath
func getThumbnailPath(str *C.char) *C.char {
	return C.CString(qdb.GetThumbnailPath(C.GoString(str)))
}
