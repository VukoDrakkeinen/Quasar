#include "infomodel.h"

#include <QBrush>
#include <QLocale>
#include <QDebug>

ComicInfoModel::ComicInfoModel(void* goComicList): QAbstractTableModel(), goComicList(goComicList) {}
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
	if (!index.isValid()) return QVariant();

	auto goComic = go_ComicList_GetComic(this->goComicList, index.row());
	auto goInfo = go_Comic_Info(goComic);
	//go_GC();
	auto info = convertComicInfo(goInfo);
	go_collectGarbage(goComic); //TODO: wrapper

	switch (role)
	{
		case Qt::DecorationRole:
		{
			return go_getThumbnailPathQ(info.thumbnailFilename);
		}
		break;

		case Qt::DisplayRole:
			switch (index.column())	//TODO: roles, not column ids?
			{
				case 0:
					return info.mainTitle;
				break;

				case 1:
					return info.titles.join(", ");
				break;

				case 2:
				{
					QStringList authors;
					authors.reserve(info.authorIds.size());
					for (auto id : info.authorIds) {
						authors.append(go_authorNameByIdQ(id));
					}
					return authors.join(", ");
				}
				break;

				case 3:
				{
					QStringList artists;
					artists.reserve(info.artistIds.size());
					for (auto id : info.artistIds) {
						artists.append(go_artistNameByIdQ(id));
					}
					return artists.join(", ");	//TODO: nah, don't join. shit must be separate to be clickable
				}
				break;
				
				case 4:
				{
					QStringList genres;
					genres.reserve(info.genreIds.size());
					for (auto id : info.genreIds) {
						genres.append(go_genreNameByIdQ(id));
					}
					return genres.join(", ");
				}
				break;
					
				case 5:
				{
					QStringList tags;
					tags.reserve(info.categoryIds.size());
					for (auto id : info.categoryIds) {
						tags.append(go_categoryNameByIdQ(id));
					}
					return tags.join(", ");
				}
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
					return info.rating;
				break;
				
				case 11:
					return info.desc;
				break;
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

void ComicInfoModel::setGoData(void* goComicList) {
	this->beginResetModel();
	this->goComicList = goComicList;
	this->endResetModel();
}

QHash<int, QByteArray> ComicInfoModel::roleNames() const
{
	QHash<int, QByteArray> roles;
	roles[Qt::DecorationRole] = "decoration";
	roles[Qt::DisplayRole] = "display";
	roles[Qt::ForegroundRole] = "foreground";
	return roles;
}

QVariant ComicInfoModel::qmlGet(int row, int column, const QString& roleName)
{
	auto role = this->roleNames().key(roleName.toLatin1(), -1);
	auto var = this->data(this->createIndex(row, column), role);
	QString typeName(var.typeName());	//WORKAROUND: QML shitty enumerator handling
	if (typeName == "ComicType") {
		return (uint) var.value<ComicType>();
	}
	if (typeName == "ComicStatus") {
		return (uint) var.value<ComicStatus>();
	}
	if (typeName == "ScanlationStatus") {
		return (uint) var.value<ScanlationStatus>();
	}
	return var;
}