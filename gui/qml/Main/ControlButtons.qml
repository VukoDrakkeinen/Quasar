import QtQuick 2.5
import QtQuick.Layouts 1.2
import QtQuick.Controls 1.4

Item {
	Layout.minimumWidth: implicitWidth
	id: root
	implicitWidth: 0
	implicitHeight: buttons.childrenRect.height
    
    function calculateWidth() {
		return Array.prototype.slice.apply(buttons.children)
		.map(
			function(btn) {
				return btn.implicitWidth;
			}
		)
		.reduce(
			function(prev, curr, idx) {
				return prev > curr ? prev : curr;
			}
		)
    }

    ColumnLayout {
		id: buttons
		anchors.top: parent.top
		anchors.left: parent.left
		anchors.right: parent.right
        Component.onCompleted: root.implicitWidth = calculateWidth()
    }
    
    onChildrenChanged: {
		var reparented = false;
		Array.prototype.slice.apply(this.children)
			.forEach(function (child, i) {
				if (i >= 1) {	//WARNING(fixme): change the value for every internal Item added! (can we use JS hasOwnProperty and stuff to fix this?)
					child.Layout.fillWidth = true;
					child.parent = buttons;
					reparented = true;
				}
			}
		);
		if (reparented) {
			root.implicitWidth = calculateWidth()
		}
	}
}
