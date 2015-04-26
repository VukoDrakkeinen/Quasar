#include "qcapi.h"
#include "model.h"
#include <QModelIndex>
#include <QList>
#include <cstdlib>
#include <cstring>
#include <QDebug>

using goint = quintptr;

QList_* newList(void* data, int elemSize, int len, int titleOffset, int chapTotalOffset, int chapReadOffset, int dateTimeOffset, int progressOffset, int statusOffset, int ptrSize, int intSize) {
	auto list = new QList<InfoRow>();
	for (int i = 0; i < len; i++) {
		quintptr elemPtr = (quintptr) data + (elemSize*i);
		char* goStr = (char*)*(quintptr*)(elemPtr+titleOffset);
		goint slen = *(goint*)(elemPtr+titleOffset+ptrSize);
		char* title = (char*) malloc(slen+1);
		memcpy(title, goStr, slen+1);
		title[slen] = '\0';
		goint chapTotal = *(goint*)(elemPtr+chapTotalOffset);
		goint chapRead = *(goint*)(elemPtr+chapReadOffset);
		qint64 dateTime = 1000 * *(qint64*)(elemPtr+dateTimeOffset);
		goint progress = *(goint*)(elemPtr+progressOffset);
		char status = *(char*)(elemPtr+statusOffset);
/*
		qDebug() << "Title:" << title;
		qDebug() << "ChapTotal:" << chapTotal;
		qDebug() << "ChapRead:" << chapRead;
		qDebug() << "DateTime:" << dateTime;
		qDebug() << "Progress:" << progress;
		qDebug() << "Status:" << (int) status;
//*/
		list->append(InfoRow{QString(title), (int)chapTotal, (int)chapRead, QDateTime::fromMSecsSinceEpoch(dateTime), (int)progress, (ComicStatus) status});
	}
	return list;
}

QModel_* newModel(QList_* data) {
	return reinterpret_cast<QModel_*>(new ComicListModel(*reinterpret_cast<QList<InfoRow>*>(data)));
}

void modelSetStore(QModel_* model, QList_* data) {
	reinterpret_cast<ComicListModel*>(model)->setStore(*reinterpret_cast<QList<InfoRow>*>(data));
}

/*
int modelAppendRow(QModel_* model, QVariantList_* data) {
	if (reinterpret_cast<ComicListModel*>(model)->appendRow()) {
		return 1;
	}
	return 0;
}
*/

int modelAppendRows(QModel_* model, QList_* data) {
	if (reinterpret_cast<ComicListModel*>(model)->appendRows(*reinterpret_cast<QList<InfoRow>*>(data))) {
		return 1;
    }
    return 0;
}

int modelRemoveRows(QModel_* model, int row, int count) {
	if (reinterpret_cast<ComicListModel*>(model)->removeRows(row, count, QModelIndex())) {
		return 1;
    }
    return 0;
}