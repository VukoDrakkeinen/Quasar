#include "chaptermodel.h"

#include <QBrush>
#include <QLocale>
#include <QDebug>

ChapterModel::ChapterModel(void* goComicList): QAbstractItemModel(), goComicList(goComicList), comicIdx(-1) {}
ChapterModel::~ChapterModel() {}

int ChapterModel::rowCount(const QModelIndex& parent) const
{
	if (parent.column() > 0) {
		return 0;
	}

	if (this->comicIdx == -1) {
		return 0;
	}
	
	//const int comicIdx = 0;	//TODO
	auto comic = go_ComicList_GetComic(this->goComicList, comicIdx);
	
	if (!parent.isValid()) {
		return go_Comic_ChaptersCount(comic);
	} else {
		auto chapter = go_Comic_GetChapter(comic, parent.row());
		int scanlationsCount = go_Chapter_ScanlationsCount(chapter);
		go_collectGarbage(chapter);
		return scanlationsCount - 1;	//start with 1
	}
	go_collectGarbage(comic);
}

int ChapterModel::columnCount(const QModelIndex& parent) const
{
	Q_UNUSED(parent)
	return 5;	//temporary?
}

QVariant ChapterModel::data(const QModelIndex& index, int role) const
{
	if (!index.isValid()) {
		return QVariant();
	}

	if (this->comicIdx == -1) {
    		return QVariant();
    }
	
	//const int comicIdx = 0;	//TODO
	
	ScanlationRow scanlation;
	int scanlationsCount;
	bool readStatus;
	
	auto goComic = go_ComicList_GetComic(this->goComicList, comicIdx);
	auto parent = index.parent();
	if (parent.isValid()) {
		auto goChapter = go_Comic_GetChapter(goComic, parent.row());
		readStatus = go_Chapter_AlreadyRead(goChapter);
		scanlationsCount = go_Chapter_ScanlationsCount(goChapter);
		auto goScanlation = go_Chapter_GetScanlation(goChapter, index.row()+1);
		//go_GC();
		scanlation = convertScanlation(goScanlation);
		go_collectGarbage(goChapter);
		go_collectGarbage(goScanlation);
	} else {
		auto goChapter = go_Comic_GetChapter(goComic, index.row());
		readStatus = go_Chapter_AlreadyRead(goChapter);
		scanlationsCount = go_Chapter_ScanlationsCount(goChapter);
		auto goScanlation = go_Chapter_GetScanlation(goChapter, 0);
        //go_GC();
        scanlation = convertScanlation(goScanlation);
        go_collectGarbage(goChapter);
        go_collectGarbage(goScanlation);
	}
	go_collectGarbage(goComic);
	Q_UNUSED(scanlationsCount) //TODO
	
	switch (role)
	{
		case Qt::ForegroundRole:
		{
			if (!readStatus) {
				return QBrush(QColor("green"));
			}
			return QBrush(QColor());
		}
		break;

		case Qt::DisplayRole:
		{
			switch (index.column())
			{
				case 0:	//TODO: show actual chapter identity
				{
					if (!parent.isValid()) {
						return QString("%1").arg(index.row()+1);
					}
					return QString("+%1").arg(index.row() + 1);
				}
				break;

				case 1:
					return scanlation.title;
				break;

				case 2:
				{
					QStringList scanlators;
					scanlators.reserve(scanlation.scanlatorIds.size());
					for (auto id : scanlation.scanlatorIds) {
						scanlators.append(go_scanlatorNameByIdQ(id));
					}
					return scanlators.join(" & ");
				}
				break;

				case 3:
					return go_langNameByIdQ(scanlation.languageId);
				break;
					
				case 4:
					return scanlation.pluginName;
				break;
			}
		}
		break;
			
		case ChapterModel::TitleRole:
		{
			return scanlation.title;
		}
		break;
		
		case ChapterModel::ScanlatorsRole:
		{
			QStringList scanlators;
			scanlators.reserve(scanlation.scanlatorIds.size());
			for (auto id : scanlation.scanlatorIds) {
				scanlators.append(go_scanlatorNameByIdQ(id));
			}
			return scanlators.join(" & ");
		}
		break;
		
		case ChapterModel::LangRole:
		{
			return go_langNameByIdQ(scanlation.languageId);
		}
		break;
		
		case ChapterModel::PluginRole:
		{
			return scanlation.pluginName;
		}
		break;
	}

	return QVariant();
}

QModelIndex ChapterModel::index(int row, int column, const QModelIndex& parent) const {
	if (!this->hasIndex(row, column, parent)) {
		return QModelIndex();
	}	
	
	if (parent.isValid() && parent.internalId() == 0) {
		return this->createIndex(row, column, static_cast<quintptr>(parent.row()+1));
	} else {
		return this->createIndex(row, column, static_cast<quintptr>(0));
	}
}

QModelIndex ChapterModel::parent(const QModelIndex& index) const {	
	if (!index.isValid() || index.internalId() == 0) {
		return QModelIndex();
	}
	
	return this->createIndex(index.internalId()-1, index.column(), static_cast<quintptr>(0));
}

bool ChapterModel::hasChildren(const QModelIndex& parent) const {
	if (parent.internalId() != 0) {
		return false;
	}
	
	return (this->rowCount(parent) != 0);
}


QVariant ChapterModel::headerData(int section, Qt::Orientation orientation, int role) const
{
	if (orientation != Qt::Horizontal)
	{
		return QVariant();
	}

	switch (role)
	{
		case Qt::DisplayRole:
			switch (section)
			{
				case 0:
					return QString("#");
					break;

				case 1:
					return QString("Title");
					break;

				case 2:
					return QString("Scanlator(s)");
					break;

				case 3:
					return QString("Language");
					break;

				case 4:
					return QString("Plugin");
					break;

				default:
					return QVariant();
			}

			break;
	}

	return QVariant();
}

void ChapterModel::setGoData(void* goComicList) {
	this->beginResetModel();
    this->goComicList = goComicList;
    this->endResetModel();
}

void ChapterModel::setComicIdx(int comicIdx) {
	this->beginResetModel();
    this->comicIdx = comicIdx;
    this->endResetModel();
}

QHash<int, QByteArray> ChapterModel::roleNames() const
{
	QHash<int, QByteArray> roles;
	roles[Qt::DisplayRole] = "display";
	roles[Qt::ForegroundRole] = "foreground";
	roles[ChapterModel::TitleRole] = "title";
	roles[ChapterModel::ScanlatorsRole] = "scanlators";
	roles[ChapterModel::LangRole] = "lang";
	roles[ChapterModel::PluginRole] = "plugin";
	return roles;
}

QVariant ChapterModel::qmlGet(int row, int column, const QString& roleName)
{
	auto role = this->roleNames().key(roleName.toLatin1(), -1);
	auto var = this->data(this->createIndex(row, column), role);
	return var;
}