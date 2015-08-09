function qmltypeof(obj, className) {
	var str = obj.toString();
	return str.indexOf(className + "(") == 0 || str.indexOf(className + "_QML") == 0;
}

function getProperties(obj) {
	var result = [];
	for (var id in obj) {
		try {
			//if (typeof(obj[id]) == "function") {
			result.push(id + ": " + obj[id].toString());
			//}
		} catch (err) {
			result.push(id + ": inaccessible (" + err.toString() + ")");
		}
	}
	return result;
} 