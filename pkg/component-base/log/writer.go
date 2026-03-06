package log

type LogWriter struct {
	logger *zapLogger
}

func NewLogWriter(logger *zapLogger) *LogWriter {
	return &LogWriter{logger: logger}
}

func (w *LogWriter) Write(p []byte) (n int, err error) {
	w.logger.Error(string(p))
	return len(p), nil
}
