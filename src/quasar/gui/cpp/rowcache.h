#ifndef ROWCACHE_H
#define ROWCACHE_H

#include <QDebug>
#include <QModelIndex>

template <typename T, int N>
class RowCache {
	public:
		RowCache() : validFor(N) {};
		~RowCache() {};
	public:
		bool valid(const QModelIndex& key) const {
				if(this->validFor <= 0) {
					return false;
				}
				return this->key == key.sibling(key.row(), 0);  //compare, disregarding column
			}
		const T& get() {    //TODO?: merge valid() and get(); return some optional type
			//qDebug() << "Cache hit!";
			this->validFor--;
			return this->item;
		}
		void hold(const QModelIndex& key, const T& item) {
			//qDebug() << "Cache miss!";
			this->validFor = N;
			this->key = key.sibling(key.row(), 0);
			this->item = item;
		}
	private:
		T item;
		QModelIndex key;
		int validFor;
};

#endif //ROWCACHE_H
