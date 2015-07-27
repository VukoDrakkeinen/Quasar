#ifndef QCAPI_H
#define QCAPI_H

#ifdef __cplusplus
extern "C" {
#endif

typedef void UpdateModel_;
typedef void InfoModel_;
typedef void ChapterModel_;
typedef void GoComicList_;
typedef void NotifiableModel_;

InfoModel_* newInfoModel(GoComicList_* data);
UpdateModel_* newUpdateModel(GoComicList_* data);
ChapterModel_* newChapterModel(GoComicList_* data);
void registerQMLTypes();
void notifyModelInsertStart(NotifiableModel_* model, int row, int count);
void notifyModelInsertEnd(NotifiableModel_* model);
void notifyModelRemoveStart(NotifiableModel_* model, int row, int count);
void notifyModelRemoveEnd(NotifiableModel_* model);

void* go_Offsets_ComicInfo;
void* go_Offsets_Scanlation;
void* go_Offsets_UpdateInfo;

//TODO: wrap all Go pointers in classes (RAII ftw)
//TODO: automatically include from _cgo_export.h (how?)
typedef unsigned char GoUint8;
typedef __SIZE_TYPE__ GoUintptr;
typedef long long GoInt64;
typedef GoInt64 GoInt;
extern char* go_authorNameById(GoInt);
extern char* go_artistNameById(GoInt);
extern char* go_genreNameById(GoInt);
extern char* go_categoryNameById(GoInt);
extern char* go_scanlatorNameById(GoInt);
extern char* go_langNameById(GoInt);
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