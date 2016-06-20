#include "qcapi.h"
#include "notifiablemodel.h"
#include "updatemodel.h"
#include "infomodel.h"
#include "chaptermodel.h"
#include "progressbar.h"
#include "modellistconverter.h"
#include "qml-regexp.h"
#include "filtermodel.h"
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
QString go_ ## entity ## NameByIdQ(void* goComicList, int id) {   \
	auto cstr = go_ ## entity ## NameById(goComicList, id); \
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

WrappedModel_* wrapModel(NotifiableModel_* model) {
	auto wrapper = new SortFilterModel();
    wrapper->setSourceModel(reinterpret_cast<QAbstractItemModel*>(model));
    return reinterpret_cast<WrappedModel_*>(wrapper);
}

ComicInfoRow_* convertComicInfo(void* info) {
	typedef struct {
		GoUintptr mainTitleIdx;
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
		GoUintptr thumbnailIdx;
		GoUintptr thumbnails;
	} infoOffsets;
	auto offsets = (infoOffsets*) go_Offsets_ComicInfo;

	GoUintptr infoPtr = (GoUintptr) info;

	int mainTitleIdx = *(GoInt*)(infoPtr + offsets->mainTitleIdx);
    auto titles = SliceQ<GoString, QString>(GoSliceC(infoPtr + offsets->titles));
    auto authors = SliceQ<GoInt, int>(GoSliceC(infoPtr + offsets->authors));
    auto artists = SliceQ<GoInt, int>(GoSliceC(infoPtr + offsets->artists));
    auto genres = SliceQ<GoInt, int>(GoSliceC(infoPtr + offsets->genres));
    auto tags = SliceQ<GoInt, int>(GoSliceC(infoPtr + offsets->tags));
    auto type = (ComicType::Enum)*(GoInt*)(infoPtr + offsets->type);
    auto status = (ComicStatus::Enum)*(GoInt*)(infoPtr + offsets->status);
    auto scanStatus = (ScanlationStatus::Enum)*(GoInt*)(infoPtr + offsets->scanlationStatus);
    auto desc = GoStringQ(infoPtr + offsets->description);
    auto rating = *(GoUint16*)(infoPtr + offsets->rating);
    bool mature = *(GoInt*)(infoPtr + offsets->mature);
    int thumbnailIdx = *(GoInt*)(infoPtr + offsets->thumbnailIdx);
    auto thumbnails = SliceQ<GoString, QString>(GoSliceC(infoPtr + offsets->thumbnails));

	return new ComicInfoRow{
        mainTitleIdx, titles, authors, artists, genres, tags, type, status, scanStatus, desc, rating, mature, thumbnailIdx, thumbnails
    };
}

ScanlationRow_* convertScanlation(void* scanlation, void* scanlatorsPtr) {
	typedef struct {
		GoUintptr pluginName;
		GoUintptr scanlators;
		GoUintptr version;
		GoUintptr color;
		GoUintptr title;
		GoUintptr language;
		GoUintptr url;
		GoUintptr pageLinks;
	} scanlationOffsets;
	auto offsets = (scanlationOffsets*) go_Offsets_Scanlation;

	GoUintptr scanlationPtr = (GoUintptr) scanlation;

	auto pluginName = GoStringQ(scanlationPtr + offsets->pluginName);
//	auto scanlatorsPtr = go_JointScanlators_ToSlice(scanlationPtr + offsets->scanlators);
    auto scanlators = SliceQ<GoInt, int>(*(GoSlice*)scanlatorsPtr);
    auto version = (int)*(GoInt*)(scanlationPtr + offsets->version);
    auto color = (bool)*(GoUint8*)(scanlationPtr + offsets->color);
	auto title = GoStringQ(scanlationPtr + offsets->title);
	auto language = (int)*(GoInt*)(scanlationPtr + offsets->language);
	//auto scanlators = SliceQ<GoInt, int>(GoSliceC(scanlationPtr + offsets->scanlators));
	auto url = GoStringQ(scanlationPtr + offsets->url);
	auto pageLinks = SliceQ<GoString, QString>(GoSliceC(scanlationPtr + offsets->pageLinks));

	return new ScanlationRow{pluginName, scanlators, version, color, title, language, url, pageLinks};
}

UpdateInfoRow_* convertUpdateInfo(void* updateInfo) {
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

	return new UpdateInfoRow{title, chaptersCount, chaptersRead, updated, progress, status};
}

void* copyRawGoData(void* data, int size) {
	auto copy = malloc(size);
	memcpy(copy, data, size);
	return copy;
}

void registerQMLTypes() {
	auto enumText = QString("Uncreatable enumeration provider");
	qmlRegisterType<ProgressBar>("QuasarGUI", 1, 0, "SaneProgressBar");
	qmlRegisterSingletonType<ModelListConverter>("QuasarGUI", 1, 0, "ModelListConverter", singleton_MLC_provider);
	qmlRegisterUncreatableType<UpdateStatus>("QuasarGUI", 1, 0, "UpdateStatus", enumText);
	qmlRegisterUncreatableType<ComicType>("QuasarGUI", 1, 0, "ComicType", enumText);
	qmlRegisterUncreatableType<ComicStatus>("QuasarGUI", 1, 0, "ComicStatus", enumText);
	qmlRegisterUncreatableType<ScanlationStatus>("QuasarGUI", 1, 0, "ScanlationStatus", enumText);
	qmlRegisterUncreatableType<UpdateInfoModel>("QuasarGUI", 1, 0, "CellType", enumText);
	qmlRegisterType<RegExp>("QuasarGUI", 1, 0, "RegExp");
}

int assertSyncedHashes() {
	if (go_Hash_ComicInfo != 104341795181230237ull) {
		return 1;
	}
	if (go_Hash_Scanlation != 11865116033827439021ull) {
		return 2;
	}
	if (go_Hash_UpdateInfo != 10184097189468485639ull) {
		return 3;
    }
    return 0;
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

void notifyModelDataChanged(NotifiableModel_* model, int row, int count, int column) {
	reinterpret_cast<NotifiableModel*>(model)->emitDataChanged(row, count, column);
}