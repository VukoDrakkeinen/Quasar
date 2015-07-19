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

typedef struct {
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
} ComicInfoRow;

typedef struct {
	QString title;
	int languageId;
	QList<int> scanlatorIds;
	QString pluginName;
	QString url;
	QStringList pageLinks;
} QScanlation;

class ComicInfoModel : public QAbstractTableModel
{
		Q_OBJECT

	public:
		ComicInfoModel(void* goComicList);
		virtual ~ComicInfoModel();
	public:
		int rowCount(const QModelIndex& parent = QModelIndex()) const;
		int columnCount(const QModelIndex& parent = QModelIndex()) const;
		QVariant data(const QModelIndex& index, int role = Qt::DisplayRole) const;
		QVariant headerData(int section, Qt::Orientation orientation, int role = Qt::DisplayRole) const;
		QHash<int, QByteArray> roleNames() const;
		void setGoData(void* goComicList);
		Q_INVOKABLE QVariant qmlGet(int row, int column, const QString& roleName);
	private:
		void* goComicList;
};

QML_DECLARE_TYPE(ComicInfoModel)

#endif // ComicInfoModel_H
