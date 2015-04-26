#ifndef QCAPI_H
#define QCAPI_H

//#include "capi.h"
#include "/home/vuko/Projects/GoLang/GoPath/src/gopkg.in/qml.v1/cpp/capi.h" //TODO: relative path!

#ifdef __cplusplus
extern "C" {
#endif

typedef void QModel_;
typedef void QList_;

QList_* newList(void* data, int elemSize, int len, int titleOffset, int chapTotalOffset, int chapReadOffset, int dateTimeOffset, int progressOffset, int statusOffset, int ptrSize, int intSize);

QModel_* newModel(QList_* data);
void modelSetStore(QModel_* model, QList_* data);
//int modelAppendRow(QModel_* model, QVariant_* data);
int modelAppendRows(QModel_* model, QList_* data);
int modelRemoveRows(QModel_* model, int row, int count);


#ifdef __cplusplus
} // extern "C"
#endif

#endif // QCAPI_H