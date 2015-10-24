#include "filtermodel.h"

int SortFilterModel::comicId() const {
	return this->mapToSource(this->index(this->m_currentRow, 0)).row();
}

void SortFilterModel::setCurrentRow(int currentRow) {
	if (this->m_currentRow != currentRow) {
		this->m_currentRow = currentRow;
		emit comicIdChanged(this->comicId());
	}
}