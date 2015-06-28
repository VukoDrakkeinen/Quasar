#ifndef QCAPI_H
#define QCAPI_H

#include "capi.h"

#ifdef __cplusplus
extern "C" {
#endif

typedef void QModel_;
typedef void QList_;
typedef void QInfoModel_;
typedef void QComicInfoList_;

QList_* newList(void* data, int elemSize, int len, int titleOffset, int chapTotalOffset, int chapReadOffset, int dateTimeOffset, int progressOffset, int statusOffset);
QComicInfoList_* newComicInfoList(void* infoSlice, void* coInfoSlice, int len, int infoSize, int sInfoSize, void* voffsets);

QModel_* newModel(QList_* data);
QInfoModel_* newInfoModel(QComicInfoList_* data);
void modelSetStore(QModel_* model, QList_* data);
//int modelAppendRow(QModel_* model, QVariant_* data);
int modelAppendRows(QModel_* model, QList_* data);
int modelRemoveRows(QModel_* model, int row, int count);

extern char* authorNameById(int);
extern char* artistNameById(int);
extern char* genreNameById(int);
extern char* categoryNameById(int);
extern char* getThumbnailPath(char*);

#ifdef __cplusplus
} // extern "C"
#endif

#endif // QCAPI_H