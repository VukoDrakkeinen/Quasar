import QtQuick 2.5
import QtQuick.Layouts 1.2
import QtGraphicalEffects 1.0

Item {
	implicitHeight: 8
	Layout.alignment: Qt.AlignHCenter | Qt.AlignTop
	Layout.fillWidth: true
	SystemPalette { id: colorOf }
	
	Column {
		x: 8
		y: 2
		width: parent.width - 16
		opacity: 0.3
		
		LinearGradient {
			width: parent.width
			height: 1
			start: Qt.point(0, 0)
			end: Qt.point(parent.width, 0)
			gradient: Gradient {
				GradientStop { color: colorOf.window; position: 0.0 }
				GradientStop { color: colorOf.base;   position: 0.1 }
				GradientStop { color: colorOf.base  ; position: 0.9 }
				GradientStop { color: colorOf.window; position: 1.0 }
			}
		}
		
		LinearGradient {
			width: parent.width
			height: 1
			start: Qt.point(0, 0)
			end: Qt.point(parent.width, 0)
			gradient: Gradient {
				GradientStop { color: colorOf.window; position: 0.0 }
				GradientStop { color: colorOf.shadow; position: 0.2 }
				GradientStop { color: colorOf.shadow; position: 0.8 }
				GradientStop { color: colorOf.window; position: 1.0 }
			}
		}
	}
} 
