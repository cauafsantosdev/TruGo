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
	Rodadas []*Rodada
	Time01  Equipe
	Time02  Equipe
	Baralho []Cartas
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

type Equipe struct {
	Jogadores []*Jogador
	Pontos    int
}

func (n *Sala) PrepararJogo() {
	n.Jogo = NovoEstadoJogo()
	n.Jogadores = []*Jogador{}
}

func (n *EstadoJogo) EscolherEquipe(escolha string, jogador *Jogador) bool {
	switch escolha {
	case "TIME_01":
		if len(n.Time01.Jogadores) < 1 {
			n.Time01.Jogadores = append(n.Time01.Jogadores, jogador)
			return true
		}
	case "TIME_02":
		if len(n.Time02.Jogadores) < 1 {
			n.Time02.Jogadores = append(n.Time02.Jogadores, jogador)
			return true
		}
	}
	return false // TIME FULL (exception)
}

func NovoEstadoJogo() EstadoJogo {
	return EstadoJogo{
		Rodadas: []*Rodada{},
		Time01:  NovaEquipe(),
		Time02:  NovaEquipe(),
	}
}

func NovaEquipe() Equipe {
	return Equipe{
		Jogadores: []*Jogador{},
		Pontos:    0,
	}
}
