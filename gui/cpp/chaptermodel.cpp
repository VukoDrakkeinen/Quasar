#include "chaptermodel.h"

#include <QBrush>
#include <QLocale>
#include <QApplication>
#include <QPalette>
#include <QDebug>

ChapterModel::ChapterModel(void* goComicList): NotifiableModel(goComicList), comicIdx(-1) {}
ChapterModel::~ChapterModel() {}

int ChapterModel::rowCount(const QModelIndex& parent) const
{
	if (parent.column() > 0) {
		return 0;
	}

	if (this->comicIdx == -1) {
		return 0;
	}

	auto comic = go_ComicList_GetComic(this->goComicList, this->comicIdx);    //TODO: cache for as long comicId doesn't change
	
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

    ScanlationRow scanlation;
    bool readStatus;
    int scanlationsCount;
	auto parent = index.parent();
    if (this->cache.valid(index)) {
        auto extScanlation = this->cache.get();
        scanlation = extScanlation.row;
        readStatus = extScanlation.readStatus;
        scanlationsCount = extScanlation.scanlationsCount;
    } else {
		auto goComic = go_ComicList_GetComic(this->goComicList, this->comicIdx);
		void* goChapter;
		if (parent.isValid()) {
			goChapter = go_Comic_GetChapter(goComic, parent.row());
			scanlation = *reinterpret_cast<ScanlationRow*>(go_Chapter_GetScanlation(goChapter, index.row()+1));
		} else {
			goChapter = go_Comic_GetChapter(goComic, index.row());
			scanlation = *reinterpret_cast<ScanlationRow*>(go_Chapter_GetScanlation(goChapter, 0));
		}
		readStatus = go_Chapter_AlreadyRead(goChapter);
		scanlationsCount = go_Chapter_ScanlationsCount(goChapter);
	    this->cache.hold(index, CachedScanlationRow{scanlation, readStatus, scanlationsCount});
	    go_collectGarbage(goChapter);
		go_collectGarbage(goComic);
		Q_UNUSED(scanlationsCount) //TODO
	}
	
	switch (role)
	{
		case Qt::ForegroundRole:
		{
			if (!readStatus) {
				return QBrush(QColor("green"));
			}
			return QApplication::palette().text();
			//return QBrush(QColor());
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
						scanlators.append(go_scanlatorNameByIdQ(this->goComicList, id));
					}
					return scanlators.join(" & ");
				}
				break;

				case 3:
					return go_langNameByIdQ(this->goComicList, scanlation.languageId);
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
				scanlators.append(go_scanlatorNameByIdQ(this->goComicList, id));
			}
			return scanlators.join(" & ");
		}
		break;
		
		case ChapterModel::LangRole:
		{
			return go_langNameByIdQ(this->goComicList, scanlation.languageId);
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

Qt::ItemFlags ChapterModel::flags(const QModelIndex& index) const {
	if (!index.isValid()) {
		return Qt::NoItemFlags;
	}

	auto flags = Qt::ItemIsSelectable | Qt::ItemIsEnabled;
	if (index.column() != 0) {
		flags |= Qt::ItemNeverHasChildren;
	}
	return flags;
}

int ChapterModel::comicId() const
{
	return this->comicIdx;
}

void ChapterModel::setComicId(int comicIdx) {
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

QVariant ChapterModel::qmlGet(int row, int column, const QString& roleName) const
{
	auto role = this->roleNames().key(roleName.toLatin1(), -1);
	return this->data(this->createIndex(row, column), role);
}