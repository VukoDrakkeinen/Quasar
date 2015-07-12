#ifndef ChapterModel_H
#define ChapterModel_H

#include <QAbstractTableModel>
#include <QDateTime>
#include <QtQml>

struct ScanlationRow {
	QString title;
	int languageId;
	QList<int> scanlatorIds;
	QString pluginName;
	QString url;
	QStringList pageLinks;
};

class ChapterModel : public QAbstractItemModel
{
		Q_OBJECT

	public:
		ChapterModel() {};
		ChapterModel(QList<UpdateInfoRow> store);
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
		void setGoData(void* goComicList);
		Q_INVOKABLE QVariant qmlGet(int row, int column, const QString& roleName);
	private:
		void* goComicList;
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
