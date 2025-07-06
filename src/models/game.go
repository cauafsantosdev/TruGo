package models

import "sync"

var (
	Salas      = make(map[string]*Sala)
	SalasMutex sync.Mutex
)

type Sala struct {
	Jogo      EstadoJogo
	Jogadores []*Jogador
}

// STRUCT QUE GERENCIA O ESTADO DO JOGO
type EstadoJogo struct {
	Rodadas      []Rodada
	Time01Pontos int16
	Time02Pontos int16
}

type Rodada struct {
	// Apostas principais
	Flor   bool
	Envido bool
	Truco  bool

	// Apostas aumentadas
	ContraFlor  bool
	RealEnvido  bool
	FaltaEnvido bool
	Retruco     bool
	ValeQuatro  bool

	CartasJogada []CartaJogada
	VezJogador   *Jogador
	// Adicionar algo ainda
}
