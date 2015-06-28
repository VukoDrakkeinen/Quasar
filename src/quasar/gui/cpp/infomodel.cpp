#include "model.h"

#include <QBrush>
#include <QLocale>
#include <QDebug>

ComicInfoModel::ComicInfoModel(QList<ComicInfoRow> store): QAbstractTableModel(), store(store) {}
ComicInfoModel::~ComicInfoModel() {}

int ComicInfoModel::rowCount(const QModelIndex& parent) const
{
	Q_UNUSED(parent)
	return this->store.length();
}

int ComicInfoModel::columnCount(const QModelIndex& parent) const
{
	Q_UNUSED(parent)
	return 12;	//temporary?
}

QVariant ComicInfoModel::data(const QModelIndex& index, int role) const
{
	if (!index.isValid()) return QVariant();

	auto row = this->store.at(index.row());

	switch (role)
	{
		case Qt::DecorationRole:
		{
			return getThumbnailPathQ(row.thumbnailFilename);
		}
		break;

		case Qt::DisplayRole:
			switch (index.column())	//TODO: roles, not column ids?
			{
				case 0:
					return row.mainTitle;
				break;

				case 1:
					return row.titles.join(", ");
				break;

				case 2:
				{
					QStringList authors;
					authors.reserve(row.authorIds.size());
					for (auto id : row.authorIds) {
						authors.append(authorNameByIdQ(id));
					}
					return authors.join(", ");
				}
				break;

				case 3:
				{
					QStringList artists;
					artists.reserve(row.artistIds.size());
					for (auto id : row.artistIds) {
						artists.append(artistNameByIdQ(id));
					}
					return artists.join(", ");	//TODO: nah, don't join. shit must be separate to be clickable
				}
				break;
				
				case 4:
				{
					QStringList genres;
					genres.reserve(row.genreIds.size());
					for (auto id : row.genreIds) {
						genres.append(genreNameByIdQ(id));
					}
					return genres.join(", ");
				}
				break;
					
				case 5:
				{
					QStringList tags;
					tags.reserve(row.categoryIds.size());
					for (auto id : row.categoryIds) {
						tags.append(categoryNameByIdQ(id));
					}
					return tags.join(", ");
				}
				break;
					
				case 6:
					return QVariant::fromValue(row.type);
				break;
				
				case 7:
					return QVariant::fromValue(row.status);
				break;
				
				case 8:
					return QVariant::fromValue(row.scanlationStatus);
				break;
				
				case 9:
					return row.mature;
				break;
				
				case 10:
					return row.rating;
				break;
				
				case 11:
					return row.desc;
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

bool ComicInfoModel::appendRow(const ComicInfoRow& row)
{
	this->beginInsertRows(QModelIndex(), this->rowCount(), this->rowCount());
	this->store.append(row);
	this->endInsertRows();
	return true;
}

void ComicInfoModel::setStore(QList<ComicInfoRow> store)
{
	this->beginResetModel();
	this->store = store;
	this->endResetModel();
}

bool ComicInfoModel::appendRows(const QList<ComicInfoRow> rows)
{
	this->beginInsertRows(QModelIndex(), this->rowCount(), this->rowCount() + rows.count());
	this->store.append(rows);
	this->endInsertRows();
	return true;
}


bool ComicInfoModel::removeRows(int row, int count, const QModelIndex& parent)
{
	Q_UNUSED(parent)
	this->beginRemoveRows(QModelIndex(), row, count);
	this->store.erase(this->store.begin()+row, this->store.begin()+row+count);
	this->endRemoveRows();
	return true;
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