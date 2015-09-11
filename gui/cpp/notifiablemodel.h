#ifndef NOTIFIABLEMODEL_H
#define NOTIFIABLEMODEL_H

#include <QAbstractTableModel>

class NotifiableModel : public QAbstractTableModel
{
	public:
		NotifiableModel(void* goComicList);
	public:
		void setGoData(void* goComicList);
		void emitBeginInsert(int row, int count);
		void emitEndInsert();
		void emitBeginRemove(int row, int count);
		void emitEndRemove();
		void emitBeginReset();
        void emitEndReset();
        void emitDataChanged(int row, int count, int column = -1);
	protected:
    	void* goComicList;
};

#endif // NOTIFIABLEMODEL_H
