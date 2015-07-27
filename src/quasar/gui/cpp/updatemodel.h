#ifndef UpdateInfoModel_H
#define UpdateInfoModel_H

#include <QDateTime>
#include <QtQml>
#include "notifiablemodel.h"
#include "rowcache.h"

class UpdateStatus : public QObject {
		Q_OBJECT
	public:
		enum Enum {
			NoUpdates,
			Updating,
			NewChapters,
			Error
		};
		Q_ENUM(Enum)
};

struct UpdateInfoRow {
	QString comicTitle;
	int chapTotal, chapRead;
	QDateTime updated;
	int progress;
	UpdateStatus::Enum status;
};

class UpdateInfoModel : public NotifiableModel
{
		Q_OBJECT

	public:
		UpdateInfoModel(void* goComicList);
		virtual ~UpdateInfoModel();
	public:
		int rowCount(const QModelIndex& parent = QModelIndex()) const;
		int columnCount(const QModelIndex& parent = QModelIndex()) const;
		QVariant data(const QModelIndex& index, int role = Qt::DisplayRole) const;
		QVariant headerData(int section, Qt::Orientation orientation, int role = Qt::DisplayRole) const;
		QHash<int, QByteArray> roleNames() const;
	public:
		Q_INVOKABLE QVariant qmlGet(int row, int column, const QString& roleName);
	private:
		mutable RowCache<UpdateInfoRow, 6> cache;
	public:
		enum DataRole {
			CellTypeRole = Qt::UserRole,
			ProgressRole,
			StatusRole
		};
		enum class CellType {
			Normal,
			ProgressBar
		};
};

Q_DECLARE_METATYPE(UpdateInfoModel::CellType)
QML_DECLARE_TYPE(UpdateInfoModel)

#endif // UpdateInfoModel_H
