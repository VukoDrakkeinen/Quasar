#include "qcapi.h"
#include "_cgo_export.h"
#include "model.h"
#include "infomodel.h"
#include <QModelIndex>
#include <QList>
#include <cstdlib>
#include <cstring>
#include <QDebug>

typedef struct {
	quintptr ptr;
	GoInt size;
	GoInt cap;
} Go_Slice;

typedef struct {
	quintptr ptr;
	GoInt size;
} Go_String;

char* GoStringC(quintptr ptr) {
	char* goStr = (char*)*(quintptr*)(ptr);
	GoInt slen = *(GoInt*)(ptr+sizeof(GoInt));
	char* cstr = (char*) malloc(slen+1);
	memcpy(cstr, goStr, slen+1);
	cstr[slen] = '\0';
	return cstr;
}

QString GoStringQ(quintptr ptr) {
	char* cstr = GoStringC(ptr);
	QString qstr(cstr);
	free(cstr);
	return qstr;
}

Go_Slice GoSliceC(quintptr ptr) {
	return *(Go_Slice*)(ptr);
}

#define declareNameByIdQFuncFor(entity) \
QString entity ## NameByIdQ(int id) {   \
	auto cstr = entity ## NameById(id); \
	QString str(cstr);                  \
	free(cstr);                         \
	return str;                         \
}

declareNameByIdQFuncFor(author)
declareNameByIdQFuncFor(artist)
declareNameByIdQFuncFor(genre)
declareNameByIdQFuncFor(category)

QString getThumbnailPathQ(const QString& str) {
	char* cstr = getThumbnailPath(str.toLatin1().data());
	QString qstr(cstr);
	free(cstr);
	return qstr;
}

template <class GoType, class CType>
typename std::enable_if<!std::is_same<GoType, Go_String>::value, QList<CType>>::type
SliceQ(Go_Slice slice) {
	QList<CType> list;
	list.reserve(slice.size);
	for (int i = 0; i < slice.size; i++) {
		list.append((CType)*(GoType*)(slice.ptr + i * sizeof(GoType)));
	}
	return list;
}

template <class GoType, class CType>
typename std::enable_if<std::is_same<GoType, Go_String>::value, QList<CType>>::type
SliceQ(Go_Slice slice) {
	QList<CType> list;
	list.reserve(slice.size);
	for (int i = 0; i < slice.size; i++) {
		char* cstr = GoStringC(slice.ptr + i * sizeof(GoString));
		list.append(CType(cstr));
		free(cstr);
	}
	return list;
}

QList_* newList(void* data, int elemSize, int len, int titleOffset, int chapTotalOffset, int chapReadOffset, int dateTimeOffset, int progressOffset, int statusOffset) {
	auto list = new QList<UpdateInfoRow>();
	list->reserve(len);
	for (int i = 0; i < len; i++) {
		quintptr elemPtr = (quintptr) data + (elemSize*i);
		QString title = GoStringQ(elemPtr+titleOffset);
		GoInt chapTotal = *(GoInt*)(elemPtr+chapTotalOffset);
		GoInt chapRead = *(GoInt*)(elemPtr+chapReadOffset);
		qint64 dateTime = 1000 * *(qint64*)(elemPtr+dateTimeOffset);
		GoInt progress = *(GoInt*)(elemPtr+progressOffset);
		char status = *(char*)(elemPtr+statusOffset);
		list->append(UpdateInfoRow{title, (int)chapTotal, (int)chapRead, QDateTime::fromMSecsSinceEpoch(dateTime), (int)progress, (UpdateStatus) status});
	}
	return list;
}

QComicInfoList_* newComicInfoList(void* infoSlice, void* coInfoSlice, int len, int infoSize, int sInfoSize, void* voffsets) {
	typedef struct {
		quintptr mainTitle;
		quintptr authors;
		quintptr artists;
		quintptr type;
		quintptr status;
		quintptr scanlationStatus;
		quintptr description;
		quintptr rating;
		quintptr mature;
		quintptr thumbnailFilename;
		quintptr titles;
		quintptr genres;
		quintptr tags;
	} infoOffsets;
	auto offsets = (infoOffsets*) voffsets;

	auto list = new QList<ComicInfoRow>();
	list->reserve(len);
	for (int i = 0; i < len; i++) {
		quintptr infoPtr = (quintptr) infoSlice + (infoSize * i);
		quintptr coInfoPtr = (quintptr) coInfoSlice + (sInfoSize * i);

		QString title = GoStringQ(infoPtr + offsets->mainTitle);
        auto authors = SliceQ<GoInt, int>(GoSliceC(infoPtr + offsets->authors));
        auto artists = SliceQ<GoInt, int>(GoSliceC(infoPtr + offsets->artists));
        auto type = (ComicType)*(GoInt*)(infoPtr + offsets->type);
        auto status = (ComicStatus)*(GoInt*)(infoPtr + offsets->status);
        auto scanStatus = (ScanlationStatus)*(GoInt*)(infoPtr + offsets->scanlationStatus);
        QString desc = GoStringQ(infoPtr + offsets->description);
        float rating = *(float*)(infoPtr + offsets->rating);
        bool mature = *(GoInt*)(infoPtr + offsets->mature);
        QString thumbnail = GoStringQ(infoPtr + offsets->thumbnailFilename);

        auto altTitles = SliceQ<Go_String, QString>(GoSliceC(coInfoPtr + offsets->titles));
        auto genres = SliceQ<GoInt, int>(GoSliceC(coInfoPtr + offsets->genres));
        auto tags = SliceQ<GoInt, int>(GoSliceC(coInfoPtr + offsets->tags));

		auto row = ComicInfoRow{
			title, altTitles, authors, artists, genres, tags, type, status, scanStatus, desc, rating, mature,
            thumbnail
		};
		list->append(row);
	}
	return list;
}

QModel_* newModel(QList_* data) {
	return reinterpret_cast<QModel_*>(new ComicListModel(*reinterpret_cast<QList<UpdateInfoRow>*>(data)));
}

QInfoModel_* newInfoModel(QComicInfoList_* data) {
	return reinterpret_cast<QInfoModel_*>(new ComicInfoModel(*reinterpret_cast<QList<ComicInfoRow>*>(data)));
}

void modelSetStore(QModel_* model, QList_* data) {
	reinterpret_cast<ComicListModel*>(model)->setStore(*reinterpret_cast<QList<UpdateInfoRow>*>(data));
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
	if (reinterpret_cast<ComicListModel*>(model)->appendRows(*reinterpret_cast<QList<UpdateInfoRow>*>(data))) {
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