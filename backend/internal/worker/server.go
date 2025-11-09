package worker

import (
	"time"

	"github.com/hibiken/asynq"
	"github.com/mahendrakalkura/saas-blueprint/internal/email"
	"github.com/rs/zerolog"
)

type Server struct {
	server    *asynq.Server
	mux       *asynq.ServeMux
	processor *TaskProcessor
}

func NewServer(redisAddr string, emailService *email.Service, sessionRepo SessionRepository, logger *zerolog.Logger) *Server {
	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
				// Exponential backoff: 1s, 2s, 4s, 8s, 16s, 32s
				return time.Duration(1<<uint(n)) * time.Second
			},
		},
	)

	mux := asynq.NewServeMux()
	processor := NewTaskProcessor(emailService, sessionRepo, logger)

	// Register task handlers
	mux.HandleFunc(TypeEmailWelcome, processor.ProcessEmailWelcome)
	mux.HandleFunc(TypeEmailVerification, processor.ProcessEmailVerification)
	mux.HandleFunc(TypeEmailPasswordReset, processor.ProcessEmailPasswordReset)
	mux.HandleFunc(TypeCleanupSessions, processor.ProcessCleanupSessions)

	return &Server{
		server:    server,
		mux:       mux,
		processor: processor,
	}
}

func (s *Server) Start() error {
	return s.server.Run(s.mux)
}

func (s *Server) Shutdown() {
	s.server.Shutdown()
}
