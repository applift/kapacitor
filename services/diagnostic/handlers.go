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
	"github.com/influxdata/kapacitor/services/httppost"
	"github.com/influxdata/kapacitor/services/k8s"
	"github.com/influxdata/kapacitor/services/mqtt"
	"github.com/influxdata/kapacitor/services/opsgenie"
	"github.com/influxdata/kapacitor/services/pagerduty"
	"github.com/influxdata/kapacitor/services/pushover"
	"github.com/influxdata/kapacitor/services/sensu"
	"github.com/influxdata/kapacitor/services/slack"
	"github.com/influxdata/kapacitor/services/smtp"
	"github.com/influxdata/kapacitor/services/snmptrap"
	"github.com/influxdata/kapacitor/services/swarm"
	"github.com/influxdata/kapacitor/services/talk"
	"github.com/influxdata/kapacitor/services/telegram"
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

type SMTPHandler struct {
	l *zap.Logger
}

func (h *SMTPHandler) Error(msg string, err error) {
	h.l.Error(msg, zap.Error(err))
}

func (h *SMTPHandler) WithContext(ctx ...keyvalue.T) smtp.Diagnostic {
	fields := []zapcore.Field{}
	for _, kv := range ctx {
		fields = append(fields, zap.String(kv.Key, kv.Value))
	}

	return &SMTPHandler{
		l: h.l.With(fields...),
	}
}

type OpsGenieHandler struct {
	l *zap.Logger
}

func (h *OpsGenieHandler) Error(msg string, err error) {
	h.l.Error(msg, zap.Error(err))
}

func (h *OpsGenieHandler) WithContext(ctx ...keyvalue.T) opsgenie.Diagnostic {
	fields := []zapcore.Field{}
	for _, kv := range ctx {
		fields = append(fields, zap.String(kv.Key, kv.Value))
	}

	return &OpsGenieHandler{
		l: h.l.With(fields...),
	}
}

// UDF service handler

type UDFServiceHandler struct {
	l *zap.Logger
}

func (h *UDFServiceHandler) LoadedUDFInfo(udf string) {
	h.l.Debug("loaded UDF info", zap.String("udf", udf))
}

// Pushover handler

type PushoverHandler struct {
	l *zap.Logger
}

func (h *PushoverHandler) Error(msg string, err error) {
	h.l.Error(msg, zap.Error(err))
}

func (h *PushoverHandler) WithContext(ctx ...keyvalue.T) pushover.Diagnostic {
	fields := []zapcore.Field{}
	for _, kv := range ctx {
		fields = append(fields, zap.String(kv.Key, kv.Value))
	}

	return &PushoverHandler{
		l: h.l.With(fields...),
	}
}

// Template handler

type HTTPPostHandler struct {
	l *zap.Logger
}

func (h *HTTPPostHandler) Error(msg string, err error) {
	h.l.Error(msg, zap.Error(err))
}

func (h *HTTPPostHandler) WithContext(ctx ...keyvalue.T) httppost.Diagnostic {
	fields := []zapcore.Field{}
	for _, kv := range ctx {
		fields = append(fields, zap.String(kv.Key, kv.Value))
	}

	return &HTTPPostHandler{
		l: h.l.With(fields...),
	}
}

// Sensu handler

type SensuHandler struct {
	l *zap.Logger
}

func (h *SensuHandler) Error(msg string, err error, ctx ...keyvalue.T) {
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

func (h *SensuHandler) WithContext(ctx ...keyvalue.T) sensu.Diagnostic {
	fields := []zapcore.Field{}
	for _, kv := range ctx {
		fields = append(fields, zap.String(kv.Key, kv.Value))
	}

	return &SensuHandler{
		l: h.l.With(fields...),
	}
}

// SNMPTrap handler

type SNMPTrapHandler struct {
	l *zap.Logger
}

func (h *SNMPTrapHandler) Error(msg string, err error) {
	h.l.Error(msg, zap.Error(err))
}

func (h *SNMPTrapHandler) WithContext(ctx ...keyvalue.T) snmptrap.Diagnostic {
	fields := []zapcore.Field{}
	for _, kv := range ctx {
		fields = append(fields, zap.String(kv.Key, kv.Value))
	}

	return &SNMPTrapHandler{
		l: h.l.With(fields...),
	}
}

// Telegram handler

type TelegramHandler struct {
	l *zap.Logger
}

func (h *TelegramHandler) Error(msg string, err error) {
	h.l.Error(msg, zap.Error(err))
}

func (h *TelegramHandler) WithContext(ctx ...keyvalue.T) telegram.Diagnostic {
	fields := []zapcore.Field{}
	for _, kv := range ctx {
		fields = append(fields, zap.String(kv.Key, kv.Value))
	}

	return &TelegramHandler{
		l: h.l.With(fields...),
	}
}

// MQTT handler

type MQTTHandler struct {
	l *zap.Logger
}

func (h *MQTTHandler) Error(msg string, err error) {
	h.l.Error(msg, zap.Error(err))
}

func (h *MQTTHandler) CreatingAlertHandler(c mqtt.HandlerConfig) {
	qos, _ := c.QoS.MarshalText()
	h.l.Debug("creating mqtt handler",
		zap.String("broker_name", c.BrokerName),
		zap.String("topic", c.Topic),
		zap.Bool("retained", c.Retained),
		zap.String("qos", string(qos)),
	)
}

func (h *MQTTHandler) HandlingEvent() {
	h.l.Debug("handling event")
}

func (h *MQTTHandler) WithContext(ctx ...keyvalue.T) mqtt.Diagnostic {
	fields := []zapcore.Field{}
	for _, kv := range ctx {
		fields = append(fields, zap.String(kv.Key, kv.Value))
	}

	return &MQTTHandler{
		l: h.l.With(fields...),
	}
}

// Talk handler

type TalkHandler struct {
	l *zap.Logger
}

func (h *TalkHandler) Error(msg string, err error) {
	h.l.Error(msg, zap.Error(err))
}

func (h *TalkHandler) WithContext(ctx ...keyvalue.T) talk.Diagnostic {
	fields := []zapcore.Field{}
	for _, kv := range ctx {
		fields = append(fields, zap.String(kv.Key, kv.Value))
	}

	return &TalkHandler{
		l: h.l.With(fields...),
	}
}

// Config handler

type ConfigOverrideHandler struct {
	l *zap.Logger
}

func (h *ConfigOverrideHandler) Error(msg string, err error) {
	h.l.Error(msg, zap.Error(err))
}

type ServerHandler struct {
	l *zap.Logger
}

func (h *ServerHandler) Error(msg string, err error, ctx ...keyvalue.T) {
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

func (h *ServerHandler) Info(msg string, ctx ...keyvalue.T) {
	if len(ctx) == 0 {
		h.l.Info(msg)
		return
	}

	if len(ctx) == 1 {
		el := ctx[0]
		h.l.Info(msg, zap.String(el.Key, el.Value))
		return
	}

	if len(ctx) == 2 {
		x := ctx[0]
		y := ctx[1]
		h.l.Info(msg, zap.String(x.Key, x.Value), zap.String(y.Key, y.Value))
		return
	}

	fields := make([]zapcore.Field, len(ctx))
	for i := 0; i < len(fields); i++ {
		kv := ctx[i]
		fields[i] = zap.String(kv.Key, kv.Value)
	}

	h.l.Info(msg, fields...)
}

func (h *ServerHandler) Debug(msg string, ctx ...keyvalue.T) {
	if len(ctx) == 0 {
		h.l.Debug(msg)
		return
	}

	if len(ctx) == 1 {
		el := ctx[0]
		h.l.Debug(msg, zap.String(el.Key, el.Value))
		return
	}

	if len(ctx) == 2 {
		x := ctx[0]
		y := ctx[1]
		h.l.Debug(msg, zap.String(x.Key, x.Value), zap.String(y.Key, y.Value))
		return
	}

	fields := make([]zapcore.Field, len(ctx))
	for i := 0; i < len(fields); i++ {
		kv := ctx[i]
		fields[i] = zap.String(kv.Key, kv.Value)
	}

	h.l.Debug(msg, fields...)
}

type ReplayHandler struct {
	l *zap.Logger
}

func (h *ReplayHandler) Error(msg string, err error, ctx ...keyvalue.T) {
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

func (h *ReplayHandler) Debug(msg string, ctx ...keyvalue.T) {
	if len(ctx) == 0 {
		h.l.Debug(msg)
		return
	}

	if len(ctx) == 1 {
		el := ctx[0]
		h.l.Debug(msg, zap.String(el.Key, el.Value))
		return
	}

	if len(ctx) == 2 {
		x := ctx[0]
		y := ctx[1]
		h.l.Debug(msg, zap.String(x.Key, x.Value), zap.String(y.Key, y.Value))
		return
	}

	fields := make([]zapcore.Field, len(ctx))
	for i := 0; i < len(fields); i++ {
		kv := ctx[i]
		fields[i] = zap.String(kv.Key, kv.Value)
	}

	h.l.Debug(msg, fields...)
}

// K8s handler

type K8sHandler struct {
	l *zap.Logger
}

func (h *K8sHandler) WithClusterContext(cluster string) k8s.Diagnostic {
	return &K8sHandler{
		l: h.l.With(zap.String("cluster_id", cluster)),
	}
}

// Swarm handler

type SwarmHandler struct {
	l *zap.Logger
}

func (h *SwarmHandler) WithClusterContext(cluster string) swarm.Diagnostic {
	return &SwarmHandler{
		l: h.l.With(zap.String("cluster_id", cluster)),
	}
}

// Deadman handler

type DeadmanHandler struct {
	l *zap.Logger
}

func (h *DeadmanHandler) ConfiguredGlobally() {
	h.l.Info("Deadman's switch is configured globally")
}

// NoAuth handler

type NoAuthHandler struct {
	l *zap.Logger
}

func (h *NoAuthHandler) FakedUserAuthentication(username string) {
	h.l.Warn("using noauth auth backend. Faked Authentication for user", zap.String("user", username))
}

func (h *NoAuthHandler) FakedSubscriptionUserToken() {
	h.l.Warn("using noauth auth backend. Faked authentication for subscription user token")
}

// Stats handler

type StatsHandler struct {
	l *zap.Logger
}

func (h *StatsHandler) Error(msg string, err error) {
	h.l.Error(msg, zap.Error(err))
}

// Template handler

//type Handler struct {
//	l *zap.Logger
//}
//
//func (h *Handler) Error(msg string, err error) {
//	h.l.Error(msg, zap.Error(err))
//}
//
//func (h *Handler) WithContext(ctx ...keyvalue.T) .Diagnostic {
//	fields := []zapcore.Field{}
//	for _, kv := range ctx {
//		fields = append(fields, zap.String(kv.Key, kv.Value))
//	}
//
//	return &Handler{
//		l: h.l.With(fields...),
//	}
//}
