package models

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Jogador struct {
	Conn    *websocket.Conn
	ID      string
	Nome    string
	Mao     []Cartas
	Time    string
	TemFlor bool
	PontosEnvido int
}

type Time struct {
	Jogadores    []*Jogador
	Pontos       int
}

func NovoJogador(n string, conn *websocket.Conn) *Jogador {
	return &Jogador{
		Conn: conn,
		ID:   uuid.New().String(),
		Nome: n,
	}
}
