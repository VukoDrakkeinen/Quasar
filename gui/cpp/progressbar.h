#ifndef PROGRESSBAR_H
#define PROGRESSBAR_H

#include <QQuickPaintedItem>
class QPainter;
class QStyleOptionProgressBar;
class QTimer;

class ProgressBar : public QQuickPaintedItem
{
		Q_OBJECT
		Q_PROPERTY(bool indeterminate READ indeterminate WRITE setIndeterminate NOTIFY indeterminateChanged)
		Q_PROPERTY(double maximumValue READ maximumValue WRITE setMaximumValue NOTIFY maximumValueChanged)
		Q_PROPERTY(double minimumValue READ minimumValue WRITE setMinimumValue NOTIFY minimumValueChanged)
		Q_PROPERTY(Qt::Orientation orientation READ orientation WRITE setOrientation NOTIFY orientationChanged)
		Q_PROPERTY(double value READ value WRITE setValue NOTIFY valueChanged)
		//Q_PROPERTY(bool selected READ selected WRITE setSelected NOTIFY selectedChanged)

	public:
		ProgressBar(QQuickItem* parent = nullptr);
		virtual ~ProgressBar() {};
	public:
		virtual void paint(QPainter* painter);
	public:
		bool indeterminate();
		double maximumValue();
		double minimumValue();
		Qt::Orientation orientation();
		double value();
		//bool selected();
	public:
		void setIndeterminate(bool);
		void setMaximumValue(double);
		void setMinimumValue(double);
		void setOrientation(Qt::Orientation);
		void setValue(double);
		//void setSelected(bool);
	signals:
		void indeterminateChanged(bool);
		void maximumValueChanged(double);
		void minimumValueChanged(double);
		void orientationChanged(Qt::Orientation);
		void valueChanged(double);
		//void selectedChanged(bool);
	protected:
		QStyleOptionProgressBar* m_styleoption;
	private:
		bool m_indeterminate;
		double m_maximumValue;
		double m_minimumValue;
		Qt::Orientation m_orientation;
		double m_value;
		//bool m_selected;
	private:
		QTimer* updateTimer;

};

QML_DECLARE_TYPE(ProgressBar)

#endif // PROGRESSBAR_H
