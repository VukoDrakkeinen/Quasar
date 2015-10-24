#ifndef FILTERMODEL_H
#define FILTERMODEL_H

#include <QSortFilterProxyModel>

class SortFilterModel : public QSortFilterProxyModel {
		Q_OBJECT
		Q_PROPERTY(int comicId READ comicId NOTIFY comicIdChanged)
		Q_PROPERTY(int currentRow MEMBER m_currentRow WRITE setCurrentRow)
	public:
		SortFilterModel(QObject* parent = nullptr) : QSortFilterProxyModel(parent), m_currentRow(-1) {};
		virtual ~SortFilterModel() {};
	public:
		void setCurrentRow(int currentRow);
		int comicId() const;
	signals:
		void comicIdChanged(int);
	private:
		int m_currentRow;
};

#endif // FILTERMODEL_H