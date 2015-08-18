#ifndef ROWCACHE_H
#define ROWCACHE_H

#include <QDebug>
#include <QModelIndex>

template <typename T, int fieldN, int queriesPerFieldN>
class RowCache {
	public:
		RowCache() : max((fieldN-1)*queriesPerFieldN), validFor(max) {};
		~RowCache() {};
	public:
		bool valid(const QModelIndex& key) const {
			//qDebug() << "IsValid?" << key;
			if(this->validFor <= 0) {
				return false;
			}
			return this->key == key.sibling(key.row(), 0);  //compare, disregarding column
		}
		const T& get() {    //TODO?: merge valid() and get(); return some optional type
			this->validFor--;
			//qDebug() << "Cache hit!" << this->validFor << '\n';
			return this->item;
		}
		void hold(const QModelIndex& key, const T& item) {
			//qDebug() << "Cache miss!" << '\n';
			this->validFor = this->max;
			this->key = key.sibling(key.row(), 0);
			this->item = item;
		}
	private:
		T item;
		QModelIndex key;
		const int max;
		int validFor;
};

#endif //ROWCACHE_H
