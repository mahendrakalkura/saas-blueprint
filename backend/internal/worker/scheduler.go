package worker

import (
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
)

type Scheduler struct {
	scheduler *asynq.Scheduler
	logger    *zerolog.Logger
}

func NewScheduler(redisAddr string, logger *zerolog.Logger) *Scheduler {
	scheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{Addr: redisAddr},
		&asynq.SchedulerOpts{
			Logger: &asynqLogger{logger: logger},
		},
	)

	return &Scheduler{
		scheduler: scheduler,
		logger:    logger,
	}
}

func (s *Scheduler) RegisterTasks() error {
	// Schedule session cleanup to run every hour
	_, err := s.scheduler.Register(
		"@every 1h",
		asynq.NewTask(TypeCleanupSessions, nil),
		asynq.Queue("low"),
	)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to register session cleanup task")
		return err
	}

	s.logger.Info().Msg("Scheduled tasks registered successfully")
	return nil
}

func (s *Scheduler) Start() error {
	if err := s.scheduler.Run(); err != nil {
		return err
	}
	return nil
}

func (s *Scheduler) Shutdown() {
	s.scheduler.Shutdown()
}

// asynqLogger is an adapter to use zerolog with asynq
type asynqLogger struct {
	logger *zerolog.Logger
}

func (l *asynqLogger) Debug(args ...interface{}) {
	l.logger.Debug().Msgf("%v", args)
}

func (l *asynqLogger) Info(args ...interface{}) {
	l.logger.Info().Msgf("%v", args)
}

func (l *asynqLogger) Warn(args ...interface{}) {
	l.logger.Warn().Msgf("%v", args)
}

func (l *asynqLogger) Error(args ...interface{}) {
	l.logger.Error().Msgf("%v", args)
}

func (l *asynqLogger) Fatal(args ...interface{}) {
	l.logger.Fatal().Msgf("%v", args)
}
