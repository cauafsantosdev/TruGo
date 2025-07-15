package models

import "sync"

var (
	Salas      = make(map[string]*Sala)
	SalasMutex sync.Mutex
)

type Sala struct {
	Status    string
	Jogo      EstadoJogo
	Jogadores []*Jogador
}

// STRUCT QUE GERENCIA O ESTADO DO JOGO
type EstadoJogo struct {
	Estado     string
	Rodada     *Rodada
	Time01     Equipe
	Time02     Equipe
	Baralho    []Cartas
	JogadorMao *Jogador
	IdxJogador int
}

type Aposta struct {
	Tipo     string
	Estado   string
	ParaTime string
}

type Rodada struct {
	ApostaAtual Aposta

	Flor   bool
	Envido bool
	Truco  bool

	// Apostas aumentadas
	ContraFlor        bool
	ContraFlorAlResto bool
	RealEnvido        bool
	FaltaEnvido       bool
	Retruco           bool
	ValeQuatro        bool

	CartasJogadas   []CartaJogada
	CartasEmDisputa []CartaJogada
	VezJogador      *Jogador

	ValorDaMao     int
	Rodada         []int
	IdxJogador     int
	TimeDaMao      string
	CadeiaEnvido   []string
	EstadoFlor     string
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
			jogador.Time = "TIME_01"
			return true
		}
	case "TIME_02":
		if len(n.Time02.Jogadores) < 1 {
			n.Time02.Jogadores = append(n.Time02.Jogadores, jogador)
			jogador.Time = "TIME_02"
			return true
		}
	}
	return false // TIME FULL (exception)
}

func NovoEstadoJogo() EstadoJogo {
	return EstadoJogo{
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
