import QtQuick 2.5
import QtQuick.Controls 1.4

RadioButton {
    id: disablingRadioButton
    property DisablingExclusiveGroup disabler: null
    property variant disablee: null

    onDisablerChanged: {
        if (disabler) {
            exclusiveGroup = disabler.data[0]	//HACK
            disabler.disableUnchecked()
        }
    }

    onDisableeChanged: {
        if (disabler) {
            if (disablee) {
                disabler.bindButton(disablingRadioButton)
            } else {
                disabler.unbindButton(disablingRadioButton)
            }
        }
    }

    //onClicked: disabler.disableUnchecked(disablee)
    onCheckedChanged: {
        if (disabler) {
            if (!checked) {
                disabler.disableUnchecked()
            }
        }
    }
}
