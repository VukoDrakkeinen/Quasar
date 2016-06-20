#include "progressbar.h"
#include <QApplication>
#include <QBrush>
#include <QStyleOptionProgressBar>
#include <QTimer>

ProgressBar::ProgressBar(QQuickItem* parent):
QQuickPaintedItem(parent),
m_styleoption(new QStyleOptionProgressBar()),
m_indeterminate(false),
m_maximumValue(100),
m_minimumValue(0),
m_value(0),
//m_selected(false),
updateTimer(new QTimer(this)) {
	updateTimer->setInterval(16);

	QObject::connect(this, SIGNAL(indeterminateChanged(bool)), this, SLOT(update()));
	QObject::connect(this, SIGNAL(maximumValueChanged(double)), this, SLOT(update()));
	QObject::connect(this, SIGNAL(minimumValueChanged(double)), this, SLOT(update()));
	QObject::connect(this, SIGNAL(orientationChanged(Qt::Orientation)), this, SLOT(update()));
	QObject::connect(this, SIGNAL(valueChanged(double)), this, SLOT(update()));
	//QObject::connect(this, SIGNAL(selectedChanged(bool)), this, SLOT(update()));
	QObject::connect(this->updateTimer, SIGNAL(timeout()), this, SLOT(update()));

	this->setImplicitWidth(100);
    this->setImplicitHeight(23);
    //qApp->style()->sizeFromContents(QStyle::CT_ProgressBar, this->m_styleoption, QSize(this->width(), this->height()));
}

void ProgressBar::paint(QPainter* painter) {
	auto stdPalette = QApplication::palette();
	
	auto m_paintMargins = 0;
	
	this->m_styleoption->state = 0;
	this->m_styleoption->direction = qApp->layoutDirection();
	this->m_styleoption->rect = QRect(m_paintMargins, 0, this->width() - 2 * m_paintMargins, this->height());
	this->m_styleoption->orientation = this->m_orientation;
	this->m_styleoption->minimum = this->m_indeterminate ? 0 : this->m_minimumValue;
	this->m_styleoption->maximum = this->m_indeterminate ? 0 : this->m_maximumValue;
	this->m_styleoption->progress = this->m_value;
	//this->m_styleoption->textVisible = true;
	//this->m_styleoption->text = "Updating...";
	if (this->isEnabled()) {
		this->m_styleoption->state |= QStyle::State_Enabled;
		this->m_styleoption->palette.setCurrentColorGroup(QPalette::Active);
	} else {
		this->m_styleoption->palette.setCurrentColorGroup(QPalette::Disabled);
	}
	this->m_styleoption->palette.setBrush(QPalette::Base, QBrush(Qt::NoBrush));
	//this->m_styleoption->palette.setBrush(QPalette::Base, this->m_selected ? stdPalette.highlight() : stdPalette.base());
	//this->m_styleoption->state |= QStyle::State_Horizontal & (this->m_orientation == Qt::Horizontal); //Hey, GCC, what is this
	this->m_styleoption->state |= QFlags<QStyle::StateFlag>(QStyle::State_Horizontal) & (this->m_orientation == Qt::Horizontal);    //Clang is being sane here
	qApp->style()->drawControl(QStyle::CE_ProgressBar, this->m_styleoption, painter);
}

bool ProgressBar::indeterminate() {
	return this->m_indeterminate;
}

double ProgressBar::maximumValue() {
	return this->m_maximumValue;
}

double ProgressBar::minimumValue() {
	return this->m_minimumValue;
}

Qt::Orientation ProgressBar::orientation() {
	return this->m_orientation;
}

double ProgressBar::value() {
	return this->m_value;
}

/*bool ProgressBar::selected() {
	return this->m_selected;
}//*/

void ProgressBar::setIndeterminate(bool indeterminate) {
	if (indeterminate == this->m_indeterminate) {
		return;
	}
	if (indeterminate) {
		this->updateTimer->start();
	} else {
		this->updateTimer->stop();
	}
	this->m_indeterminate = indeterminate;
	emit indeterminateChanged(indeterminate);
}

void ProgressBar::setMaximumValue(double maximumValue) {
	if (maximumValue != this->m_maximumValue) {
		this->m_maximumValue = maximumValue;
		emit maximumValueChanged(maximumValue);
	}
}

void ProgressBar::setMinimumValue(double minimumValue) {
	if (minimumValue != this->m_minimumValue) {
		this->m_minimumValue = minimumValue;
		emit minimumValueChanged(minimumValue);
	}
}

void ProgressBar::setOrientation(Qt::Orientation orientation) {
	if (orientation != this->m_orientation) {
		this->m_orientation = orientation;
		emit orientationChanged(orientation);
	}
}

void ProgressBar::setValue(double value) {
	if (value < this->m_minimumValue) {
		value = this->m_minimumValue;
	} else if (value > this->m_maximumValue) {
		value = this->m_maximumValue;
	} else if (value == this->m_value) {
		return;
	}
	this->m_value = value;
	emit valueChanged(value);
}

/*void ProgressBar::setSelected(bool selected) {
	if (selected != this->m_selected) {
		this->m_selected = selected;
		emit selectedChanged(selected);
	}
}//*/
