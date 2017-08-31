package diagnostic

import (
	"fmt"
	"log"
	"time"

	"github.com/influxdata/kapacitor"
	"github.com/influxdata/kapacitor/alert"
	"github.com/influxdata/kapacitor/keyvalue"
	"github.com/influxdata/kapacitor/models"
	alertservice "github.com/influxdata/kapacitor/services/alert"
	"github.com/influxdata/kapacitor/services/alerta"
	"github.com/influxdata/kapacitor/services/hipchat"
	"github.com/influxdata/kapacitor/services/pagerduty"
	"github.com/influxdata/kapacitor/services/slack"
	"github.com/influxdata/kapacitor/services/victorops"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type AlertServiceHandler struct {
	l *zap.Logger
}

func (h *AlertServiceHandler) WithHandlerContext(ctx ...keyvalue.T) alertservice.HandlerDiagnostic {
	fields := []zapcore.Field{}
	for _, kv := range ctx {
		fields = append(fields, zap.String(kv.Key, kv.Value))
	}

	return &AlertServiceHandler{
		l: h.l.With(fields...),
	}
}

func (h *AlertServiceHandler) MigratingHandlerSpecs() {
	h.l.Debug("migrating old v1.2 handler specs")
}

func (h *AlertServiceHandler) MigratingOldHandlerSpec(spec string) {
	h.l.Debug("migrating old handler spec", zap.String("handler", spec))
}

func (h *AlertServiceHandler) FoundHandlerRows(length int) {
	h.l.Debug("found handler rows", zap.Int("handler_row_count", length))
}

func (h *AlertServiceHandler) CreatingNewHandlers(length int) {
	h.l.Debug("creating new handlers in place of old handlers", zap.Int("handler_row_count", length))
}

func (h *AlertServiceHandler) FoundNewHandler(key string) {
	h.l.Debug("found new handler skipping", zap.String("handler", key))
}

func (h *AlertServiceHandler) Error(msg string, err error, ctx ...keyvalue.T) {
	if len(ctx) == 0 {
		h.l.Error(msg, zap.Error(err))
		return
	}

	if len(ctx) == 1 {
		el := ctx[0]
		h.l.Error(msg, zap.Error(err), zap.String(el.Key, el.Value))
		return
	}

	if len(ctx) == 2 {
		x := ctx[0]
		y := ctx[1]
		h.l.Error(msg, zap.Error(err), zap.String(x.Key, x.Value), zap.String(y.Key, y.Value))
		return
	}

	// This isn't great wrt to allocation, but should not ever actually occur
	fields := make([]zapcore.Field, len(ctx)+1) // +1 for error
	fields[0] = zap.Error(err)
	for i := 1; i < len(fields); i++ {
		kv := ctx[i-1]
		fields[i] = zap.String(kv.Key, kv.Value)
	}

	h.l.Error(msg, fields...)
}

// Kapcitor Handler

type KapacitorHandler struct {
	l *zap.Logger
}

// TODO: create TaskMasterHandler
func (h *KapacitorHandler) WithTaskContext(task string) kapacitor.TaskDiagnostic {
	return &KapacitorHandler{
		l: h.l.With(zap.String("task", task)),
	}
}

func (h *KapacitorHandler) WithTaskMasterContext(tm string) kapacitor.Diagnostic {
	return &KapacitorHandler{
		l: h.l.With(zap.String("task_master", tm)),
	}
}

func (h *KapacitorHandler) WithNodeContext(node string) kapacitor.NodeDiagnostic {
	return &KapacitorHandler{
		l: h.l.With(zap.String("node", node)),
	}
}

func (h *KapacitorHandler) WithEdgeContext(task, parent, child string) kapacitor.EdgeDiagnostic {
	return &KapacitorHandler{
		l: h.l.With(zap.String("task", task), zap.String("parent", parent), zap.String("child", child)),
	}
}

func (h *KapacitorHandler) TaskMasterOpened() {
	h.l.Info("opened task master")
}

func (h *KapacitorHandler) TaskMasterClosed() {
	h.l.Info("closed task master")
}

func (h *KapacitorHandler) StartingTask(task string) {
	h.l.Debug("starting task", zap.String("task", task))
}

func (h *KapacitorHandler) StartedTask(task string) {
	h.l.Info("started task", zap.String("task", task))
}

func (h *KapacitorHandler) StoppedTask(task string) {
	h.l.Info("stopped task", zap.String("task", task))
}

func (h *KapacitorHandler) StoppedTaskWithError(task string, err error) {
	h.l.Error("failed to stop task with out error", zap.String("task", task), zap.Error(err))
}

func (h *KapacitorHandler) TaskMasterDot(d string) {
	h.l.Debug("listing dot", zap.String("dot", d))
}

func (h *KapacitorHandler) ClosingEdge(collected int64, emitted int64) {
	h.l.Debug("closing edge", zap.Int64("collected", collected), zap.Int64("emitted", emitted))
}

//func (h *KapacitorHandler) WithContext(ctx ...keyvalue.T) kapacitor.Diagnostic {
//	fields := []zapcore.Field{}
//	for _, kv := range ctx {
//		fields = append(fields, zap.String(kv.Key, kv.Value))
//	}
//
//	return &KapacitorHandler{
//		l: h.l.With(fields...),
//	}
//}

func (h *KapacitorHandler) Error(msg string, err error, ctx ...keyvalue.T) {
	// Special case the three ways that the function is actually used
	// to avoid allocations
	if len(ctx) == 0 {
		h.l.Error(msg, zap.Error(err))
		return
	}

	if len(ctx) == 1 {
		el := ctx[0]
		h.l.Error(msg, zap.Error(err), zap.String(el.Key, el.Value))
		return
	}

	if len(ctx) == 2 {
		x := ctx[0]
		y := ctx[1]
		h.l.Error(msg, zap.Error(err), zap.String(x.Key, x.Value), zap.String(y.Key, y.Value))
		return
	}

	// This isn't great wrt to allocation, but should not ever actually occur
	fields := make([]zapcore.Field, len(ctx)+1) // +1 for error
	fields[0] = zap.Error(err)
	for i := 1; i < len(fields); i++ {
		kv := ctx[i-1]
		fields[i] = zap.String(kv.Key, kv.Value)
	}

	h.l.Error(msg, fields...)
}

func (h *KapacitorHandler) AlertTriggered(level alert.Level, id string, message string, rows *models.Row) {
	h.l.Debug("alert triggered",
		zap.Stringer("level", level),
		zap.String("id", id),
		zap.String("event_message", message),
		zap.String("data", fmt.Sprintf("%v", rows)),
	)
}

func (h *KapacitorHandler) SettingReplicas(new int, old int, id string) {
	h.l.Debug("setting replicas",
		zap.Int("new", new),
		zap.Int("old", old),
		// TODO: what is this ID?
		zap.String("id", id),
	)
}

func (h *KapacitorHandler) StartingBatchQuery(q string) {
	h.l.Debug("starting next batch query", zap.String("query", q))
}

func (h *KapacitorHandler) CannotPerformDerivative(reason string) {
	h.l.Error("cannot perform derivative", zap.String("reason", reason))
}

func (h *KapacitorHandler) MissingTagForFlattenOp(tag string) {
	h.l.Error("point missing tag for flatten operation", zap.String("tag", tag))
}

func (h *KapacitorHandler) IndexOutOfRangeForRow(idx int) {
	h.l.Error("index out of range for row update", zap.Int("index", idx))
}

func (h *KapacitorHandler) LoopbackWriteFailed() {
	h.l.Error("failed to write point over loopback")
}

func (h *KapacitorHandler) LogData(level string, prefix, data string) {
	switch level {
	case "info":
		h.l.Info("listing data", zap.String("prefix", prefix), zap.String("data", data))
	default:
	}
	h.l.Info("listing data", zap.String("prefix", prefix), zap.String("data", data))
}

func (h *KapacitorHandler) UDFLog(s string) {
	h.l.Info("UDF log", zap.String("text", s))
}

// Alerta handler

type AlertaHandler struct {
	l *zap.Logger
}

func (h *AlertaHandler) WithContext(ctx ...keyvalue.T) alerta.Diagnostic {
	fields := []zapcore.Field{}
	for _, kv := range ctx {
		fields = append(fields, zap.String(kv.Key, kv.Value))
	}

	return &AlertaHandler{
		l: h.l.With(fields...),
	}
}

func (h *AlertaHandler) TemplateError(err error, kv keyvalue.T) {
	h.l.Error("failed to evaluate Alerta template", zap.Error(err), zap.String(kv.Key, kv.Value))
}

func (h *AlertaHandler) Error(msg string, err error) {
	h.l.Error(msg, zap.Error(err))
}

// HipChat handler
type HipChatHandler struct {
	l *zap.Logger
}

func (h *HipChatHandler) WithContext(ctx ...keyvalue.T) hipchat.Diagnostic {
	fields := []zapcore.Field{}
	for _, kv := range ctx {
		fields = append(fields, zap.String(kv.Key, kv.Value))
	}

	return &HipChatHandler{
		l: h.l.With(fields...),
	}
}

func (h *HipChatHandler) Error(msg string, err error) {
	h.l.Error(msg, zap.Error(err))
}

// HTTPD handler

type HTTPDHandler struct {
	l *zap.Logger
}

func (h *HTTPDHandler) NewHTTPServerErrorLogger() *log.Logger {
	// TODO: implement
	//panic("not implemented")
	return nil
}

func (h *HTTPDHandler) StartingService() {
	h.l.Info("starting HTTP service")
}

func (h *HTTPDHandler) StoppedService() {
	h.l.Info("closed HTTP service")
}

func (h *HTTPDHandler) ShutdownTimeout() {
	h.l.Error("shutdown timedout, forcefully closing all remaining connections")
}

func (h *HTTPDHandler) AuthenticationEnabled(enabled bool) {
	h.l.Info("authentication", zap.Bool("enabled", enabled))
}

func (h *HTTPDHandler) ListeningOn(addr string, proto string) {
	h.l.Info("listening on", zap.String("addr", addr), zap.String("protocol", proto))
}

func (h *HTTPDHandler) WriteBodyReceived(body string) {
	h.l.Debug("write body received by handler: %s", zap.String("body", body))
}

func (h *HTTPDHandler) HTTP(
	host string,
	username string,
	start time.Time,
	method string,
	uri string,
	proto string,
	status int,
	referer string,
	userAgent string,
	reqID string,
	duration time.Duration,
) {
	// TODO: what is the message?
	h.l.Info("???",
		zap.String("host", host),
		zap.String("username", username),
		zap.Time("start", start),
		zap.String("method", method),
		zap.String("uri", uri),
		zap.String("protocol", proto),
		zap.Int("status", status),
		zap.String("referer", referer),
		zap.String("user-agent", userAgent),
		zap.String("request-id", reqID),
		zap.Duration("duration", duration),
	)
}

func (h *HTTPDHandler) RecoveryError(
	msg string,
	err string,
	host string,
	username string,
	start time.Time,
	method string,
	uri string,
	proto string,
	status int,
	referer string,
	userAgent string,
	reqID string,
	duration time.Duration,
) {
	h.l.Error(
		msg,
		zap.String("err", err),
		zap.String("host", host),
		zap.String("username", username),
		zap.Time("start", start),
		zap.String("method", method),
		zap.String("uri", uri),
		zap.String("protocol", proto),
		zap.Int("status", status),
		zap.String("referer", referer),
		zap.String("user-agent", userAgent),
		zap.String("request-id", reqID),
		zap.Duration("duration", duration),
	)
}

func (h *HTTPDHandler) Error(msg string, err error) {
	h.l.Error(msg, zap.Error(err))
}

// Reporting handler
type ReportingHandler struct {
	l *zap.Logger
}

func (h *ReportingHandler) Error(msg string, err error) {
	h.l.Error(msg, zap.Error(err))
}

// PagerDuty handler
type PagerDutyHandler struct {
	l *zap.Logger
}

func (h *PagerDutyHandler) WithContext(ctx ...keyvalue.T) pagerduty.Diagnostic {
	fields := []zapcore.Field{}
	for _, kv := range ctx {
		fields = append(fields, zap.String(kv.Key, kv.Value))
	}

	return &PagerDutyHandler{
		l: h.l.With(fields...),
	}
}

func (h *PagerDutyHandler) Error(msg string, err error) {
	h.l.Error(msg, zap.Error(err))
}

// Slack Handler

type SlackHandler struct {
	l *zap.Logger
}

func (h *SlackHandler) InsecureSkipVerify() {
	h.l.Warn("service is configured to skip ssl verification")
}

func (h *SlackHandler) Error(msg string, err error) {
	h.l.Error(msg, zap.Error(err))
}

func (h *SlackHandler) WithContext(ctx ...keyvalue.T) slack.Diagnostic {
	fields := []zapcore.Field{}
	for _, kv := range ctx {
		fields = append(fields, zap.String(kv.Key, kv.Value))
	}

	return &SlackHandler{
		l: h.l.With(fields...),
	}
}

// Storage Handler

type StorageHandler struct {
	l *zap.Logger
}

func (h *StorageHandler) Error(msg string, err error) {
	h.l.Error(msg, zap.Error(err))
}

// TaskStore Handler

type TaskStoreHandler struct {
	l *zap.Logger
}

func (h *TaskStoreHandler) StartingTask(taskID string) {
	h.l.Debug("starting enabled task on startup", zap.String("task", taskID))
}

func (h *TaskStoreHandler) StartedTask(taskID string) {
	h.l.Debug("started task during startup", zap.String("task", taskID))
}

func (h *TaskStoreHandler) FinishedTask(taskID string) {
	h.l.Debug("task finished", zap.String("task", taskID))
}

func (h *TaskStoreHandler) Debug(msg string) {
	h.l.Debug(msg)
}

func (h *TaskStoreHandler) Error(msg string, err error, ctx ...keyvalue.T) {
	// Special case the three ways that the function is actually used
	// to avoid allocations
	if len(ctx) == 0 {
		h.l.Error(msg, zap.Error(err))
		return
	}

	if len(ctx) == 1 {
		el := ctx[0]
		h.l.Error(msg, zap.Error(err), zap.String(el.Key, el.Value))
		return
	}

	if len(ctx) == 2 {
		x := ctx[0]
		y := ctx[1]
		h.l.Error(msg, zap.Error(err), zap.String(x.Key, x.Value), zap.String(y.Key, y.Value))
		return
	}

	// This isn't great wrt to allocation, but should not ever actually occur
	fields := make([]zapcore.Field, len(ctx)+1) // +1 for error
	fields[0] = zap.Error(err)
	for i := 1; i < len(fields); i++ {
		kv := ctx[i-1]
		fields[i] = zap.String(kv.Key, kv.Value)
	}

	h.l.Error(msg, fields...)
}

func (h *TaskStoreHandler) AlreadyMigrated(entity, id string) {
	h.l.Debug("entity has already been migrated skipping", zap.String(entity, id))
}

func (h *TaskStoreHandler) Migrated(entity, id string) {
	h.l.Debug("entity was migrated to new storage service", zap.String(entity, id))
}

// VictorOps Handler

type VictorOpsHandler struct {
	l *zap.Logger
}

func (h *VictorOpsHandler) Error(msg string, err error) {
	h.l.Error(msg, zap.Error(err))
}

func (h *VictorOpsHandler) WithContext(ctx ...keyvalue.T) victorops.Diagnostic {
	fields := []zapcore.Field{}
	for _, kv := range ctx {
		fields = append(fields, zap.String(kv.Key, kv.Value))
	}

	return &VictorOpsHandler{
		l: h.l.With(fields...),
	}
}

type UDFServiceHandler struct {
	l *zap.Logger
}

func (h *UDFServiceHandler) LoadedUDFInfo(udf string) {
	h.l.Debug("loaded UDF info", zap.String("udf", udf))
}
