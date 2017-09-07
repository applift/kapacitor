package diagnostic

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path"

	//"github.com/influxdata/kapacitor/server"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type nopCloser struct {
	f io.Writer
}

func (c *nopCloser) Write(b []byte) (int, error) { return c.f.Write(b) }
func (c *nopCloser) Close() error                { return nil }

type Service struct {
	c Config

	logger *zap.Logger

	f      io.WriteCloser
	stdout io.Writer
	stderr io.Writer
}

func NewService(c Config, stdout, stderr io.Writer) *Service {
	return &Service{
		c:      c,
		stdout: stdout,
		stderr: stderr,
	}
}

func (s *Service) Open() error {
	var level zapcore.Level
	switch s.c.Level {
	case "INFO":
		level = zapcore.InfoLevel
	case "ERROR":
		level = zapcore.ErrorLevel
	case "DEBUG":
		level = zapcore.DebugLevel
	}
	p := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= level
	})

	switch s.c.File {
	case "STDERR":
		s.f = &nopCloser{f: s.stderr}
	case "STDOUT":
		s.f = &nopCloser{f: s.stdout}
	default:
		dir := path.Dir(s.c.File)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				return err
			}
		}

		f, err := os.OpenFile(s.c.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
		if err != nil {
			return err
		}
		s.f = f
	}

	out := zapcore.AddSync(s.f)
	encConfig := zap.NewProductionEncoderConfig()
	encConfig.EncodeTime = zapcore.EpochNanosTimeEncoder
	consoleEnc := zapcore.NewConsoleEncoder(encConfig)
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEnc, out, p),
	)
	l := zap.New(core)
	s.logger = l
	return nil
}

func (s *Service) Close() error {
	if s.f != nil {
		return s.f.Close()
	}
	return nil
}

func (s *Service) NewVictorOpsHandler() *VictorOpsHandler {
	return &VictorOpsHandler{
		l: s.logger.With(zap.String("service", "victorops")),
	}
}

func (s *Service) NewSlackHandler() *SlackHandler {
	return &SlackHandler{
		l: s.logger.With(zap.String("service", "slack")),
	}
}

func (s *Service) NewTaskStoreHandler() *TaskStoreHandler {
	return &TaskStoreHandler{
		l: s.logger.With(zap.String("service", "task_store")),
	}
}

func (s *Service) NewReportingHandler() *ReportingHandler {
	return &ReportingHandler{
		l: s.logger.With(zap.String("service", "reporting")),
	}
}

func (s *Service) NewStorageHandler() *StorageHandler {
	return &StorageHandler{
		l: s.logger.With(zap.String("service", "storage")),
	}
}

func (s *Service) NewHTTPDHandler() *HTTPDHandler {
	return &HTTPDHandler{
		l: s.logger.With(zap.String("service", "http")),
	}
}

func (s *Service) NewAlertaHandler() *AlertaHandler {
	return &AlertaHandler{
		l: s.logger.With(zap.String("service", "alerta")),
	}
}

func (s *Service) NewKapacitorHandler() *KapacitorHandler {
	return &KapacitorHandler{
		l: s.logger.With(zap.String("service", "kapacitor")), // TODO: what here
	}
}

func (s *Service) NewAlertServiceHandler() *AlertServiceHandler {
	return &AlertServiceHandler{
		l: s.logger.With(zap.String("service", "alert")),
	}
}

func (s *Service) NewHipChatHandler() *HipChatHandler {
	return &HipChatHandler{
		l: s.logger.With(zap.String("service", "hipchat")),
	}
}

func (s *Service) NewPagerDutyHandler() *PagerDutyHandler {
	return &PagerDutyHandler{
		l: s.logger.With(zap.String("service", "pagerduty")),
	}
}

func (s *Service) NewSMTPHandler() *SMTPHandler {
	return &SMTPHandler{
		l: s.logger.With(zap.String("service", "smtp")),
	}
}

func (s *Service) NewUDFServiceHandler() *UDFServiceHandler {
	return &UDFServiceHandler{
		l: s.logger.With(zap.String("service", "udf")),
	}
}

func (s *Service) NewOpsGenieHandler() *OpsGenieHandler {
	return &OpsGenieHandler{
		l: s.logger.With(zap.String("service", "opsgenie")),
	}
}

func (s *Service) NewPushoverHandler() *PushoverHandler {
	return &PushoverHandler{
		l: s.logger.With(zap.String("service", "pushover")),
	}
}

func (s *Service) NewHTTPPostHandler() *HTTPPostHandler {
	return &HTTPPostHandler{
		l: s.logger.With(zap.String("service", "httppost")),
	}
}

func (s *Service) NewSensuHandler() *SensuHandler {
	return &SensuHandler{
		l: s.logger.With(zap.String("service", "sensu")),
	}
}

func (s *Service) NewSNMPTrapHandler() *SNMPTrapHandler {
	return &SNMPTrapHandler{
		l: s.logger.With(zap.String("service", "snmp")),
	}
}

func (s *Service) NewTelegramHandler() *TelegramHandler {
	return &TelegramHandler{
		l: s.logger.With(zap.String("service", "telegram")),
	}
}

func (s *Service) NewMQTTHandler() *MQTTHandler {
	return &MQTTHandler{
		l: s.logger.With(zap.String("service", "mqtt")),
	}
}

func (s *Service) NewTalkHandler() *TalkHandler {
	return &TalkHandler{
		l: s.logger.With(zap.String("service", "talk")),
	}
}

func (s *Service) NewConfigOverrideHandler() *ConfigOverrideHandler {
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
