package diagnostic

import (
	"bytes"
	"errors"

	"github.com/influxdata/kapacitor"
	//"github.com/influxdata/kapacitor/server"
	alertservice "github.com/influxdata/kapacitor/services/alert"
	"github.com/influxdata/kapacitor/services/alerta"
	"github.com/influxdata/kapacitor/services/config"
	"github.com/influxdata/kapacitor/services/hipchat"
	"github.com/influxdata/kapacitor/services/httpd"
	"github.com/influxdata/kapacitor/services/httppost"
	"github.com/influxdata/kapacitor/services/mqtt"
	"github.com/influxdata/kapacitor/services/opsgenie"
	"github.com/influxdata/kapacitor/services/pagerduty"
	"github.com/influxdata/kapacitor/services/pushover"
	"github.com/influxdata/kapacitor/services/reporting"
	"github.com/influxdata/kapacitor/services/sensu"
	"github.com/influxdata/kapacitor/services/slack"
	"github.com/influxdata/kapacitor/services/smtp"
	"github.com/influxdata/kapacitor/services/snmptrap"
	"github.com/influxdata/kapacitor/services/storage"
	"github.com/influxdata/kapacitor/services/talk"
	"github.com/influxdata/kapacitor/services/task_store"
	"github.com/influxdata/kapacitor/services/telegram"
	udfservice "github.com/influxdata/kapacitor/services/udf"
	"github.com/influxdata/kapacitor/services/victorops"
	"go.uber.org/zap"
)

type Service struct {
	logger *zap.Logger
}

func NewService() *Service {
	// TODO: change it
	l := zap.NewExample()
	return &Service{
		logger: l,
	}
}

func (s *Service) NewVictorOpsHandler() victorops.Diagnostic {
	return &VictorOpsHandler{
		l: s.logger.With(zap.String("service", "victorops")),
	}
}

func (s *Service) NewSlackHandler() slack.Diagnostic {
	return &SlackHandler{
		l: s.logger.With(zap.String("service", "slack")),
	}
}

func (s *Service) NewTaskStoreHandler() task_store.Diagnostic {
	return &TaskStoreHandler{
		l: s.logger.With(zap.String("service", "task_store")),
	}
}

func (s *Service) NewReportingHandler() reporting.Diagnostic {
	return &ReportingHandler{
		l: s.logger.With(zap.String("service", "reporting")),
	}
}

func (s *Service) NewStorageHandler() storage.Diagnostic {
	return &StorageHandler{
		l: s.logger.With(zap.String("service", "storage")),
	}
}

func (s *Service) NewHTTPDHandler() httpd.Diagnostic {
	return &HTTPDHandler{
		l: s.logger.With(zap.String("service", "http")),
	}
}

func (s *Service) NewAlertaHandler() alerta.Diagnostic {
	return &AlertaHandler{
		l: s.logger.With(zap.String("service", "alerta")),
	}
}

func (s *Service) NewKapacitorHandler() kapacitor.Diagnostic {
	return &KapacitorHandler{
		l: s.logger.With(zap.String("service", "kapacitor")), // TODO: what here
	}
}

func (s *Service) NewAlertServiceHandler() alertservice.Diagnostic {
	return &AlertServiceHandler{
		l: s.logger.With(zap.String("service", "alert")),
	}
}

func (s *Service) NewHipChatHandler() hipchat.Diagnostic {
	return &HipChatHandler{
		l: s.logger.With(zap.String("service", "hipchat")),
	}
}

func (s *Service) NewPagerDutyHandler() pagerduty.Diagnostic {
	return &PagerDutyHandler{
		l: s.logger.With(zap.String("service", "pagerduty")),
	}
}

func (s *Service) NewSMTPHandler() smtp.Diagnostic {
	return &SMTPHandler{
		l: s.logger.With(zap.String("service", "smtp")),
	}
}

func (s *Service) NewUDFServiceHandler() udfservice.Diagnostic {
	return &UDFServiceHandler{
		l: s.logger.With(zap.String("service", "udf")),
	}
}

func (s *Service) NewOpsGenieHandler() opsgenie.Diagnostic {
	return &OpsGenieHandler{
		l: s.logger.With(zap.String("service", "opsgenie")),
	}
}

func (s *Service) NewPushoverHandler() pushover.Diagnostic {
	return &PushoverHandler{
		l: s.logger.With(zap.String("service", "pushover")),
	}
}

func (s *Service) NewHTTPPostHandler() httppost.Diagnostic {
	return &HTTPPostHandler{
		l: s.logger.With(zap.String("service", "httppost")),
	}
}

func (s *Service) NewSensuHandler() sensu.Diagnostic {
	return &SensuHandler{
		l: s.logger.With(zap.String("service", "sensu")),
	}
}

func (s *Service) NewSNMPTrapHandler() snmptrap.Diagnostic {
	return &SNMPTrapHandler{
		l: s.logger.With(zap.String("service", "snmp")),
	}
}

func (s *Service) NewTelegramHandler() telegram.Diagnostic {
	return &TelegramHandler{
		l: s.logger.With(zap.String("service", "telegram")),
	}
}

func (s *Service) NewMQTTHandler() mqtt.Diagnostic {
	return &MQTTHandler{
		l: s.logger.With(zap.String("service", "mqtt")),
	}
}

func (s *Service) NewTalkHandler() talk.Diagnostic {
	return &TalkHandler{
		l: s.logger.With(zap.String("service", "talk")),
	}
}

func (s *Service) NewConfigOverrideHandler() config.Diagnostic {
	return &ConfigOverrideHandler{
		l: s.logger.With(zap.String("service", "config-override")),
	}
}

func (s *Service) NewServerHandler() *ServerHandler {
	return &ServerHandler{
		// TODO: what to make key here
		l: s.logger.With(zap.String("source", "srv")),
	}
}

func (s *Service) NewReplayHandler() *ReplayHandler {
	return &ReplayHandler{
		l: s.logger.With(zap.String("service", "replay")),
	}
}

func (s *Service) NewK8sHandler() *K8sHandler {
	return &K8sHandler{
		l: s.logger.With(zap.String("service", "kubernetes")),
	}
}

func (s *Service) NewSwarmHandler() *SwarmHandler {
	return &SwarmHandler{
		l: s.logger.With(zap.String("service", "swarm")),
	}
}

func (s *Service) NewDeadmanHandler() *DeadmanHandler {
	return &DeadmanHandler{
		l: s.logger.With(zap.String("service", "deadman")),
	}
}

func (s *Service) NewNoAuthHandler() *NoAuthHandler {
	return &NoAuthHandler{
		l: s.logger.With(zap.String("service", "noauth")),
	}
}

func (s *Service) NewStatsHandler() *StatsHandler {
	return &StatsHandler{
		l: s.logger.With(zap.String("service", "stats")),
	}
}

func (s *Service) NewUDPHandler() *UDPHandler {
	return &UDPHandler{
		l: s.logger.With(zap.String("service", "udp")),
	}
}

func (s *Service) NewInfluxDBHandler() *InfluxDBHandler {
	return &InfluxDBHandler{
		l: s.logger.With(zap.String("service", "influxdb")),
	}
}

func (s *Service) NewScraperHandler() *ScraperHandler {
	return &ScraperHandler{
		l:   s.logger.With(zap.String("service", "scraper")),
		buf: bytes.NewBuffer(nil),
	}
}

func (s *Service) NewAzureHandler() *ScraperHandler {
	return &ScraperHandler{
		l:   s.logger.With(zap.String("service", "azure")),
		buf: bytes.NewBuffer(nil),
	}
}

func (s *Service) NewConsulHandler() *ScraperHandler {
	return &ScraperHandler{
		l:   s.logger.With(zap.String("service", "consul")),
		buf: bytes.NewBuffer(nil),
	}
}

func (s *Service) NewDNSHandler() *ScraperHandler {
	return &ScraperHandler{
		l:   s.logger.With(zap.String("service", "dns")),
		buf: bytes.NewBuffer(nil),
	}
}

func (s *Service) NewEC2Handler() *ScraperHandler {
	return &ScraperHandler{
		l:   s.logger.With(zap.String("service", "ec2")),
		buf: bytes.NewBuffer(nil),
	}
}

func (s *Service) NewFileDiscoveryHandler() *ScraperHandler {
	return &ScraperHandler{
		l:   s.logger.With(zap.String("service", "file-discovery")),
		buf: bytes.NewBuffer(nil),
	}
}

func (s *Service) NewGCEHandler() *ScraperHandler {
	return &ScraperHandler{
		l:   s.logger.With(zap.String("service", "gce")),
		buf: bytes.NewBuffer(nil),
	}
}

func (s *Service) NewMarathonHandler() *ScraperHandler {
	return &ScraperHandler{
		l:   s.logger.With(zap.String("service", "marathon")),
		buf: bytes.NewBuffer(nil),
	}
}

func (s *Service) NewNerveHandler() *ScraperHandler {
	return &ScraperHandler{
		l:   s.logger.With(zap.String("service", "nerve")),
		buf: bytes.NewBuffer(nil),
	}
}

func (s *Service) NewServersetHandler() *ScraperHandler {
	return &ScraperHandler{
		l:   s.logger.With(zap.String("service", "serverset")),
		buf: bytes.NewBuffer(nil),
	}
}

func (s *Service) NewStaticDiscoveryHandler() *ScraperHandler {
	return &ScraperHandler{
		l:   s.logger.With(zap.String("service", "static-discovery")),
		buf: bytes.NewBuffer(nil),
	}
}

func (s *Service) NewTritonHandler() *ScraperHandler {
	return &ScraperHandler{
		l:   s.logger.With(zap.String("service", "triton")),
		buf: bytes.NewBuffer(nil),
	}
}

func (s *Service) NewStaticLevelHandler(level string, service string) (*StaticLevelHandler, error) {
	var ll LogLevel

	switch level {
	case "debug":
		ll = LLDebug
	case "error":
		ll = LLError
	case "fatal":
		ll = LLFatal
	case "info":
		ll = LLInfo
	case "warn":
		ll = LLWarn
	default:
		ll = LLInvalid
	}

	if ll == LLInvalid {
		return nil, errors.New("invalid log level")
	}

	return &StaticLevelHandler{
		l:     s.logger.With(zap.String("service", service)),
		level: ll,
	}, nil
}

func (s *Service) NewCmdHandler() *CmdHandler {
	return &CmdHandler{
		l: s.logger.With(zap.String("service", "run")),
	}
}
