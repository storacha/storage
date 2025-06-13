package storage

import (
	"strings"

	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/fx/fxevent"
)

var fxlog = logging.Logger("storage/fx")

// fxLogger adapts ipfs/go-log to fx's logger interface
type fxLogger struct{}

// NewFxLogger creates a new fx logger that uses ipfs/go-log
func NewFxLogger() fxevent.Logger {
	return &fxLogger{}
}

func (l *fxLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		fxlog.Debugw("OnStart hook executing",
			"callerName", e.CallerName,
			"functionName", e.FunctionName,
		)
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			fxlog.Errorw("OnStart hook failed",
				"callerName", e.CallerName,
				"functionName", e.FunctionName,
				"error", e.Err,
			)
		} else {
			fxlog.Debugw("OnStart hook executed",
				"callerName", e.CallerName,
				"functionName", e.FunctionName,
				"runtime", e.Runtime.String(),
			)
		}
	case *fxevent.OnStopExecuting:
		fxlog.Debugw("OnStop hook executing",
			"callerName", e.CallerName,
			"functionName", e.FunctionName,
		)
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			fxlog.Errorw("OnStop hook failed",
				"callerName", e.CallerName,
				"functionName", e.FunctionName,
				"error", e.Err,
			)
		} else {
			fxlog.Debugw("OnStop hook executed",
				"callerName", e.CallerName,
				"functionName", e.FunctionName,
				"runtime", e.Runtime.String(),
			)
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			fxlog.Errorw("Failed to supply",
				"typeName", e.TypeName,
				"error", e.Err,
			)
		} else {
			fxlog.Debugw("Supplied",
				"typeName", e.TypeName,
			)
		}
	case *fxevent.Provided:
		if e.Err != nil {
			fxlog.Errorw("Failed to provide",
				"constructorName", e.ConstructorName,
				"error", e.Err,
			)
		} else {
			outputTypes := make([]string, len(e.OutputTypeNames))
			for i, t := range e.OutputTypeNames {
				outputTypes[i] = t
			}
			fxlog.Debugw("Provided",
				"constructorName", e.ConstructorName,
				"outputTypes", strings.Join(outputTypes, ", "),
			)
		}
	case *fxevent.Decorated:
		if e.Err != nil {
			fxlog.Errorw("Failed to decorate",
				"decoratorName", e.DecoratorName,
				"error", e.Err,
			)
		} else {
			outputTypes := make([]string, len(e.OutputTypeNames))
			for i, t := range e.OutputTypeNames {
				outputTypes[i] = t
			}
			fxlog.Debugw("Decorated",
				"decoratorName", e.DecoratorName,
				"outputTypes", strings.Join(outputTypes, ", "),
			)
		}
	case *fxevent.Invoking:
		fxlog.Debugw("Invoking",
			"functionName", e.FunctionName,
		)
	case *fxevent.Invoked:
		if e.Err != nil {
			fxlog.Errorw("Failed to invoke",
				"functionName", e.FunctionName,
				"error", e.Err,
			)
		} else {
			fxlog.Debugw("Invoked",
				"functionName", e.FunctionName,
				"trace", e.Trace,
			)
		}
	case *fxevent.Stopping:
		fxlog.Debugw("Stopping",
			"signal", e.Signal.String(),
		)
	case *fxevent.Stopped:
		if e.Err != nil {
			fxlog.Errorw("Failed to stop",
				"error", e.Err,
			)
		} else {
			fxlog.Debugw("Stopped")
		}
	case *fxevent.RollingBack:
		fxlog.Warnw("Rolling back",
			"error", e.StartErr,
		)
	case *fxevent.RolledBack:
		if e.Err != nil {
			fxlog.Errorw("Failed to rollback",
				"error", e.Err,
			)
		} else {
			fxlog.Debugw("Rolled back")
		}
	case *fxevent.Started:
		if e.Err != nil {
			fxlog.Errorw("Failed to start",
				"error", e.Err,
			)
		} else {
			fxlog.Infow("Started")
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			fxlog.Errorw("Failed to initialize logger",
				"error", e.Err,
			)
		} else {
			fxlog.Debugw("Logger initialized",
				"loggerName", e.ConstructorName,
			)
		}
	}
}
