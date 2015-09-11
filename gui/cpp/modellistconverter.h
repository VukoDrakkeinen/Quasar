#ifndef MODELLISTCONVERTER_H
#define MODELLISTCONVERTER_H

#include <QObject>
#include <QList>
#include <QModelIndex>
#include <QQmlEngine>
#include <QJSEngine>
#include <QVariant>

class ModelListConverter : public QObject {
		Q_OBJECT
	public:
		ModelListConverter() {};
		virtual ~ModelListConverter() {};
	public:
		Q_INVOKABLE QVariantList convertMany(QModelIndexList);
		Q_INVOKABLE QVariant convert(QModelIndex);

};

static QObject* singleton_MLC_provider(QQmlEngine* engine, QJSEngine* scriptEngine);

#endif // MODELLISTCONVERTER_H