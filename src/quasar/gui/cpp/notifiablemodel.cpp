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