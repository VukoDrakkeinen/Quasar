#include "qml-regexp.h"

RegExp::RegExp(const QString& pattern, bool caseSensitive, QObject* parent) : QObject(parent), internal() {
	this->internal.setMinimal(true);
	this->setCaseSensitive(caseSensitive);
	this->setPattern(pattern);
}

void RegExp::setPattern(const QString& pattern) {
	if (this->pattern() == pattern) {
		return;
	}

	auto prevValidity = this->isValid();
	this->internal.setPattern(pattern);
	auto newValidity = this->isValid();

	if (newValidity != prevValidity) {
		emit validityChanged(newValidity);
	}
	if (!newValidity) {
		emit errorChanged(this->errorString());
	}
	emit patternChanged(pattern);
}

QString RegExp::pattern() const {
	return this->internal.pattern();
}

bool RegExp::isValid() const {
	return this->internal.isValid();
}

QString RegExp::errorString() const {
	return this->internal.errorString();
}

bool RegExp::caseSensitive() const {
	return this->internal.caseSensitivity() == Qt::CaseSensitive;
}

void RegExp::setCaseSensitive(bool cs) {
	if (this->caseSensitive() == cs) {
		return;
	}

	this->internal.setCaseSensitivity(cs ? Qt::CaseSensitive : Qt::CaseInsensitive);
	emit caseSensitivityChanged(cs);
}