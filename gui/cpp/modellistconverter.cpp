#include "modellistconverter.h"
#include <QPoint>

QVariantList ModelListConverter::convertMany(QModelIndexList list) {
	auto ret = QVariantList();
	ret.reserve(list.size());
	for (const QModelIndex& m : list) {
		if (m.internalId() == 0) {
			ret.append(QPoint(m.row(), 0));  //main
		} else {
			ret.append(QPoint(static_cast<int>(m.internalId()-1), m.row()+1));    //child
		}
	}
	return ret;
}

QVariant ModelListConverter::convert(QModelIndex index) {
	if (index.internalId() == 0) {
		return QPoint(index.row(), 0);  //main
	} else {
		return QPoint(static_cast<int>(index.internalId()-1), index.row()+1);    //child
	}
}

static QObject* singleton_MLC_provider(QQmlEngine* engine, QJSEngine* scriptEngine) {
	Q_UNUSED(engine)
	Q_UNUSED(scriptEngine)

	return new ModelListConverter();
}