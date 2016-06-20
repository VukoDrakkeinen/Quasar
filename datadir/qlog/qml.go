package qlog

import "github.com/VukoDrakkeinen/qml"

func init() {
	qml.SetLogger(QMLRedirect{&defaultLogger})
}

type QMLRedirect struct {
	logger *QLogger
}

func (this QMLRedirect) QmlOutput(msg qml.LogMessage) error {
	var s msgSeverity
	switch msg.Severity() {
	case qml.LogDebug:
		s = Info
	case qml.LogWarning:
		s = Warning
	case qml.LogCritical:
		s = Error
	case qml.LogFatal:
		panic("Fatal QML error:" + msg.String())
	}
	this.logger.Log(s, msg)
	return nil
}
