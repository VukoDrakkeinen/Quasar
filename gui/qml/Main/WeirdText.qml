import QtQuick 2.5

Canvas {
	id: root
	implicitWidth: textSize.width
	implicitHeight: textSize.height
	property string text;
	property font font;
	contextType: "2d"
	
	property string __cachedFontStr;
	
	onFontChanged: {
		__cachedFontStr = font.pointSize + "pt " + font.family;
	}

	onTextChanged: requestPaint()

	onPaint: {
		var ctxt = context ? context : getContext("2d");
		ctxt.reset();
		ctxt.font = __cachedFontStr
		ctxt.fillStyle = colorOf.text
		ctxt.fillText(text, 0, height)
	}
	
	SystemPalette { id: colorOf }
	TextMetrics { 
		id: textSize
		font: root.font
		text: root.text
	}
} 
