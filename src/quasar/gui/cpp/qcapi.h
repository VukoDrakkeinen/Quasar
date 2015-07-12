#ifndef QCAPI_H
#define QCAPI_H

#include "capi.h"

#ifdef __cplusplus
extern "C" {
#endif

typedef void QUpdateModel_;
typedef void QInfoModel_;
typedef void QChapterModel_;
typedef void GoComicList_;

QInfoModel_* newInfoModel(GoComicList_* data);
QUpdateModel_* newUpdateModel(GoComicList_* data);
QChapterModel_* newChapterModel(GoComicList_* data);

void* go_Offsets_ComicInfo;
void* go_Offsets_Scanlation;
void* go_Offsets_UpdateInfo;

//TODO: wrap all Go pointers in classes (RAII ftw)
//TODO: automatically include from _cgo_export.h
typedef unsigned char GoUint8;
typedef __SIZE_TYPE__ GoUintptr;
extern char* go_authorNameById(int);
extern char* go_artistNameById(int);
extern char* go_genreNameById(int);
extern char* go_categoryNameById(int);
extern char* go_scanlatorNameById(int);
extern char* go_langNameById(int);
extern char* go_getThumbnailPath(char*);
extern void* go_ComicList_GetComic(void*, int);
extern int   go_ComicList_Len(void*);
extern void* go_ComicList_ComicUpdateInfo(void*, int);
extern int   go_Comic_ChaptersCount(void*);
extern int   go_Comic_ChaptersReadCount(void*);
extern void* go_Comic_Info(void*);
extern void* go_Comic_GetChapter(void*, int);
extern GoUint8 go_Chapter_AlreadyRead(void*);
extern int   go_Chapter_ScanlationsCount(void*);
extern void* go_Chapter_GetScanlation(void*, int);
extern void  go_collectGarbage(void*);
extern GoUintptr go_JointScanlators_ToSlice(GoUintptr);

#ifdef __cplusplus
} // extern "C"
#endif

#endif // QCAPI_H