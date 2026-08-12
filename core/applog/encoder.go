package applog

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newJSONEncoder() zapcore.Encoder {
	encCfg := zap.NewProductionEncoderConfig()
	loc := effectiveLocation()
	encCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.In(loc).Format(time.RFC3339Nano))
	}
	return zapcore.NewJSONEncoder(encCfg)
}
