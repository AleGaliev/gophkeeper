package server

import "gophkeeper/internal/log"

type Config struct {
	Port        string
	Key         string
	DatabaseDSN string
	LogLevel    string
	Logger      log.Logger
	Migration   Migration
}

type Migration struct {
	Up   bool
	Down bool
}
