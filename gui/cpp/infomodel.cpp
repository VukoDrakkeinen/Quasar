#include "infomodel.h"

#include <QBrush>
#include <QLocale>
#include <QDebug>
#include <QUrl>

ComicInfoModel::ComicInfoModel(void* goComicList): NotifiableModel(goComicList) {}
ComicInfoModel::~ComicInfoModel() {}

int ComicInfoModel::rowCount(const QModelIndex& parent) const
{
	Q_UNUSED(parent)
	return go_ComicList_Len(this->goComicList);
}

int ComicInfoModel::columnCount(const QModelIndex& parent) const
{
	Q_UNUSED(parent)
	return 12;	//temporary?
}

QVariant ComicInfoModel::data(const QModelIndex& index, int role) const
{
	static auto joinIds = [this](const QList<int>& ids, QString (*nameOf)(void* goComicList, int id), int avgLen) {
		QString joined;
        joined.reserve(ids.size() * avgLen);
        for (auto id : ids) {
            joined.append("<a href=\"");
            joined.append(QString::number(id));
            joined.append("\">");
            joined.append(nameOf(this->goComicList, id));
            joined.append("</a>, ");
        }
        joined.chop(2);    //remove the trailing ", "
        return joined;
	};
	if (!index.isValid()) return QVariant();

	ComicInfoRow info;
	if (this->cache.valid(index)) {
		info = this->cache.get();
	} else {
		auto goComic = go_ComicList_GetComic(this->goComicList, index.row());
		info = *reinterpret_cast<ComicInfoRow*>(go_Comic_Info(goComic));
		this->cache.hold(index, info);
		go_collectGarbage(goComic); //todo: wrapper
	}

	switch (role)
	{
		case Qt::DecorationRole:
		{
			return QUrl(go_getThumbnailPathQ(info.thumbnails[info.thumbnailIdx]));
		}
		break;

		case Qt::DisplayRole:
			switch (index.column())	//todo: roles, not column ids?
			{
				case 0:
					return info.titles[info.mainTitleIdx];
				break;

				case 1:
					info.titles.removeAt(info.mainTitleIdx);
					return info.titles.join(", ");
				break;

				case 2:
					return joinIds(info.authorIds, go_authorNameByIdQ, 33);
				break;

				case 3:
					return joinIds(info.artistIds, go_artistNameByIdQ, 33);
				break;
				
				case 4:
					return joinIds(info.genreIds, go_genreNameByIdQ, 28);
				break;
					
				case 5:
					return joinIds(info.categoryIds, go_categoryNameByIdQ, 34);
				break;
					
				case 6:
					return QVariant::fromValue(info.type);
				break;
				
				case 7:
					return QVariant::fromValue(info.status);
				break;
				
				case 8:
					return QVariant::fromValue(info.scanlationStatus);
				break;
				
				case 9:
					return info.mature;
				break;
				
				case 10:
					return QString("%1/10").arg(float(info.rating)/100);
				break;
				
				case 11:
					return info.desc;
				break;
			}

			case ComicInfoModel::IdRole:
				switch (index.column())
				{
					case 2:	return QVariant::fromValue(info.authorIds);
					case 3:	return QVariant::fromValue(info.artistIds);
					case 4:	return QVariant::fromValue(info.genreIds);
					case 5:	return QVariant::fromValue(info.categoryIds);
				}
	}

	return QVariant();
}


QVariant ComicInfoModel::headerData(int section, Qt::Orientation orientation, int role) const
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
					return QString("Title");
					break;

				case 1:
					return QString("AKA");
					break;

				case 2:
					return QString("Author(s)");
					break;

				case 3:
					return QString("Artist(s)");
					break;

				case 4:
					return QString("Genres");
					break;
					
				case 5:
					return QString("Categories");
					break;
					
				case 6:
					return QString("Type");
					break;
					
				case 7:
					return QString("Status");
					break;
					
				case 8:
					return QString("Scanlation");
					break;
					
				case 9:
					return QString("Mature");
					break;
					
				case 10:
					return QString("Rating");
					break;
					
				case 11:
					return QString("Description");
					break;

				default:
					return QVariant();
			}

			break;
	}

	return QVariant();
}

QHash<int, QByteArray> ComicInfoModel::roleNames() const
{
	QHash<int, QByteArray> roles;
	roles[Qt::DecorationRole] = "decoration";
	roles[Qt::DisplayRole] = "display";
	roles[Qt::ForegroundRole] = "foreground";
	roles[ComicInfoModel::IdRole] = "id";
	return roles;
}

QVariant ComicInfoModel::qmlGet(int row, int column, const QString& roleName)
{
	auto role = this->roleNames().key(roleName.toLatin1(), -1);
	return this->data(this->createIndex(row, column), role);
}