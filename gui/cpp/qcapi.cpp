#include "qcapi.h"
#include "notifiablemodel.h"
#include "updatemodel.h"
#include "infomodel.h"
#include "chaptermodel.h"
#include "progressbar.h"
#include <QModelIndex>
#include <QList>
#include <cstdlib>
#include <cstring>
#include <QDebug>

typedef struct {
	GoUintptr ptr;
	GoInt size;
	GoInt cap;
} GoSlice;

typedef struct {
	GoUintptr ptr;
	GoInt size;
} GoString;

void go_collectGarbage(GoUintptr ptr) {
	go_collectGarbage(reinterpret_cast<void*>(ptr));
}

char* GoStringC(GoUintptr ptr) {
	char* goStr = (char*)*(GoUintptr*)(ptr);
	GoInt slen = *(GoInt*)(ptr+sizeof(GoUintptr));
	char* cstr = (char*) malloc(slen+1);
	if (slen > 0) {
		memcpy(cstr, goStr, slen+1);
	}
	cstr[slen] = '\0';
	return cstr;
}

QString GoStringQ(GoUintptr ptr) {
	char* cstr = GoStringC(ptr);
	QString qstr(cstr);
	free(cstr);
	return qstr;
}

GoSlice GoSliceC(GoUintptr ptr) {
	return *(GoSlice*)(ptr);
}

#define declareNameByIdQFuncFor(entity) \
QString go_ ## entity ## NameByIdQ(int id) {   \
	auto cstr = go_ ## entity ## NameById(id); \
	QString str(cstr);                  \
	free(cstr);                         \
	return str;                         \
}

declareNameByIdQFuncFor(author)
declareNameByIdQFuncFor(artist)
declareNameByIdQFuncFor(genre)
declareNameByIdQFuncFor(category)
declareNameByIdQFuncFor(scanlator)
declareNameByIdQFuncFor(lang)

QString go_getThumbnailPathQ(const QString& str) {
	char* cstr = go_getThumbnailPath(str.toLatin1().data());
	QString qstr(cstr);
	free(cstr);
	return qstr;
}

template <class GoType, class CType>
typename std::enable_if<!std::is_same<GoType, GoString>::value, QList<CType>>::type
SliceQ(GoSlice slice) {
	QList<CType> list;
	list.reserve(slice.size);
	for (int i = 0; i < slice.size; i++) {
		list.append((CType)*(GoType*)(slice.ptr + i * sizeof(GoType)));
	}
	return list;
}

template <class GoType, class CType>
typename std::enable_if<std::is_same<GoType, GoString>::value, QList<CType>>::type
SliceQ(GoSlice slice) {
	QList<CType> list;
	list.reserve(slice.size);
	for (int i = 0; i < slice.size; i++) {
		char* cstr = GoStringC(slice.ptr + i * sizeof(GoString));
		list.append(CType(cstr));
		free(cstr);
	}
	return list;
}

InfoModel_* newInfoModel(GoComicList_* data) {
	return reinterpret_cast<InfoModel_*>(new ComicInfoModel(data));
}

UpdateModel_* newUpdateModel(GoComicList_* data) {
	return reinterpret_cast<UpdateModel_*>(new UpdateInfoModel(data));
}

ChapterModel_* newChapterModel(GoComicList_* data) {
	return reinterpret_cast<ChapterModel_*>(new ChapterModel(data));
}

ComicInfoRow convertComicInfo(void* info) {
	typedef struct {
		GoUintptr mainTitle;
		GoUintptr titles;
		GoUintptr authors;
		GoUintptr artists;
		GoUintptr genres;
		GoUintptr tags;
		GoUintptr type;
		GoUintptr status;
		GoUintptr scanlationStatus;
		GoUintptr description;
		GoUintptr rating;
		GoUintptr mature;
		GoUintptr thumbnailFilename;
	} infoOffsets;
	auto offsets = (infoOffsets*) go_Offsets_ComicInfo;

	GoUintptr infoPtr = (GoUintptr) info;

	QString mainTitle = GoStringQ(infoPtr + offsets->mainTitle);
    auto titles = SliceQ<GoString, QString>(GoSliceC(infoPtr + offsets->titles));
    auto authors = SliceQ<GoInt, int>(GoSliceC(infoPtr + offsets->authors));
    auto artists = SliceQ<GoInt, int>(GoSliceC(infoPtr + offsets->artists));
    auto genres = SliceQ<GoInt, int>(GoSliceC(infoPtr + offsets->genres));
    auto tags = SliceQ<GoInt, int>(GoSliceC(infoPtr + offsets->tags));
    auto type = (ComicType::Enum)*(GoInt*)(infoPtr + offsets->type);
    auto status = (ComicStatus::Enum)*(GoInt*)(infoPtr + offsets->status);
    auto scanStatus = (ScanlationStatus::Enum)*(GoInt*)(infoPtr + offsets->scanlationStatus);
    auto desc = GoStringQ(infoPtr + offsets->description);
    auto rating = *(float*)(infoPtr + offsets->rating);
    bool mature = *(GoInt*)(infoPtr + offsets->mature);
    QString thumbnail = GoStringQ(infoPtr + offsets->thumbnailFilename);

    go_collectGarbage(info);

	return ComicInfoRow{
        mainTitle, titles, authors, artists, genres, tags, type, status, scanStatus, desc, rating, mature, thumbnail
    };
}

ScanlationRow convertScanlation(void* scanlation) {
	typedef struct {
		GoUintptr title;
		GoUintptr language;
		GoUintptr scanlators;
		GoUintptr pluginName;
		GoUintptr url;
		GoUintptr pageLinks;
	} scanlationOffsets;
	auto offsets = (scanlationOffsets*) go_Offsets_Scanlation;

	GoUintptr scanlationPtr = (GoUintptr) scanlation;

	auto title = GoStringQ(scanlationPtr + offsets->title);
	auto language = (int)*(GoInt*)(scanlationPtr + offsets->language);
	auto scanlatorsPtr = go_JointScanlators_ToSlice(scanlationPtr + offsets->scanlators); //TODO: don't convert every time (see scanlators.go)
    auto scanlators = SliceQ<GoInt, int>(GoSliceC(scanlatorsPtr));
	//auto scanlators = SliceQ<GoInt, int>(GoSliceC(scanlationPtr + offsets->scanlators));
	auto pluginName = GoStringQ(scanlationPtr + offsets->pluginName);
	auto url = GoStringQ(scanlationPtr + offsets->url);
	auto pageLinks = SliceQ<GoString, QString>(GoSliceC(scanlationPtr + offsets->pageLinks));

	go_collectGarbage(scanlation);
	go_collectGarbage(scanlatorsPtr);

	return ScanlationRow{title, language, scanlators, pluginName, url, pageLinks};
}

UpdateInfoRow convertUpdateInfo(void* updateInfo) {
	typedef struct {
		GoUintptr title;
        GoUintptr chaptersCount;
        GoUintptr chaptersRead;
        GoUintptr updated;
        GoUintptr progress;
        GoUintptr status;
	} updateInfoOffsets;
	auto offsets = (updateInfoOffsets*) go_Offsets_UpdateInfo;

	GoUintptr updateInfoPtr = (GoUintptr) updateInfo;

	auto title = GoStringQ(updateInfoPtr + offsets->title);
	auto chaptersCount = (int)*(GoInt*)(updateInfoPtr + offsets->chaptersCount);
	auto chaptersRead = (int)*(GoInt*)(updateInfoPtr + offsets->chaptersRead);
	auto updated = QDateTime::fromMSecsSinceEpoch(1000 * *(GoInt64*)(updateInfoPtr + offsets->updated));
	auto progress = (int)*(GoInt8*)(updateInfoPtr + offsets->progress);
	auto status = (UpdateStatus::Enum)*(GoInt8*)(updateInfoPtr + offsets->status);

	go_collectGarbage(updateInfo);

	return UpdateInfoRow{title, chaptersCount, chaptersRead, updated, progress, status};
}

void* copyRawGoData(void* data, int size) {
	auto copy = malloc(size);
	memcpy(copy, data, size);
	return copy;
}

void registerQMLTypes() {
	auto enumText = QString("Uncreatable enumeration provider");
	qmlRegisterType<ProgressBar>("QuasarGUI", 1, 0, "SaneProgressBar");
	qmlRegisterUncreatableType<UpdateStatus>("QuasarGUI", 1, 0, "UpdateStatus", enumText);
	qmlRegisterUncreatableType<ComicType>("QuasarGUI", 1, 0, "ComicType", enumText);
	qmlRegisterUncreatableType<ComicStatus>("QuasarGUI", 1, 0, "ComicStatus", enumText);
	qmlRegisterUncreatableType<ScanlationStatus>("QuasarGUI", 1, 0, "ScanlationStatus", enumText);
	qmlRegisterUncreatableType<UpdateInfoModel>("QuasarGUI", 1, 0, "CellType", enumText);
}

void modelSetGoData(NotifiableModel_* model, void* goData) {
	reinterpret_cast<NotifiableModel*>(model)->setGoData(goData);
}

void notifyModelInsertStart(NotifiableModel_* model, int row, int count) {
	reinterpret_cast<NotifiableModel*>(model)->emitBeginInsert(row, count);
}

void notifyModelInsertEnd(NotifiableModel_* model) {
	reinterpret_cast<NotifiableModel*>(model)->emitEndInsert();
}

void notifyModelRemoveStart(NotifiableModel_* model, int row, int count) {
	reinterpret_cast<NotifiableModel*>(model)->emitBeginRemove(row, count);
}

void notifyModelRemoveEnd(NotifiableModel_* model) {
	reinterpret_cast<NotifiableModel*>(model)->emitEndRemove();
}

void notifyModelResetStart(NotifiableModel_* model) {
	reinterpret_cast<NotifiableModel*>(model)->emitBeginReset();
}

void notifyModelResetEnd(NotifiableModel_* model) {
	reinterpret_cast<NotifiableModel*>(model)->emitEndReset();
}