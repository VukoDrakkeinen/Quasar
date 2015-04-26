#include "model.h"

#include <QBrush>
#include <QLocale>
#include <QDebug>

ComicListModel::ComicListModel(QList<InfoRow> store): QAbstractTableModel(), store(store) {}
ComicListModel::~ComicListModel() {}

int ComicListModel::rowCount(const QModelIndex& parent) const
{
	Q_UNUSED(parent)
	return this->store.length();
}

int ComicListModel::columnCount(const QModelIndex& parent) const
{
	Q_UNUSED(parent)
	return 5;	//temporary?
}

QVariant ComicListModel::data(const QModelIndex& index, int role) const
{
	if (!index.isValid()) return QVariant();

	auto row = this->store.at(index.row());

	switch (role)
	{
		case Qt::ForegroundRole:
		{
			QString color("red");

			switch (row.status)
			{
				case ComicStatus::Error:
					color = "red";
					break;

				case ComicStatus::NewChapters:
					color = "green";
					break;

				case ComicStatus::NoUpdates:
					color = "gray";
					break;

				case ComicStatus::Updating:
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
					return row.title;
					break;

				case 1:
					return row.chapTotal;
					break;

				case 2:
					return QString("%1 (%2%)").arg(row.chapRead).arg(row.chapRead ? (int)((float) row.chapRead*100/row.chapTotal) : 0);
					break;

				case 3:
				{
					auto dateTime = row.updated;
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

		case ComicListModel::CellTypeRole:
			if (index.column() == 4)
			{
				return QVariant::fromValue(CellType::ProgressBar);
			}
			else
			{
				return QVariant::fromValue(CellType::Normal);
			}

			break;

		case ComicListModel::ProgressRole:
			return row.progress;
			break;

		case ComicListModel::StatusRole:
			return QVariant::fromValue(row.status);
			break;
	}

	return QVariant();
}


QVariant ComicListModel::headerData(int section, Qt::Orientation orientation, int role) const
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

bool ComicListModel::appendRow(const InfoRow& row)
{
	this->beginInsertRows(QModelIndex(), this->rowCount(), this->rowCount());
	this->store.append(row);
	this->endInsertRows();
	return true;
}

void ComicListModel::setStore(QList<InfoRow> store)
{
	this->beginResetModel();
	this->store = store;
	this->endResetModel();
}

bool ComicListModel::appendRows(const QList<InfoRow> rows)
{
	this->beginInsertRows(QModelIndex(), this->rowCount(), this->rowCount() + rows.count());
	this->store.append(rows);
	this->endInsertRows();
	return true;
}


bool ComicListModel::removeRows(int row, int count, const QModelIndex& parent)
{
	Q_UNUSED(parent)
	this->beginRemoveRows(QModelIndex(), row, count);
	this->store.erase(this->store.begin()+row, this->store.begin()+row+count);
	this->endRemoveRows();
	return true;
}

QHash<int, QByteArray> ComicListModel::roleNames() const
{
	QHash<int, QByteArray> roles;
	roles[Qt::DisplayRole] = "display";
	roles[Qt::ForegroundRole] = "foreground";
	roles[ComicListModel::CellTypeRole] = "cellType";
	roles[ComicListModel::ProgressRole] = "progress";
	roles[ComicListModel::StatusRole] = "status";
	return roles;
}

QVariant ComicListModel::qmlGet(int row, int column, const QString& roleName)
{
	auto role = this->roleNames().key(roleName.toLatin1(), -1);
	auto var = this->data(this->createIndex(row, column), role);
	if (QString(var.typeName()) == "ComicStatus") {    //WORKAROUND: QML shitty enumerator handling
		return (uint) var.value<ComicStatus>();
	}
	return var;
}