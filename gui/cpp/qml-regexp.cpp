#include "qml-regexp.h"

RegExp::RegExp(const QString& pattern, QObject* parent) : QObject(parent), internal() {
	this->setPattern(pattern);
}

void RegExp::setPattern(const QString& pattern) {
	auto prevPattern = this->pattern();
	if (prevPattern == pattern) {
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