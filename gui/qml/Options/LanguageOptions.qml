import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2
import "../Workarounds"
import "../utils.js" as U

RowLayout {
	id: root
	
	property alias modelOfDisabled: tableOfDisabled.model
	property alias modelOfEnabled: tableOfEnabled.model
	
	function getLanguages() {
		var langMap = {}
		modelOfEnabled.forEach(function(item) {
			return Object.defineProperty(langMap, item, {value: true, enumerable: true})
		})
		modelOfDisabled.forEach(function(item) {
			return Object.defineProperty(langMap, item, {value: false, enumerable: true})
		})
		return langMap
	}
	
	function setLanguages(languages) {
		var langsDisabled = []
		var langsEnabled = []
		for (var key in languages) {
			if (languages[key]) {
				langsEnabled.push(key)
			} else {
				langsDisabled.push(key)
			}
		}
		langsDisabled.sort()
		langsEnabled.sort()
		modelOfDisabled = langsDisabled
		modelOfEnabled = langsEnabled
	}
	
	function __byLangId(a, b) {
		return a.goid - b.goid
	}
	
	function __moveLangs(src, dst) {
		//var row = sources.currentRow	//currentRow may not be the first row in selection sometimes, so we can't use it :(
		var row = -1;
		src.selection.forEach(function (i) {if (row == -1) row = i;})	//so wasteful ;_;
		var count = src.selection.count
		
		if (count === 0) return;
		
		src.selection.clear()
		Array.prototype.push.apply(dst.model, src.model.splice(row, count))
		dst.model.sort()
		
		src.model = src.model	//trigger views update
		dst.model = dst.model
	}
	
	TableViewPatched {
		id: tableOfDisabled
		selectionMode: SelectionMode.ExtendedSelection
		TableViewColumn {
			title: "Disabled"
			role: "name"
		}
	}
	
	ColumnLayout {
		Layout.fillWidth: true
		Button {
			Layout.fillWidth: true
			id: btnMoveRight
			text: "->"
			action: Action {
				onTriggered: root.__moveLangs(tableOfDisabled, tableOfEnabled)
			}
			//TODO: how the fuck do I disable the buttons when there is no selection?! shit always reenables itself and removes my bindings ಠ_ಠ
		}
		Button {
			Layout.fillWidth: true
			id: btnMoveLeft
			text: "<-"
			action: Action {
				onTriggered: root.__moveLangs(tableOfEnabled, tableOfDisabled)
			}
		}
	}
	TableViewPatched {
		id: tableOfEnabled
		selectionMode: SelectionMode.ExtendedSelection
		TableViewColumn {
			title: "Enabled"
			role: "name"
		}
		
	}
} 
