// Package server は internal/server の公開ラッパーです。
package server

import "github.com/tomo1015/personal-kpi-tool/internal/server"

// Config は server.Config の公開型エイリアスです。
type Config = server.Config

// New は Config を受け取り Server を初期化して返します。
func New(cfg Config) *server.Server {
	return server.New(cfg)
}
