package logger

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"math/rand"

	"go.uber.org/zap"
	"moul.io/zapfilter"
)

//const rule = "*:*,-udptracker* warn+:udptracker*"
const rule = "*"

var Log *zap.Logger

func Named(s string) *zap.Logger {
	return Log.Named(s)
}

type ctxKey string

var kCtxID = ctxKey("ctxID")

func Ctx(prev *zap.Logger, ctx context.Context) *zap.Logger {
	if v, ok := ctx.Value(kCtxID).(string); ok {
		return prev.With(zap.String("ctxID", v))
	} else {
		return prev
	}
}

func NewContextid(ctx context.Context) context.Context {
	w := make([]byte, 8)
	v := rand.Uint64()
	binary.BigEndian.PutUint64(w, v)
	f := hex.EncodeToString(w)
	return context.WithValue(ctx, kCtxID, f)

}

func init() {
	devLog, _ := zap.NewDevelopment()
	core := devLog.Core()
	// *=myns             => any level, myns namespace
	// info,warn:myns.*   => info or warn level, any namespace matching myns.*
	// error=*            => everything with error level
	//rule := "*:myns info,warn:myns.* error:*"
	Log = zap.New(zapfilter.NewFilteringCore(core, zapfilter.MustParseRules(rule)))
}
