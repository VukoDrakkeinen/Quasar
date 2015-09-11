#include "updatemodel.h"

#include <QBrush>
#include <QLocale>
#include <QDebug>

UpdateInfoModel::UpdateInfoModel(void* goComicList): NotifiableModel(goComicList) {}
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
	static auto formatTime = [](const QDateTime& updated) -> auto {
		auto date = updated.date();
        auto time = updated.time();
        auto currentDate = QDate::currentDate();

        if (date == currentDate)
        {
            return tr("Today") + " \t" + time.toString(Qt::SystemLocaleShortDate);
        }
        else if (date == currentDate.addDays(-1))
        {
            return tr("Yesterday") + " \t" + time.toString(Qt::SystemLocaleShortDate);
        } else if (date >= currentDate.addDays(-7)) {
            return date.toString("dddd") + " \t" + time.toString(Qt::SystemLocaleShortDate);
        }
        return date.toString(Qt::SystemLocaleShortDate) + " \t" + time.toString(Qt::SystemLocaleShortDate);
	};

	static auto formatRead = [](int chapTotal, int chapRead) -> auto {
		return QString("%1 (%2%)").arg(chapRead).arg(chapRead ? int(float(chapRead*100/chapTotal)) : 0);
	};

	if (!index.isValid()) return QVariant();

	UpdateInfoRow updateInfo;
	if (this->cache.valid(index)) {
		updateInfo = this->cache.get();
	} else {
		auto goComic = go_ComicList_ComicUpdateInfo(this->goComicList, index.row());
		updateInfo = convertUpdateInfo(goComic);
		this->cache.hold(index, updateInfo);
	}

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
					return formatRead(updateInfo.chapTotal, updateInfo.chapRead);
					break;

				case 3:
				{
					return formatTime(updateInfo.updated);
				}
				break;
			}
			break;

		case UpdateInfoModel::CellTypeRole:
			if (index.column() == 4) {
				return QVariant::fromValue(CellType::ProgressBar);
			} else {
				return QVariant::fromValue(CellType::Normal);
			}
			break;

		case UpdateInfoModel::ProgressRole:
			return updateInfo.progress;
			break;

		case UpdateInfoModel::StatusRole:
			return QVariant::fromValue(updateInfo.status);
			break;
			
		case UpdateInfoModel::TitleRole:
			return updateInfo.comicTitle;
			break;
			
		case UpdateInfoModel::ChaptersRole:
		  	return updateInfo.chapTotal;
			break;
			
		case UpdateInfoModel::ReadRole:
			return formatRead(updateInfo.chapTotal, updateInfo.chapRead);
			break;
						
		case UpdateInfoModel::TimeRole:
			return formatTime(updateInfo.updated);
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

QHash<int, QByteArray> UpdateInfoModel::roleNames() const
{
	QHash<int, QByteArray> roles;
	roles[Qt::DisplayRole] = "display";
	roles[Qt::ForegroundRole] = "foreground";
	roles[UpdateInfoModel::CellTypeRole] = "cellType";
	roles[UpdateInfoModel::ProgressRole] = "progress";
	roles[UpdateInfoModel::StatusRole] = "status";
	roles[UpdateInfoModel::TitleRole] = "title";
	roles[UpdateInfoModel::ChaptersRole] = "chapters";
	roles[UpdateInfoModel::ReadRole] = "read";
	roles[UpdateInfoModel::TimeRole] = "time";
	return roles;
}

QVariant UpdateInfoModel::qmlGet(int row, int column, const QString& roleName)
{
	auto role = this->roleNames().key(roleName.toLatin1(), -1);
	return this->data(this->createIndex(row, column), role);
}