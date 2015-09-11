#include "notifiablemodel.h"

NotifiableModel::NotifiableModel(void* goComicList) : goComicList(goComicList) {}

void NotifiableModel::emitBeginInsert(int row, int count) {
	this->beginInsertRows(QModelIndex(), row, row + count - 1);
}

void NotifiableModel::emitEndInsert() {
	this->endInsertRows();
}

void NotifiableModel::emitBeginRemove(int row, int count) {
	this->beginRemoveRows(QModelIndex(), row, row + count - 1);
}

void NotifiableModel::emitEndRemove() {
	this->endRemoveRows();
}

void NotifiableModel::setGoData(void* goComicList) {
	this->beginResetModel();
	this->goComicList = goComicList;
	this->endResetModel();
}

void NotifiableModel::emitBeginReset() {
	this->beginResetModel();
}

void NotifiableModel::emitEndReset() {
	this->endResetModel();
}

void NotifiableModel::emitDataChanged(int row, int count, int column) {
	if (count == 0) {
		return;
	}

	int colCount = this->columnCount();

	int firstColumn = column;
	int lastColumn = column;
	if (column == -1) {
		firstColumn = 0;
		lastColumn = colCount - 1;
	}

	for (int i = row; i < row + count; i++) {
		auto parent = this->index(i, 0);
		int childCount = this->rowCount(parent);
		if (childCount) {
			emit dataChanged(this->index(0, firstColumn, parent), this->index(childCount - 1, lastColumn, parent));
		}
	}

	emit dataChanged(this->index(row, firstColumn), this->index(row + count - 1, lastColumn));
}