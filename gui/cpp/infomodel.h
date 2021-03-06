#ifndef ComicInfoModel_H
#define ComicInfoModel_H

#include <QDateTime>
#include <QList>
#include <QStringList>
#include <QtQml>
#include "rowcache.h"

class ComicType : public QObject {
		Q_OBJECT
	public:
		enum Enum {
			Invalid,
			Manga,
			Manhwa,
			Manhua,
			Western,
			Webcomic,
			Other
		};
		Q_ENUM(Enum)
};
class ComicStatus : public QObject {
        Q_OBJECT
    public:
	    enum Enum {
	        Invalid,
			Complete,
			Ongoing,
			OnHiatus,
			Discontinued
		};
		Q_ENUM(Enum)
};
class ScanlationStatus : public QObject {
        Q_OBJECT
    public:
        enum Enum {
	        Invalid,
			Complete,
			Ongoing,
			OnHiatus,
			Dropped,
			InDesperateNeedOfMoreStaff
		};
		Q_ENUM(Enum)
};

typedef struct {
	int mainTitleIdx;
	QStringList titles;
	QList<int> authorIds;
	QList<int> artistIds;
	QList<int> genreIds;
	QList<int> categoryIds;
	ComicType::Enum type;
	ComicStatus::Enum status;
	ScanlationStatus::Enum scanlationStatus;
	QString desc;
	quint16 rating;
	bool mature;
	int thumbnailIdx;
	QStringList thumbnails;
} ComicInfoRow;

class ComicInfoModel : public NotifiableModel
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
	public:
		enum DataRole {
            IdRole = Qt::UserRole,
        };
	private:
		mutable RowCache<ComicInfoRow, 14, 1> cache;
};

QML_DECLARE_TYPE(ComicInfoModel)

#endif // ComicInfoModel_H
