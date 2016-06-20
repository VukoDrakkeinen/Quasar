#ifndef ChapterModel_H
#define ChapterModel_H

#include <QDateTime>
#include <QtQml>
#include "notifiablemodel.h"
#include "rowcache.h"

struct ScanlationRow {  //TODO: make use of new fields
	QString pluginName;
	QList<int> scanlatorIds;
	int version;
	bool color;
	QString title;
	int languageId;
	QString url;
	QStringList pageLinks;
};

struct CachedScanlationRow {
	ScanlationRow row;
	bool readStatus;
	int scanlationsCount;
};

class ChapterModel : public NotifiableModel
{
		Q_OBJECT
		Q_PROPERTY(int comicId READ comicId WRITE setComicId NOTIFY comicIdChanged)
	public:
		ChapterModel(void* goComicList);
		virtual ~ChapterModel();
	public:
		int rowCount(const QModelIndex& parent = QModelIndex()) const;
		int columnCount(const QModelIndex& parent = QModelIndex()) const;
		QVariant data(const QModelIndex& index, int role = Qt::DisplayRole) const;
		QVariant headerData(int section, Qt::Orientation orientation, int role = Qt::DisplayRole) const;
		QModelIndex index(int row, int column, const QModelIndex & parent = QModelIndex()) const;
		QModelIndex parent(const QModelIndex& index) const;
		bool hasChildren(const QModelIndex& parent = QModelIndex()) const;
		QHash<int, QByteArray> roleNames() const;
		Qt::ItemFlags flags(const QModelIndex& index) const;
	public:
		Q_INVOKABLE int comicId() const;
		Q_INVOKABLE void setComicId(int comicIdx);
		Q_INVOKABLE QVariant qmlGet(int row, int column, const QString& roleName) const;
	signals:
		void comicIdChanged(int comicId);
	private:
		mutable RowCache<CachedScanlationRow, 8, 2> cache;
		int comicIdx;
	public:
		enum DataRole {
			TitleRole = Qt::UserRole,
			ScanlatorsRole,
			LangRole,
			PluginRole
		};
};

QML_DECLARE_TYPE(ChapterModel)

#endif // ChapterModel_H
