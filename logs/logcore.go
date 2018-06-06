package logs

import "go.uber.org/zap/zapcore"

type loggerCore struct {
	zapcore.LevelEnabler
	enc zapcore.Encoder
	out ExtWriterSync
}

func (c *loggerCore) With(fields []zapcore.Field) zapcore.Core {
	clone := c.clone()
	addFields(clone.enc, fields)
	return clone
}

func (c *loggerCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *loggerCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	buf, err := c.enc.EncodeEntry(ent, fields)
	if err != nil {
		return err
	}
	_, err = c.out.ExtWrite(buf.Bytes(), ent)
	buf.Free()
	if err != nil {
		return err
	}
	if ent.Level > zapcore.ErrorLevel {
		// Since we may be crashing the program, sync the output. Ignore Sync
		// errors, pending a clean solution to issue #370.
		c.Sync()
	}
	return nil
}

func (c *loggerCore) Sync() error {
	return c.out.Sync()
}

func (c *loggerCore) clone() *loggerCore {
	return &loggerCore{
		LevelEnabler: c.LevelEnabler,
		enc:          c.enc.Clone(),
		out:          c.out,
	}
}

func addFields(enc zapcore.ObjectEncoder, fields []zapcore.Field) {
	for i := range fields {
		fields[i].AddTo(enc)
	}
}
