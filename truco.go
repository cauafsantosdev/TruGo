package main

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"
	"slices"
)

// STRUCTS PRINCIPAIS

// Carta
type Carta struct {
	Valor int
	Naipe string
	Forca int
}

// print da carta
func (c Carta) String() string {
	return fmt.Sprintf("%d de %s", c.Valor, c.Naipe)
}

// Jogador
type Jogador struct {
	ID string
	Nome string
	Time int // 0 para Time 1, 1 para Time 2
	Mao []Carta
	PontosEnvido int // pontos de envido do jogador na mão atual
	PontosFlor int // pontos da flor, caso tenha
}

// print do jogador
func (j Jogador) String() string {
	return fmt.Sprintf("%s (Time %d)", j.Nome, j.Time+1)
}

// Acao = qualquer jogada
type Acao struct {
	Tipo string
	JogadorID string
	Valor any // Carta, "QUERO", etc.
}

// carta na mesa
type CartaJogada struct {
	Jogador *Jogador
	Carta   Carta
}

// nome auto-explicativo
func CriarBaralho() []Carta {
	naipes := []string{"Copas", "Espadas", "Paus", "Ouros"}
	valores := []int{1, 2, 3, 4, 5, 6, 7, 10, 11, 12}
	baralho := make([]Carta, 0, 40)

	for _, naipe := range naipes {
		for _, valor := range valores {
			carta := Carta{Valor: valor, Naipe: naipe}

			switch {
			// Manilhas
			case valor == 1 && naipe == "Espadas":
				carta.Forca = 13
			case valor == 1 && naipe == "Paus":
				carta.Forca = 12
			case valor == 7 && naipe == "Espadas":
				carta.Forca = 11
			case valor == 7 && naipe == "Ouros":
				carta.Forca = 10
			// Cartas Comuns
			case valor == 3:
				carta.Forca = 9
			case valor == 2:
				carta.Forca = 8
			case valor == 1:
				carta.Forca = 7
			case valor == 12:
				carta.Forca = 6
			case valor == 11:
				carta.Forca = 5
			case valor == 10:
				carta.Forca = 4
			case valor == 7:
				carta.Forca = 3
			case valor == 6:
				carta.Forca = 2
			case valor == 5:
				carta.Forca = 1
			case valor == 4:
				carta.Forca = 0
			}
			baralho = append(baralho, carta)
		}
	}
	return baralho
}

type EstadoDoJogo string

// estados do jogo
const (
	AguardandoJogadores EstadoDoJogo = "AGUARDANDO_JOGADORES"
	AguardandoAcao EstadoDoJogo = "AGUARDANDO_ACAO"
	AguardandoRespostaTruco EstadoDoJogo = "AGUARDANDO_RESPOSTA_TRUCO"
	AguardandoRespostaEnvido EstadoDoJogo = "AGUARDANDO_RESPOSTA_ENVIDO"
	AguardandoRespostaFlor EstadoDoJogo = "AGUARDANDO_RESPOSTA_FLOR"
	AguardandoRespostaFamiliaReal EstadoDoJogo = "AGUARDANDO_RESPOSTA_FAMILIA_REAL"
	MaoFinalizada EstadoDoJogo = "MAO_FINALIZADA"
	FimDeJogo EstadoDoJogo = "FIM_DE_JOGO"
)

// TRUCO
type Jogo struct {
	ID string
	Estado EstadoDoJogo
	Baralho []Carta
	Jogadores []*Jogador
	Placar [2]int
	mutex sync.Mutex // garante que apenas uma ação seja processada por vez

	// controle da mao
	IndiceDoDistribuidor int
	TurnoDoJogador int
	RodadaAtual int
	Rodada [3]int // -1: Não jogada, 0: Time 0, 1: Time 1, 2: Empate
	PrimeiroJogadorDaRodada int
	CartasNaMesa []CartaJogada
	UltimaCarta int

	// controle das apostas
	PodeChamarEnvido bool
	CadeiaDeEnvido []string // ["ENVIDO", "REAL_ENVIDO"]
	TimeComAFalaDoTruco int
	ValorDaMao int // 1, 2, 3 ou 4
	PontosApostaRecusada int // pontos a serem ganhos se a aposta for recusada
	PontosApostaAceita int // pontos a serem ganhos se a aposta for aceita
	JogadorComFlor []*Jogador
	EstadoFlor string

	// Familia Real
	FamiliaReal bool
	JogadorComFamiliaReal []*Jogador
}

// FUNÇÕES DE FUNCIONAMENTO DO JOGO

// cria uma nova instância de jogo
func CriarJogo() *Jogo {
	j := &Jogo{
		ID: fmt.Sprintf("jogo_%d", rand.Intn(10000)),
		Baralho: CriarBaralho(),
		Estado: AguardandoJogadores,
		Jogadores: make([]*Jogador, 0, 2),
		IndiceDoDistribuidor: -1,
		FamiliaReal: true,
	}

	for i := range j.Rodada {
		j.Rodada[i] = -1
	}
	return j
}

// adiciona um jogador no jogo
func (j *Jogo) AdicionarJogador(jogador *Jogador) {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	jogador.Time = len(j.Jogadores) % 2
	j.Jogadores = append(j.Jogadores, jogador)

	// inicia o jogo quando tiver 2 jogadores
	if len(j.Jogadores) == 2 {
		j.proximaMao()
	}
}

// passa para a próxima mão
func (j *Jogo) proximaMao() {
	log.Println("!!!PREPARANDO NOVA MÃO!!!")
	// troca o mao
	j.IndiceDoDistribuidor = (j.IndiceDoDistribuidor + 1) % len(j.Jogadores)
	j.TurnoDoJogador = (j.IndiceDoDistribuidor + 1) % len(j.Jogadores)
	j.PrimeiroJogadorDaRodada = j.TurnoDoJogador
	
	// reseta todos os estados de aposta e da mão para os valores iniciais
	j.ValorDaMao = 1
	j.TimeComAFalaDoTruco = -1
	j.RodadaAtual = 1
	for i := range j.Rodada {
		j.Rodada[i] = -1
	}
	j.CartasNaMesa = []CartaJogada{}
	j.PodeChamarEnvido = true
	j.CadeiaDeEnvido = []string{}
	j.PontosApostaAceita = 0
	j.PontosApostaRecusada = 0
	j.JogadorComFlor = nil
	j.EstadoFlor = ""
	j.JogadorComFamiliaReal = nil
	j.UltimaCarta = 0

	// cria um novo baralho e embaralha (avaliar possibilidade de não criar um novo baralho a cada mão)
	baralho := j.Baralho
	rand.Shuffle(len(baralho), func(i, j int) {
		baralho[i], baralho[j] = baralho[j], baralho[i]
	})

	// dá as cartas
	for _, jogador := range j.Jogadores {
		jogador.Mao = baralho[0:3]
		baralho = baralho[3:]
		j.UltimaCarta += 3

		if jogador.Mao[0].Valor == 4 && jogador.Mao[1].Valor == 4 && jogador.Mao[2].Valor == 4 {
			jogador.Mao = baralho[0:3]
			baralho = baralho[3:]
			j.UltimaCarta += 3
		}

		// depois de dar as cartas calcula os pontos de envido e verifica se há flor
		jogador.PontosEnvido = j.calcularPontosEnvido(jogador)
		if j.temFlor(jogador) {
			jogador.PontosFlor = j.calcularPontosFlor(jogador)
			j.JogadorComFlor = append(j.JogadorComFlor, jogador)
		}

		// verifica se o jogador tem familia real, caso não tenha sido negada no jogo
		if j.FamiliaReal && j.temFamiliaReal(jogador) {
			j.JogadorComFamiliaReal = append(j.JogadorComFamiliaReal, jogador)
		}
	}

	j.Estado = AguardandoAcao
}

// core do jogo, identiica a ação e chama sua função específica
func (j *Jogo) ProcessarAcao(acao Acao) error {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	
	var err error
	// aciona a função específica da ação
	switch acao.Tipo {
	case "JOGAR_CARTA":
		err = j.jogarCarta(acao)
	case "TRUCO", "RETRUCO", "VALE_CUATRO":
		err = j.cantarTruco(acao)
	case "ENVIDO", "REAL_ENVIDO", "FALTA_ENVIDO":
		err = j.cantarEnvido(acao)
	case "FLOR", "CONTRA_FLOR", "CONTRA_FLOR_AL_RESTO":
		err = j.cantarFlor(acao)
	case "TEM_FAMILIA_REAL":
		err = j.perguntarFamiliaReal(acao)
	case "FAMILIA_REAL":
		err = j.trocarFamiliaReal(acao)
	case "QUERO", "NAO_QUERO":
		err = j.responder(acao)
	case "IR_AO_MAZO":
		err = j.irAoMazo(acao)
	}

	if err != nil {
		return err
	}

	// verifica se a ação encerra a mão ou o jogo
	if j.Estado == MaoFinalizada {
		if j.Placar[0] >= 30 || j.Placar[1] >= 30 {
			j.Estado = FimDeJogo
			log.Println("!!!!!!!!!! FIM DE JOGO !!!!!!!!!!")
		} else {
			time.Sleep(1 * time.Second)
			j.proximaMao()
		}
	}
	return nil
}

// FUNÇÕES DE CANTOS, JOGAR CARTA, QUERO NÃO QUERO

func (j *Jogo) jogarCarta(acao Acao) error {
	jogador := j.getJogadorPorID(acao.JogadorID)
	cartaValor := acao.Valor.(map[string]any)
	
	cartaJogada := Carta{Valor: cartaValor["Valor"].(int), Naipe: cartaValor["Naipe"].(string)}

	// pega o index da carta jogada
	idxCarta := -1
	for i, c := range jogador.Mao {
		if c.Valor == cartaJogada.Valor && c.Naipe == cartaJogada.Naipe {
			idxCarta = i
			break
		}
	}
	
	cartaReal := jogador.Mao[idxCarta] // pega a carta com a força
	log.Printf("%s jogou a carta: %s", jogador.Nome, cartaReal)
	j.CartasNaMesa = append(j.CartasNaMesa, CartaJogada{Jogador: jogador, Carta: cartaReal}) // bota a carta na mesa
	jogador.Mao = slices.Delete(jogador.Mao, idxCarta, idxCarta+1) // tira da mão do jogador

	// se ninguém chamou envido na primeira rodada não tem mais envido na mão
	if j.PodeChamarEnvido == true && j.RodadaAtual == 1 && len(j.CartasNaMesa) == len(j.Jogadores) {
		j.PodeChamarEnvido = false
	}

	// se todos jogaram, resolve a rodada
	if len(j.CartasNaMesa) == len(j.Jogadores) {
		j.resolverRodada()
	} else {
		j.avancarTurno()
	}
	return nil
}

func (j *Jogo) cantarTruco(acao Acao) error {
	// se chamar com 29 volta pra 15
	jogador := j.getJogadorPorID(acao.JogadorID)
	if j.Placar[jogador.Time] == 29 {
		log.Printf("RATIOU! %s cantou %s com 29 pontos e voltou para 15.", jogador.Nome, acao.Tipo)
		j.Placar[jogador.Time] = 15
	}
	
	// aumenta o valor da mão com os cantos
	switch acao.Tipo {
	case "TRUCO":
		j.ValorDaMao = 2
		j.PontosApostaRecusada = 1
	case "RETRUCO":
		j.ValorDaMao = 3
		j.PontosApostaRecusada = 2
	case "VALE_CUATRO":
		j.ValorDaMao = 4
		j.PontosApostaRecusada = 3
	}

	log.Printf("%s cantou %s!", jogador.Nome, acao.Tipo)
	j.Estado = AguardandoRespostaTruco
	j.avancarTurnoParaOponente(jogador.ID)
	return nil
}

func (j *Jogo) cantarEnvido(acao Acao) error {
	// se chamar com 29 volta pra 15
	jogador := j.getJogadorPorID(acao.JogadorID)
	if j.Placar[jogador.Time] == 29 {
		log.Printf("RATIOU! %s cantou %s com 29 pontos e voltou para 15.", jogador.Nome, acao.Tipo)
		j.Placar[jogador.Time] = 15
	}
	
	// adiciona o canto na cadeia de envido
	j.CadeiaDeEnvido = append(j.CadeiaDeEnvido, acao.Tipo)
	log.Printf("%s cantou %s!", acao.JogadorID, acao.Tipo)
	
	j.Estado = AguardandoRespostaEnvido
	j.avancarTurnoParaOponente(acao.JogadorID)
	return nil
}

func (j *Jogo) cantarFlor(acao Acao) error {
	jogador := j.getJogadorPorID(acao.JogadorID)

	// Anula qualquer disputa de envido que estivesse em andamento.
	j.CadeiaDeEnvido = []string{}
	j.PodeChamarEnvido = false

	switch acao.Tipo {
	case "FLOR":
		log.Printf("%s cantou FLOR!", jogador.Nome)

		oponente := j.getOponente(jogador)
		// verifica Contra-Flor
		if len(j.JogadorComFlor) >= 2{
			log.Printf("%s também tem flor! A bola está com ele para responder.", oponente.Nome)
			j.Estado = AguardandoRespostaFlor
			j.EstadoFlor = "FLOR_CANTADA"
			j.PontosApostaAceita = 3
		} else {
			// se o oponente não tem flor quem chamou ganha 3 e segue o jogo
			log.Printf("%s ganha 3 pontos pela Flor.", jogador.Nome)
			j.Placar[jogador.Time] += 3
			j.Estado = AguardandoAcao
		}

	case "CONTRA_FLOR":
		log.Printf("%s respondeu com CONTRA-FLOR!", jogador.Nome)
		j.Estado = AguardandoRespostaFlor
		j.EstadoFlor = "CONTRA_FLOR_CANTADA"
		j.PontosApostaAceita = 6   // se aceito vale 6 pontos
		j.PontosApostaRecusada = 3 // se recusado  quem cantou ganha 3

	case "CONTRA_FLOR_AL_RESTO":
		log.Printf("%s escalou para CONTRA-FLOR AL RESTO!", jogador.Nome)

		// define os pontos para recusa, dependendo do que foi cantado anteriormente
		if j.EstadoFlor == "FLOR_CANTADA" {
			j.PontosApostaRecusada = 3
		} else { // se veio de uma Contra-Flor
			j.PontosApostaRecusada = 6
		}		

		j.Estado = AguardandoRespostaFlor
		j.EstadoFlor = "CONTRA_FLOR_AL_RESTO_CANTADA"
		
	}

	j.avancarTurnoParaOponente(jogador.ID)
	return nil
}

func (j *Jogo) perguntarFamiliaReal(acao Acao) error {
	jogador := j.getJogadorPorID(acao.JogadorID)

	log.Printf("%s perguntou se tem FAMILIA REAL!", jogador.Nome)
	j.Estado = AguardandoRespostaFamiliaReal
	j.avancarTurnoParaOponente(acao.JogadorID)
	return nil
}

func (j *Jogo) responder(acao Acao) error {
	jogador := j.getJogadorPorID(acao.JogadorID)
	oponente := j.getOponente(jogador)
	resposta := acao.Valor.(string)

	estado := j.Estado
	
	switch estado {
	case AguardandoRespostaTruco:
		if resposta == "NAO_QUERO" {
			log.Printf("%s não quis o %s. %s ganha %d ponto(s).", jogador.Nome, "TRUCO", oponente.Nome, j.PontosApostaRecusada)
			j.Placar[oponente.Time] += j.PontosApostaRecusada
			j.Estado = MaoFinalizada
		} else { // QUERO
			log.Printf("%s aceitou! A mão agora vale %d pontos.", jogador.Nome, j.ValorDaMao)
			// quem respondeu pode aumentar
			j.TimeComAFalaDoTruco = jogador.Time
		}
	case AguardandoRespostaEnvido:
		j.PodeChamarEnvido = false
		if resposta == "NAO_QUERO" {
			pontos := j.calcularPontosRecusaEnvido()
			log.Printf("%s não quis o ENVIDO. %s ganha %d ponto(s).", jogador.Nome, oponente.Nome, pontos)
			j.Placar[oponente.Time] += pontos
		} else { // QUERO
			log.Printf("%s aceitou o ENVIDO! Comparando pontos...", jogador.Nome)
			j.resolverDisputaEnvido()
		}
	case AguardandoRespostaFlor:
		if resposta == "NAO_QUERO" {
			log.Printf("%s não quis a aposta da Flor. %s ganha %d ponto(s).", jogador.Nome, oponente.Nome, j.PontosApostaRecusada)
			j.Placar[oponente.Time] += j.PontosApostaRecusada
		} else { // QUERO
			log.Printf("%s aceitou a disputa da Flor! Valendo %d ponto(s).", jogador.Nome, j.PontosApostaAceita)
			j.resolverDisputaFlor()
		}
	case AguardandoRespostaFamiliaReal:
		if resposta == "NAO_QUERO" { // NÃO TEM
			log.Printf("%s determinou que NÃO tem FAMILIA REAL.", jogador.Nome)
			j.FamiliaReal = false
		} else { // TEM
			log.Printf("%s determinou que TEM FAMILIA REAL.", jogador.Nome)
		}
	}
	
	j.Estado = AguardandoAcao // volta pro estado normal de ação depois da resposta
	j.TurnoDoJogador = j.getIndiceJogador(oponente.ID) // volta o turno pra quem cantou
	return nil
}

func (j *Jogo) trocarFamiliaReal(acao Acao) error {
	jogador := j.getJogadorPorID(acao.JogadorID)
	jogador.Mao = j.Baralho[j.UltimaCarta: j.UltimaCarta+3]
	j.UltimaCarta += 3

	return nil
}

func (j *Jogo) irAoMazo(acao Acao) error {
    jogador := j.getJogadorPorID(acao.JogadorID)
    oponente := j.getOponente(jogador)
    log.Printf("%s foi ao mazo! %s ganha %d ponto(s).", jogador.Nome, oponente.Nome, j.ValorDaMao)
    j.Placar[oponente.Time] += j.ValorDaMao
    j.Estado = MaoFinalizada
    return nil
}

// RESOLUÇÃO DAS DISPUTAS

func (j *Jogo) resolverRodada() {
	log.Printf("Resolvendo rodada %d...", j.RodadaAtual)
	vencedorRodada := j.CartasNaMesa[0].Jogador
	maiorCarta := j.CartasNaMesa[0].Carta
	empate := false

	// encontra a carta mais forte
	for i := 1; i < len(j.CartasNaMesa); i++ {
		jogadaAtual := j.CartasNaMesa[i]
		if jogadaAtual.Carta.Forca > maiorCarta.Forca {
			maiorCarta = jogadaAtual.Carta
			vencedorRodada = jogadaAtual.Jogador
			empate = false
		} else if jogadaAtual.Carta.Forca == maiorCarta.Forca {
			empate = true
		}
	}
	
	if empate {
		j.Rodada[j.RodadaAtual-1] = 2
		log.Printf("Rodada %d empatou.", j.RodadaAtual)
	} else {
		j.Rodada[j.RodadaAtual-1] = vencedorRodada.Time
		log.Printf("%s venceu a rodada %d com %s.", vencedorRodada.Nome, j.RodadaAtual, maiorCarta)
	}

	// limpa CartasNaMesa e passa para a próxima rodada
	j.CartasNaMesa = []CartaJogada{}
	j.RodadaAtual++
	
	// troca o mão
	if !empate { // jogador que venceu, em caso de vitória
		j.TurnoDoJogador = j.getIndiceJogador(vencedorRodada.ID)
	} else { // o mão segue o mesmo, em caso de empate
		j.TurnoDoJogador = j.PrimeiroJogadorDaRodada
	}
	j.PrimeiroJogadorDaRodada = j.TurnoDoJogador

	// verifica se a mão já foi resolvida
	j.resolverMao()
}

func (j *Jogo) resolverMao() {
	// calculo das vitorias de cada time
	vitoriasTime0, vitoriasTime1 := 0, 0
	for _, resultadoRodada := range j.Rodada {
		switch resultadoRodada {
		case 0:
			vitoriasTime0++
		case 1:
			vitoriasTime1++
		}
	}

	vencedorDaMao := -1

	// caso 1: vitória normal
	if vitoriasTime0 >= 2 {
		vencedorDaMao = 0 
	} else if vitoriasTime1 >= 2 {
		vencedorDaMao = 1 
	}
	
	// caso 2: empates
	if vencedorDaMao == -1 {
		if j.Rodada[0] == 2 && j.Rodada[1] == 2 && j.Rodada[2] == 2 { // empate triplo
			vencedorDaMao = j.Jogadores[(j.IndiceDoDistribuidor+1) % len(j.Jogadores)].Time
			// ganha o mão

		} else if j.Rodada[0] == 2 && j.Rodada[1] == 2 { // empatou as duas primeiras
			vencedorDaMao = j.Rodada[2]
			// quem ganhou a terceira leva

		} else if j.Rodada[0] == 2 { // empatou a primeira
			if j.Rodada[1] != -1 && j.Rodada[1] != 2 {
				vencedorDaMao = j.Rodada[1]
				} // quem ganhou a segunda leva

		} else if j.Rodada[1] == 2 { // empatou a segunda
			if j.Rodada[0] != -1 {
				vencedorDaMao = j.Rodada[0]
				} // quem ganhou a primeira leva

		} else if j.Rodada[0] != -1 && j.Rodada[1] != -1 && j.Rodada[2] == 2 { // empatou a terceira
			if j.Rodada[0] != -1 {
				vencedorDaMao = j.Rodada[0]
				} // quem ganhou a primeira leva
		}
	}

	// se um time venceu finaliza a mão
	if vencedorDaMao != -1 {
		log.Printf("Time %d venceu a mão e ganhou %d ponto(s).", vencedorDaMao+1, j.ValorDaMao)
		j.Placar[vencedorDaMao] += j.ValorDaMao
		j.Estado = MaoFinalizada
	}
}

func (j *Jogo) resolverDisputaEnvido() {
	pontosTime0 := j.Jogadores[0].PontosEnvido
	pontosTime1 := j.Jogadores[1].PontosEnvido
	vencedorEnvido := -1

	if pontosTime0 > pontosTime1 {
		vencedorEnvido = 0
	} else if pontosTime1 > pontosTime0 {
		vencedorEnvido = 1
	} else { // empate, o mão vence
		vencedorEnvido = j.Jogadores[(j.IndiceDoDistribuidor+1) % len(j.Jogadores)].Time
	}
	
	pontosGanhos := j.calcularPontosAceiteEnvido(vencedorEnvido)
	log.Printf("Disputa de Envido: %s (%d) vs %s (%d). Vencedor: Time %d. Ganhando %d ponto(s).",
		j.Jogadores[0].Nome, pontosTime0, j.Jogadores[1].Nome, pontosTime1, vencedorEnvido+1, pontosGanhos)
	j.Placar[vencedorEnvido] += pontosGanhos
}

func (j *Jogo) resolverDisputaFlor() {
	pontosTime0 := j.Jogadores[0].PontosFlor
	pontosTime1 := j.Jogadores[1].PontosFlor
	vencedorFlor := -1

	if pontosTime0 > pontosTime1 {
		vencedorFlor = 0
	} else if pontosTime1 > pontosTime0 {
		vencedorFlor = 1
	} else { // empate, o mão vence
		vencedorFlor = j.Jogadores[(j.IndiceDoDistribuidor+1) % len(j.Jogadores)].Time
	}
	
	if j.EstadoFlor == "CONTRA_FLOR_AL_RESTO_CANTADA"{
		oponente := j.getOponente(j.Jogadores[vencedorFlor])
		j.PontosApostaAceita = 30 - j.Placar[oponente.Time]
	}

	log.Printf("Disputa de Envido: %s (%d) vs %s (%d). Vencedor: Time %d. Ganhando %d ponto(s).",
		j.Jogadores[0].Nome, pontosTime0, j.Jogadores[1].Nome, pontosTime1, vencedorFlor+1, j.PontosApostaAceita)
	j.Placar[vencedorFlor] += j.PontosApostaAceita
}


// FUNÇÕES AUXILIARES

func (j *Jogo) avancarTurno() {
	j.TurnoDoJogador = (j.TurnoDoJogador + 1) % len(j.Jogadores)
}

func (j *Jogo) avancarTurnoParaOponente(idAtuante string) {
	oponente := j.getOponente(j.getJogadorPorID(idAtuante))
	j.TurnoDoJogador = j.getIndiceJogador(oponente.ID)
}

func (j *Jogo) getJogadorPorID(id string) *Jogador {
	for _, p := range j.Jogadores {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (j *Jogo) getOponente(jogador *Jogador) *Jogador {
	if jogador.Time == 0 {
		return j.Jogadores[1]
	}
	return j.Jogadores[0]
}

func (j *Jogo) getIndiceJogador(id string) int {
	for i, p := range j.Jogadores {
		if p.ID == id {
			return i
		}
	}
	return -1
}

func (j *Jogo) temFlor(jogador *Jogador) bool {
	return jogador.Mao[0].Naipe == jogador.Mao[1].Naipe && jogador.Mao[1].Naipe == jogador.Mao[2].Naipe
}

func (j *Jogo) temFamiliaReal(jogador *Jogador) bool {
	familiaReal := true

	for _, carta := range jogador.Mao {
		if carta.Valor < 10 {
			familiaReal = false
		}
	}

	return familiaReal
}

func (j *Jogo) calcularPontosEnvido(jogador *Jogador) int {
	// cria dicionario (key = naipe, value = valor das cartas do naipe)
	naipes := make(map[string][]int)

	// define a pontuação caso o jogador não tenha 2 do mesmo naipe
	maiorCarta := 0
	for _, c := range jogador.Mao {
		valor := 0
		switch c.Valor {
		case 10, 11, 12:
			valor = 0
		default:
			valor = c.Valor
		}
		naipes[c.Naipe] = append(naipes[c.Naipe], valor)
		if valor > maiorCarta {
			maiorCarta = valor
		}
	}
	
	// define a pontuação de envido, caso tenha
	maiorPontuacao := 0
	for _, valores := range naipes {
		if len(valores) >= 2 {
			sort.Sort(sort.Reverse(sort.IntSlice(valores)))
			pontuacao := valores[0] + valores[1] + 20
			if pontuacao > maiorPontuacao {
				maiorPontuacao = pontuacao
			}
		}
	}
	
	if maiorPontuacao > 0 {
		return maiorPontuacao
	}
	return maiorCarta
}

func (j *Jogo) calcularPontosFlor(jogador *Jogador) int {
	pontos := 20

	for _, c := range jogador.Mao {
		switch c.Valor {
		case 10, 11, 12:
			pontos += 0
		default:
			pontos += c.Valor
		}
	}

	return pontos
}

func (j *Jogo) calcularPontosRecusaEnvido() int {
	pontos := 0
	for i, canto := range j.CadeiaDeEnvido {
		switch canto {
		case "ENVIDO":
			pontos += 1
		case "REAL_ENVIDO":
			pontos += 1
		case "FALTA_ENVIDO":
			switch pontos {
			case 0:
				pontos += 1
			case 1:
				switch j.CadeiaDeEnvido[i-1] {
				case "ENVIDO":
					pontos += 1
				case "REAL_ENVIDO":
					pontos += 2
				}
			case 2:
				pontos += 3
			}
		}
	}
	return pontos
}

func (j *Jogo) calcularPontosAceiteEnvido(timeVencedor int) int {
	querFaltaEnvido := false
	pontos := 0
	for _, canto := range j.CadeiaDeEnvido {
		switch canto {
		case "ENVIDO": pontos += 2
		case "REAL_ENVIDO": pontos += 3
		case "FALTA_ENVIDO": querFaltaEnvido = true
		}
	}
	
	if querFaltaEnvido {
		placarDoAdversario := j.Placar[1 - timeVencedor]
		return 30 - placarDoAdversario
	}

	return pontos
}

// FUNÇÕES DE TESTE DO CÓDIGO (GERADAS PELO GPT)

// mostra o estado do jogo e os stats dos jogadores
func (j *Jogo) gerarStringEstado() string {
	estado := "==================================================\n"
	estado += fmt.Sprintf("ESTADO ATUAL: %s\n", j.Estado)
	estado += fmt.Sprintf("PLACAR: Time 1: %d x Time 2: %d\n", j.Placar[0], j.Placar[1])
	estado += fmt.Sprintf("Mão atual vale: %d ponto(s)\n", j.ValorDaMao)
	if j.Estado != FimDeJogo && j.Estado != AguardandoJogadores {
		estado += fmt.Sprintf("Turno de: %s\n", j.Jogadores[j.TurnoDoJogador])
	}
	estado += "--- Rodadas ---\n"
	for i, r := range j.Rodada {
		var res string
		switch r {
		case -1:
			res = "Não jogada"
		case 0:
			res = "Time 1"
		case 1:
			res = "Time 2"
		case 2:
			res = "Empate"
		}
		estado += fmt.Sprintf("  Rodada %d: %s\n", i+1, res)
	}
	estado += "--- Jogadores ---\n"
	for _, jogador := range j.Jogadores {
		estado += fmt.Sprintf("  > %s (Time %d)\n    Mão: %v\n    Envido: %d, Flor: %d\n", jogador.Nome, jogador.Time+1, jogador.Mao, jogador.PontosEnvido, jogador.PontosFlor)
	}
	if len(j.CartasNaMesa) > 0 {
		estado += "--- Cartas na mesa ---\n"
		for _, cj := range j.CartasNaMesa {
			estado += fmt.Sprintf("  > %s jogou %s\n", cj.Jogador.Nome, cj.Carta)
		}
	}
	if len(j.CadeiaDeEnvido) > 0 {
		estado += fmt.Sprintf("Cadeia de Envido: %v\n", j.CadeiaDeEnvido)
	}
	if j.EstadoFlor != "" {
		estado += fmt.Sprintf("Estado da Flor: %s\n", j.EstadoFlor)
	}
	estado += "==================================================\n"
	return estado
}


// TÁ ERRADO PRA KRL ESSA FUNÇÃO, MAS FOI CULPA DO GPT
func SimularPartidaCompleta(j *Jogo) {
	jogador1 := &Jogador{ID: "j1", Nome: "Ana"}
	jogador2 := &Jogador{ID: "j2", Nome: "Beto"}
	j.AdicionarJogador(jogador1)
	j.AdicionarJogador(jogador2)

	darCartas := func(c1, c2 []Carta) { // dá cartas manualmente
		j.proximaMao()
		jogador1.Mao = c1
		jogador2.Mao = c2
		j.UltimaCarta += 6
		jogador1.PontosEnvido = j.calcularPontosEnvido(jogador1)
		jogador2.PontosEnvido = j.calcularPontosEnvido(jogador2)
		jogador1.PontosFlor = j.calcularPontosFlor(jogador1)
		jogador2.PontosFlor = j.calcularPontosFlor(jogador2)
		j.JogadorComFlor = nil
		if j.temFlor(jogador1) {
			j.JogadorComFlor = append(j.JogadorComFlor, jogador1)
		}
		if j.temFlor(jogador2) {
			j.JogadorComFlor = append(j.JogadorComFlor, jogador2)
		}
		fmt.Println(j.gerarStringEstado())
	}

	action := func(a Acao) { // processa toda ação e printa o estado do jogo
		j.ProcessarAcao(a)
		fmt.Println(j.gerarStringEstado())
	}

	log.Println("--- ENVIDO ACEITO ---")
	darCartas(
		[]Carta{{3, "Espadas", 9}, {6, "Espadas", 2}, {2, "Copas", 8}},
		[]Carta{{4, "Ouros", 0}, {5, "Copas", 1}, {7, "Copas", 3}},
	)
	action(Acao{"ENVIDO", "j1", nil})
	action(Acao{"QUERO", "j2", "QUERO"})

	log.Println("--- REAL ENVIDO RECUSADO ---")
	darCartas(
		[]Carta{{4, "Espadas", 0}, {6, "Espadas", 2}, {2, "Copas", 8}},
		[]Carta{{4, "Ouros", 0}, {5, "Copas", 1}, {7, "Copas", 3}},
	)
	action(Acao{"REAL_ENVIDO", "j2", nil})
	action(Acao{"NAO_QUERO", "j1", "NAO_QUERO"})

	log.Println("--- FALTA ENVIDO ACEITO ---")
	j.Placar[1] = 10
	darCartas(
		[]Carta{{4, "Espadas", 0}, {6, "Espadas", 2}, {7, "Espadas", 3}},
		[]Carta{{4, "Copas", 0}, {5, "Copas", 1}, {7, "Copas", 3}},
	)
	action(Acao{"FALTA_ENVIDO", "j1", nil})
	action(Acao{"QUERO", "j2", "QUERO"})

	log.Println("--- FLOR NORMAL ACEITA ---")
	darCartas(
		[]Carta{{5, "Copas", 1}, {6, "Copas", 2}, {7, "Copas", 3}},
		[]Carta{{1, "Espadas", 13}, {2, "Paus", 8}, {4, "Ouros", 0}},
	)
	action(Acao{"FLOR", "j1", nil})
	action(Acao{"QUERO", "j2", "QUERO"})

	log.Println("--- CONTRA-FLOR RECUSADA ---")
	darCartas(
		[]Carta{{5, "Copas", 1}, {6, "Copas", 2}, {7, "Copas", 3}},
		[]Carta{{5, "Copas", 1}, {6, "Copas", 2}, {7, "Copas", 3}},
	)
	action(Acao{"FLOR", "j1", nil})
	action(Acao{"CONTRA_FLOR", "j2", nil})
	action(Acao{"NAO_QUERO", "j1", "NAO_QUERO"})

	log.Println("--- CONTRA-FLOR AL RESTO ACEITA ---")
	j.Placar[0] = 25
	j.Placar[1] = 28
	darCartas(
		[]Carta{{5, "Espadas", 1}, {6, "Espadas", 2}, {7, "Espadas", 3}},
		[]Carta{{5, "Espadas", 1}, {6, "Espadas", 2}, {7, "Espadas", 3}},
	)
	action(Acao{"FLOR", "j1", nil})
	action(Acao{"CONTRA_FLOR", "j2", nil})
	action(Acao{"CONTRA_FLOR_AL_RESTO", "j1", nil})
	action(Acao{"QUERO", "j2", "QUERO"})

	log.Println("--- FAMÍLIA REAL ACEITA ---")
	action(Acao{"TEM_FAMILIA_REAL", "j1", nil})
	action(Acao{"QUERO", "j2", "QUERO"})

	log.Println("--- FAMÍLIA REAL RECUSADA ---")
	action(Acao{"TEM_FAMILIA_REAL", "j1", nil})
	action(Acao{"NAO_QUERO", "j2", "NAO_QUERO"})

	log.Println("--- TRUCO → RETRUCO → VALE 4 ACEITOS ---")
	darCartas(
		[]Carta{{3, "Espadas", 9}, {6, "Ouros", 2}, {7, "Espadas", 3}},
		[]Carta{{1, "Ouros", 7}, {2, "Copas", 8}, {4, "Copas", 0}},
	)
	action(Acao{"TRUCO", "j1", nil})
	action(Acao{"QUERO", "j2", "QUERO"})
	action(Acao{"RETRUCO", "j2", nil})
	action(Acao{"QUERO", "j1", "QUERO"})
	action(Acao{"VALE_CUATRO", "j1", nil})
	action(Acao{"QUERO", "j2", "QUERO"})

	log.Println("--- IR AO MAZO ---")
	action(Acao{"IR_AO_MAZO", "j2", nil})

	log.Println("--- CANTO COM 29 PONTOS ---")
	j.Placar[0] = 29
	action(Acao{"TRUCO", "j1", nil})

	log.Println("--- EMPATES ---")
	j.Rodada = [3]int{2, 2, 2}
	j.resolverMao()
	fmt.Println(j.gerarStringEstado())
	j.Rodada = [3]int{2, 2, 1}
	j.resolverMao()
	fmt.Println(j.gerarStringEstado())
	j.Rodada = [3]int{2, 0, 1}
	j.resolverMao()
	fmt.Println(j.gerarStringEstado())
	j.Rodada = [3]int{1, 2, 0}
	j.resolverMao()
	fmt.Println(j.gerarStringEstado())
	j.Rodada = [3]int{1, 0, 2}
	j.resolverMao()
	fmt.Println(j.gerarStringEstado())

	log.Println("=== ESTADO FINAL ===")
	fmt.Println(j.gerarStringEstado())
}

func main() {
	jogo := CriarJogo()
	SimularPartidaCompleta(jogo)
}