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
typedef void WrappedModel_;
typedef void ComicInfoRow_;
typedef void ScanlationRow_;
typedef void UpdateInfoRow_;

InfoModel_* newInfoModel(GoComicList_* data);
UpdateModel_* newUpdateModel(GoComicList_* data);
ChapterModel_* newChapterModel(GoComicList_* data);
WrappedModel_* wrapModel(NotifiableModel_* model);  //wrap in SortFilter proxy
void registerQMLTypes();
int assertSyncedHashes();
ComicInfoRow_* convertComicInfo(void* info);
ScanlationRow_* convertScanlation(void* scanlation, void* scanlatorsPtr);
UpdateInfoRow_* convertUpdateInfo(void* updateInfo);
void modelSetGoData(NotifiableModel_* model, void* goData);
void notifyModelInsertStart(NotifiableModel_* model, int row, int count);
void notifyModelInsertEnd(NotifiableModel_* model);
void notifyModelRemoveStart(NotifiableModel_* model, int row, int count);
void notifyModelRemoveEnd(NotifiableModel_* model);
void notifyModelResetStart(NotifiableModel_* model);
void notifyModelResetEnd(NotifiableModel_* model);
void notifyModelDataChanged(NotifiableModel_* model, int row, int count, int column);

void* go_Offsets_ComicInfo;
void* go_Offsets_Scanlation;
void* go_Offsets_UpdateInfo;
unsigned long long go_Hash_ComicInfo;
unsigned long long go_Hash_Scanlation;
unsigned long long go_Hash_UpdateInfo;

/// from _cgo_export.h
typedef signed char GoInt8;
typedef unsigned char GoUint8;
typedef short GoInt16;
typedef unsigned short GoUint16;
typedef int GoInt32;
typedef unsigned int GoUint32;
typedef long long GoInt64;
typedef unsigned long long GoUint64;
#if __SIZEOF_SIZE_T__ == 4  //x86
typedef GoInt32 GoInt;
typedef GoUint32 GoUint;
#elif __SIZEOF_SIZE_T__ == 8  //x64
typedef GoInt64 GoInt;
typedef GoUint64 GoUint;
#else
#error Your architecture is not supported! (not 32/64 bits)
#endif //__SIZEOF_SIZE_T__
typedef __SIZE_TYPE__ GoUintptr;
typedef float GoFloat32;
typedef double GoFloat64;
typedef __complex float GoComplex64;
typedef __complex double GoComplex128;

typedef struct { char *p; GoInt n; } cgo_GoString;  //symbol conflict without "cgo_" prefix
typedef void *cgo_GoMap;
typedef void *cgo_GoChan;
typedef struct { void *t; void *v; } cgo_GoInterface;
typedef struct { void *data; GoInt len; GoInt cap; } cgo_GoSlice;
///

//For some reason we have to declare those manually if the file has a .cpp extension
//TODO: wrap all Go pointers in classes (RAII!)
char* go_authorNameById(void*, GoInt);
char* go_artistNameById(void*, GoInt);
char* go_genreNameById(void*, GoInt);
char* go_categoryNameById(void*, GoInt);
char* go_scanlatorNameById(void*, GoInt);
char* go_langNameById(void*, GoInt);
char* go_getThumbnailPath(char*);
void* go_ComicList_GetComic(void*, int);
int   go_ComicList_Len(void*);
void* go_ComicList_ComicUpdateInfo(void*, int);
int   go_Comic_ChaptersCount(void*);
int   go_Comic_ChaptersReadCount(void*);
void* go_Comic_Info(void*);
void* go_Comic_GetChapter(void*, int);
GoUint8 go_Chapter_AlreadyRead(void*);
int   go_Chapter_ScanlationsCount(void*);
void* go_Chapter_GetScanlation(void*, int);
void  go_collectGarbage(void*);
GoUintptr go_JointScanlators_ToSlice(GoUintptr);


#ifdef __cplusplus
} // extern "C"
#endif

#endif // QCAPI_H