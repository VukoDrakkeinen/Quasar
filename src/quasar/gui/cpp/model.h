#ifndef ComicListModel_H
#define ComicListModel_H

#include <QAbstractTableModel>
#include <QDateTime>
#include <QtQml>

enum class UpdateStatus {
	NoUpdates,
	Updating,
	NewChapters,
	Error
};
Q_DECLARE_METATYPE(UpdateStatus)

struct UpdateInfoRow {
	QString comicTitle;
	int chapTotal, chapRead;
	QDateTime updated;
	int progress;
	UpdateStatus status;
};

class ComicListModel : public QAbstractTableModel
{
		Q_OBJECT

	public:
		ComicListModel() {};
		ComicListModel(QList<UpdateInfoRow> store);
		virtual ~ComicListModel();
	public:
		int rowCount(const QModelIndex& parent = QModelIndex()) const;
		int columnCount(const QModelIndex& parent = QModelIndex()) const;
		QVariant data(const QModelIndex& index, int role = Qt::DisplayRole) const;
		QVariant headerData(int section, Qt::Orientation orientation, int role = Qt::DisplayRole) const;
		QHash<int, QByteArray> roleNames() const;
		void setStore(QList<UpdateInfoRow> store);
		bool appendRow(const UpdateInfoRow& row);
		bool appendRows(const QList<UpdateInfoRow> rows);
		bool removeRows(int row, int count, const QModelIndex& parent = QModelIndex());
		Q_INVOKABLE QVariant qmlGet(int row, int column, const QString& roleName);
		//bool insertRows(int row, int count, const QModelIndex& parent = QModelIndex());
		//bool insertColumns(int column, int count, const QModelIndex & parent = QModelIndex());
		//bool removeColumns(int column, int count, const QModelIndex & parent = QModelIndex());
	private:
		QList<UpdateInfoRow> store;
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

Q_DECLARE_METATYPE(ComicListModel::CellType)
QML_DECLARE_TYPE(ComicListModel)

#endif // ComicListModel_H
