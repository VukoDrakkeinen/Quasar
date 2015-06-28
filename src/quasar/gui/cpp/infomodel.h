#ifndef ComicInfoModel_H
#define ComicInfoModel_H

#include <QAbstractTableModel>
#include <QDateTime>
#include <QList>
#include <QStringList>
#include <QtQml>

enum class ComicType {
	Invalid,
	Manga,
	Manhwa,
	Manhua,
	Western,
	Webcomic,
	Other
};
enum class ComicStatus {
	Invalid,
	Complete,
	Ongoing,
	OnHiatus,
	Discontinued
};
enum class ScanlationStatus {
	Invalid,
	Complete,
	Ongoing,
	OnHiatus,
	Dropped,
	InDesperateNeedOfMoreStaff
};
Q_DECLARE_METATYPE(ComicType)
Q_DECLARE_METATYPE(ComicStatus)
Q_DECLARE_METATYPE(ScanlationStatus)

struct ComicInfoRow {
	QString mainTitle;
	QStringList titles; 
	QList<int> authorIds;
	QList<int> artistIds;
	QList<int> genreIds;
	QList<int> categoryIds;
	ComicType type;
	ComicStatus status;
	ScanlationStatus scanlationStatus;
	QString desc;
	float rating;
	bool mature;
	QString thumbnailFilename;
};

class ComicInfoModel : public QAbstractTableModel
{
		Q_OBJECT

	public:
		ComicInfoModel() {};
		ComicInfoModel(QList<ComicInfoRow> store);
		virtual ~ComicInfoModel();
	public:
		int rowCount(const QModelIndex& parent = QModelIndex()) const;
		int columnCount(const QModelIndex& parent = QModelIndex()) const;
		QVariant data(const QModelIndex& index, int role = Qt::DisplayRole) const;
		QVariant headerData(int section, Qt::Orientation orientation, int role = Qt::DisplayRole) const;
		QHash<int, QByteArray> roleNames() const;
		void setStore(QList<ComicInfoRow> store);
		bool appendRow(const ComicInfoRow& row);
		bool appendRows(const QList<ComicInfoRow> rows);
		bool removeRows(int row, int count, const QModelIndex& parent = QModelIndex());
		Q_INVOKABLE QVariant qmlGet(int row, int column, const QString& roleName);
		//bool insertRows(int row, int count, const QModelIndex& parent = QModelIndex());
		//bool insertColumns(int column, int count, const QModelIndex & parent = QModelIndex());
		//bool removeColumns(int column, int count, const QModelIndex & parent = QModelIndex());
	private:
		QList<ComicInfoRow> store;
};

QML_DECLARE_TYPE(ComicInfoModel)

#endif // ComicInfoModel_H
