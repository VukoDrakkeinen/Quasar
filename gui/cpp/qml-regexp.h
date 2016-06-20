#ifndef QMLREGEXP_H
#define QMLREGEXP_H

#include <QRegExp>

class RegExp : public QObject {
		Q_OBJECT
		Q_PROPERTY(QString pattern READ pattern WRITE setPattern NOTIFY patternChanged)
		Q_PROPERTY(bool valid READ isValid NOTIFY validityChanged)
		Q_PROPERTY(QString error READ errorString NOTIFY errorChanged)
		Q_PROPERTY(QRegExp regexp READ Internal NOTIFY patternChanged)
		Q_PROPERTY(bool caseSensitive READ caseSensitive WRITE setCaseSensitive NOTIFY caseSensitivityChanged)
	public:
		RegExp(QObject* parent = nullptr) : QObject(parent), internal() {};
		RegExp(const QString& pattern, bool caseSensitive = false, QObject* parent = nullptr);
		virtual ~RegExp() {};
	public:
		operator QRegExp() const { return this->internal; }
		QRegExp Internal() const { return this->internal; }
	public:
		QString pattern() const;
		void setPattern(const QString& pattern);
		bool isValid() const;
		QString errorString() const;
		bool caseSensitive() const;
		void setCaseSensitive(bool);
	signals:
		void validityChanged(bool);
		void patternChanged(const QString&);
		void errorChanged(const QString&);
		void caseSensitivityChanged(bool);
		private:
			QRegExp internal;

};

#endif // QMLREGEXP_H