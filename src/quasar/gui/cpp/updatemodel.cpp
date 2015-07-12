#include "updatemodel.h"

#include <QBrush>
#include <QLocale>
#include <QDebug>

UpdateInfoModel::UpdateInfoModel(void* goComicList): QAbstractTableModel(), goComicList(goComicList) {}
UpdateInfoModel::~UpdateInfoModel() {}

int UpdateInfoModel::rowCount(const QModelIndex& parent) const
{
	Q_UNUSED(parent)
	return go_ComicList_Len(this->goComicList);
}

int UpdateInfoModel::columnCount(const QModelIndex& parent) const
{
	Q_UNUSED(parent)
	return 5;	//temporary?
}

QVariant UpdateInfoModel::data(const QModelIndex& index, int role) const
{
	if (!index.isValid()) return QVariant();

	auto goComic = go_ComicList_ComicUpdateInfo(this->goComicList, index.row());
	//go_GC();
	auto updateInfo = convertUpdateInfo(goComic);

	switch (role)
	{
		case Qt::ForegroundRole:
		{
			QString color("red");

			switch (updateInfo.status)
			{
				case UpdateStatus::Error:
					color = "red";
					break;

				case UpdateStatus::NewChapters:
					color = "green";
					break;

				case UpdateStatus::NoUpdates:
					color = "gray";
					break;

				case UpdateStatus::Updating:
					color = "deepskyblue";
					break;
			}

			return QBrush(QColor(color));
		}
		break;

		case Qt::DisplayRole:
			switch (index.column())
			{
				case 0:
					return updateInfo.comicTitle;
					break;

				case 1:
					return updateInfo.chapTotal;
					break;

				case 2:
					return QString("%1 (%2%)").arg(updateInfo.chapRead).arg(updateInfo.chapRead ? (int)((float) updateInfo.chapRead*100/updateInfo.chapTotal) : 0);
					break;

				case 3:
				{
					auto dateTime = updateInfo.updated;
					auto date = dateTime.date();
					auto time = dateTime.time();
					auto currentDate = QDate::currentDate();

					if (date == currentDate)
					{
						return tr("Today") + " \t" + time.toString(Qt::SystemLocaleShortDate);
					}
					else if (date == currentDate.addDays(-1))
					{
						return tr("Yesterday") + " \t" + time.toString(Qt::SystemLocaleShortDate);
					} else if (date >= currentDate.addDays(-7)) {
						auto sysLocale = QLocale::system();	//weird day names handling workaround
						auto locale = QLocale(sysLocale.language(), sysLocale.country());
						return locale.toString(date, "dddd") + " \t" + time.toString(Qt::SystemLocaleShortDate);
					}
					return date.toString(Qt::SystemLocaleShortDate) + " \t" + time.toString(Qt::SystemLocaleShortDate);
				}
				break;
			}

		case UpdateInfoModel::CellTypeRole:
			if (index.column() == 4)
			{
				return QVariant::fromValue(CellType::ProgressBar);
			}
			else
			{
				return QVariant::fromValue(CellType::Normal);
			}

			break;

		case UpdateInfoModel::ProgressRole:
			return updateInfo.progress;
			break;

		case UpdateInfoModel::StatusRole:
			return QVariant::fromValue(updateInfo.status);
			break;
	}

	return QVariant();
}


QVariant UpdateInfoModel::headerData(int section, Qt::Orientation orientation, int role) const
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
					return QString("Chapters");
					break;

				case 2:
					return QString("Read");
					break;

				case 3:
					return QString("Updated");
					break;

				case 4:
					return QString("Status");
					break;

				default:
					return QVariant();
			}

			break;
	}

	return QVariant();
}

/*
bool UpdateInfoModel::appendRow(const UpdateInfoRow& row)
{
	this->beginInsertRows(QModelIndex(), this->rowCount(), this->rowCount());
	this->store.append(row);
	this->endInsertRows();
	return true;
}

void UpdateInfoModel::setStore(QList<UpdateInfoRow> store)
{
	this->beginResetModel();
	this->store = store;
	this->endResetModel();
}

bool UpdateInfoModel::appendRows(const QList<UpdateInfoRow> rows)
{
	this->beginInsertRows(QModelIndex(), this->rowCount(), this->rowCount() + rows.count());
	this->store.append(rows);
	this->endInsertRows();
	return true;
}


bool UpdateInfoModel::removeRows(int row, int count, const QModelIndex& parent)
{
	Q_UNUSED(parent)
	this->beginRemoveRows(QModelIndex(), row, count);
	this->store.erase(this->store.begin()+row, this->store.begin()+row+count);
	this->endRemoveRows();
	return true;
}//*/

void UpdateInfoModel::setGoData(void* goComicList) {
	this->beginResetModel();
    this->goComicList = goComicList;
    this->endResetModel();
}

QHash<int, QByteArray> UpdateInfoModel::roleNames() const
{
	QHash<int, QByteArray> roles;
	roles[Qt::DisplayRole] = "display";
	roles[Qt::ForegroundRole] = "foreground";
	roles[UpdateInfoModel::CellTypeRole] = "cellType";
	roles[UpdateInfoModel::ProgressRole] = "progress";
	roles[UpdateInfoModel::StatusRole] = "status";
	return roles;
}

QVariant UpdateInfoModel::qmlGet(int row, int column, const QString& roleName)
{
	auto role = this->roleNames().key(roleName.toLatin1(), -1);
	auto var = this->data(this->createIndex(row, column), role);
	if (QString(var.typeName()) == "UpdateStatus") {    //WORKAROUND: QML shitty enumerator handling
		return (uint) var.value<UpdateStatus>();
	}
	return var;
}