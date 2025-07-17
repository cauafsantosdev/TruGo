package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"slices"
	"sort"
	"strings"
	"trugo/models"

	"github.com/gorilla/websocket"
)

// Constantes para controle do estado do jogo
const (
	StatusAguardandoAposta = "AGUARDANDO_RESPOSTA_APOSTA"
	EstadoQuero            = "QUERO"

	TipoTruco      = "TRUCO"
	TipoRetruco    = "RETRUCO"
	TipoValeQuatro = "VALE_QUATRO"

	TipoEnvido      = "ENVIDO"
	TipoRealEnvido  = "REAL_ENVIDO"
	TipoFaltaEnvido = "FALTA_ENVIDO"

	TipoFlor              = "FLOR"
	TipoContraFlor        = "CONTRA_FLOR"
	TipoContraFlorAlResto = "CONTRA_FLOR_AL_RESTO"

	Time01 = "TIME_01"
	Time02 = "TIME_02"
)

// Inicia a partida
func ComecarPartida(sala *models.Sala) {
	// Inicia a partida
	sala.Status = "EM_ANDAMENTO"
	sala.Jogo.IdxJogador = 0

	// Cria o baralho e atribuir ao Estado do Jogo
	sala.Jogo.Baralho = CriarBaralho()

	// Adicinar um for caso haja mais de 2 jogadores
	sala.Jogo.JogadorMao = sala.Jogo.Time01.Jogadores[0]
	IniciarRodada(sala)
}

// Começa uma rodada
func IniciarRodada(sala *models.Sala) {
	// Embaralha o baralho
	rand.Shuffle(len(sala.Jogo.Baralho), func(i, j int) {
		sala.Jogo.Baralho[i], sala.Jogo.Baralho[j] = sala.Jogo.Baralho[j], sala.Jogo.Baralho[i]
	})

	if sala.Jogo.Estado == "PARTIDA_FINALIZADA" {
		return
	}

	// Limpa as mãos dos jogadores antes de atribuir as cartas
	for _, jogador := range sala.Jogadores {
		jogador.Mao = []models.Cartas{}
		jogador.TemFlor = false
	}

	// Atribui as cartas aos jogadores
	idxBaralho := 0
	for _, jogador := range sala.Jogadores {
		jogador.Mao = sala.Jogo.Baralho[idxBaralho : idxBaralho+3]
		idxBaralho += 3

		if jogador.Mao[0].Valor == 4 && jogador.Mao[1].Valor == 4 && jogador.Mao[2].Valor == 4 {
			jogador.Mao = sala.Jogo.Baralho[idxBaralho : idxBaralho+3]
			idxBaralho += 3
		}

		if jogador.Mao[0].Naipe == jogador.Mao[1].Naipe && jogador.Mao[1].Naipe == jogador.Mao[2].Naipe {
			jogador.TemFlor = true
		}
	}

	EnviarMaosAosJogadores(sala)
	sala.Jogo.Estado = "EM_ANDAMENTO"

	// Reseta os status da rodada
	rodada := models.Rodada{
		Flor:              true,
		Envido:            true,
		Truco:             true,
		ContraFlor:        false,
		ContraFlorAlResto: false,
		RealEnvido:        true,
		FaltaEnvido:       true,
		Retruco:           false,
		ValeQuatro:        false,
		ValorDaMao:        1,
		CadeiaEnvido:      []string{},
		EstadoFlor:        "",
		CartasJogadas:     []models.CartaJogada{},
		CartasEmDisputa:   []models.CartaJogada{},
		VezJogador:        AlternarVezJogador(sala),
	}

	rodada.TimeDaMao = rodada.VezJogador.Time
	sala.Jogo.Rodada = &rodada

	AvisarJogadorVez(rodada.VezJogador, &rodada, sala)
	NotificarJogadores(sala)
}

// Mostra as cartas para os jogadores
func EnviarMaosAosJogadores(s *models.Sala) {
	for _, jogador := range s.Jogadores {
		payload := models.MaoDaRodada{
			Type: "MAO_RODADA",
		}
		payload.Equipes = make(map[string]string)

		for _, c := range jogador.Mao {
			payload.Mao = append(payload.Mao, models.CartaResposta(c))
		}

		payload.Equipes[Time01] = s.Jogo.Time01.Jogadores[0].Nome
		payload.Equipes[Time02] = s.Jogo.Time02.Jogadores[0].Nome

		if data, err := json.Marshal(payload); err == nil {
			jogador.Conn.WriteMessage(websocket.TextMessage, data)
		}
	}
}

// Passa o turno para o próximo jogador
func AlternarVezJogador(s *models.Sala) *models.Jogador {
	if s.Jogo.IdxJogador == 0 {
		s.Jogo.IdxJogador ^= 1 // XOR 0 ↔ 1
		return s.Jogo.Time01.Jogadores[0]
	}
	s.Jogo.IdxJogador ^= 1 // XOR 0 ↔ 1
	return s.Jogo.Time02.Jogadores[0]
}

// Avisa o jogador que é sua vez
func AvisarJogadorVez(j *models.Jogador, r *models.Rodada, s *models.Sala) {
	payload := models.StatusRodada{
		Type:              "SUA_VEZ",
		CartasJogadas:     CartasNaMesa(r),
		ApostasDiponiveis: ApostasAtivas(r),
		Placar:            MostrarPlacar(s),
	}

	data, _ := json.Marshal(payload)
	j.Conn.WriteMessage(websocket.TextMessage, data)
}

// Mostra o placar do jogo
func MostrarPlacar(s *models.Sala) map[string]int {
	placar := make(map[string]int)

	placar[Time01] = s.Jogo.Time01.Pontos
	placar[Time02] = s.Jogo.Time02.Pontos

	return placar
}

// Mostra as apostas disponíveis na rodada
func ApostasAtivas(r *models.Rodada) map[string]bool {
	// Calcula rodadaAtual: maior número de cartas jogadas por qualquer jogador + 1
	jogadores := make(map[string]int)
	for _, jogada := range r.CartasJogadas {
		if jogada.Jogador != nil {
			jogadores[jogada.Jogador.Nome]++
		}
	}
	maxCartas := 0
	for _, v := range jogadores {
		if v > maxCartas {
			maxCartas = v
		}
	}
	rodadaAtual := maxCartas + 1 // 1 = primeira rodada, 2 = segunda, etc

	apostas := map[string]bool{
		"Flor":              r.Flor,
		"Truco":             r.Truco,
		"ContraFlor":        r.ContraFlor,
		"ContraFlorAlResto": r.ContraFlorAlResto,
		"Retruco":           r.Retruco,
		"ValeQuatro":        r.ValeQuatro,
	}
	// Só permite Envido, RealEnvido e FaltaEnvido na primeira rodada
	if rodadaAtual == 1 {
		apostas["Envido"] = r.Envido
		apostas["RealEnvido"] = r.RealEnvido
		apostas["FaltaEnvido"] = r.FaltaEnvido
	} else {
		apostas["Envido"] = false
		apostas["RealEnvido"] = false
		apostas["FaltaEnvido"] = false
	}
	return apostas
}

// Função para jogar uma carta
func FazerJogada(m []byte, conn *websocket.Conn) {
	var payload models.FazerJogada
	json.Unmarshal(m, &payload)

	salaExiste := VerificarSalaExiste(payload.IDSala, conn)

	if salaExiste.Jogo.Estado == "PARTIDA_FINALIZADA" {
		return
	}

	rodadaAtual := RodadaAtual(salaExiste)

	if salaExiste == nil {
		// Não existe a sala
		return
	}
	if !VerificarVezJogadorRodada(salaExiste, conn) {
		return
	}
	if salaExiste.Jogo.Estado != "EM_ANDAMENTO" {
		// Aposta em andamento
		return
	}

	// Adiciona carta jogada à mesa e coloca em disputa
	cartaJogada, _ := VerificarCartaJogada(rodadaAtual.VezJogador, payload)
	rodadaAtual.CartasEmDisputa = append(RodadaAtual(salaExiste).CartasEmDisputa, cartaJogada)
	rodadaAtual.CartasJogadas = append(RodadaAtual(salaExiste).CartasJogadas, cartaJogada)

	// Verifica quem ganhou, caso o jogador seja o último da mão
	if rodadaAtual.IdxJogador == 1 {
		jogadorGanhouMao := ResolverRodada(rodadaAtual.CartasEmDisputa)

		// PASSAR A VEZ PARA O PRÓXIMO JOGADOR (o jogador que ganhou a mão)
		if jogadorGanhouMao == nil {
			rodadaAtual.Rodada = append(rodadaAtual.Rodada, 0)
			switch rodadaAtual.VezJogador.Time {
			case "TIME_01":
				rodadaAtual.VezJogador = salaExiste.Jogo.Time02.Jogadores[0]
			case "TIME_02":
				rodadaAtual.VezJogador = salaExiste.Jogo.Time01.Jogadores[0]
			}
		} else if jogadorGanhouMao.Time == "TIME_01" {
			rodadaAtual.VezJogador = jogadorGanhouMao
			rodadaAtual.Rodada = append(rodadaAtual.Rodada, 1)
		} else if jogadorGanhouMao.Time == "TIME_02" {
			rodadaAtual.VezJogador = jogadorGanhouMao
			rodadaAtual.Rodada = append(rodadaAtual.Rodada, 2)
		}

		NotificarJogadores(salaExiste)

		rodadaAtual.CartasEmDisputa = []models.CartaJogada{}

		rodadaAtual.Envido = false
		rodadaAtual.RealEnvido = false
		rodadaAtual.FaltaEnvido = false

		rodadaAtual.Flor = false
		rodadaAtual.ContraFlor = false
		rodadaAtual.ContraFlorAlResto = false

		rodadaAtual.IdxJogador = 0

	} else {
		switch rodadaAtual.VezJogador.Time {
		case "TIME_01":
			rodadaAtual.VezJogador = salaExiste.Jogo.Time02.Jogadores[0]
		case "TIME_02":
			rodadaAtual.VezJogador = salaExiste.Jogo.Time01.Jogadores[0]
		}

		rodadaAtual.IdxJogador = 1
		NotificarJogadores(salaExiste)

	}

	// Verifica se ouve um vencedor da rodada e passa para a próxima
	equipe, fimDaMao := TimeGanhadorMao(rodadaAtual.Rodada, &salaExiste.Jogo.Time01, &salaExiste.Jogo.Time02)
	if fimDaMao {
		pontosGanhos := rodadaAtual.ValorDaMao

		AtribuirPontoTime(equipe, pontosGanhos, salaExiste)
		NotificarMaoFinalizada(salaExiste, equipe, pontosGanhos)
		IniciarRodada(salaExiste)
		return
	}

	RetirarCartaJogador(rodadaAtual.VezJogador, cartaJogada.Carta)
	AvisarJogadorVez(rodadaAtual.VezJogador, rodadaAtual, salaExiste)
}

func NotificarPontosEnvido(s *models.Sala, time string) {
	payload := models.PontosDaMao{
		Type:     "PONTOS_ENVIDO",
		Equipe:   make(map[string]int),
		Vencedor: time,
		Placar:   MostrarPlacar(s),
	}

	for _, jogador := range s.Jogadores {
		switch jogador.Time {
		case Time01:
			payload.Equipe[Time01] = jogador.PontosEnvido
		case Time02:
			payload.Equipe[Time02] = jogador.PontosEnvido
		}
	}

	data, _ := json.Marshal(payload)

	for _, jogador := range s.Jogadores {
		jogador.Conn.WriteMessage(websocket.TextMessage, data)
	}
}

// Remove a carta da mão do jogador
func RetirarCartaJogador(j *models.Jogador, c *models.Cartas) {
	cartas := []models.Cartas{}

	for _, carta := range j.Mao {
		if carta.Valor != c.Valor || carta.Naipe != c.Naipe {
			cartas = append(cartas, carta)
		}
	}

	j.Mao = cartas
}

// Determina o vencedor de uma rodada
func ResolverRodada(cartasJogada []models.CartaJogada) *models.Jogador {
	if cartasJogada[0].Carta.Forca > cartasJogada[1].Carta.Forca {
		return cartasJogada[0].Jogador
	} else if cartasJogada[1].Carta.Forca > cartasJogada[0].Carta.Forca {
		return cartasJogada[1].Jogador
	}

	return nil
}

// Define o time vencedor da mão
func TimeGanhadorMao(m []int, time01, time02 *models.Equipe) (*models.Equipe, bool) {
	time01Pnts := 0
	time02Pnts := 0

	for _, pnt := range m {
		switch pnt {
		case 1:
			time01Pnts++
		case 2:
			time02Pnts++
		}
	}

	if time02Pnts < 2 && time01Pnts < 2 {

		// empates
		if len(m) == 3 && m[0] == 0 && m[1] == 0 && m[2] == 0 { // empate triplo
			return time01, true
			// ganha o mão

		} else if len(m) == 3 && m[0] == 0 && m[1] == 0 { // empatou as duas primeiras
			switch m[2] {
			case 1:
				return time01, true
			case 2:
				return time02, true
			}
			// quem ganhou a terceira leva

		} else if len(m) == 2 && m[0] == 0 { // empatou a primeira
			switch m[1] {
			case 1:
				return time01, true
			case 2:
				return time02, true
			} // quem ganhou a segunda leva

		} else if len(m) == 3 && m[0] != 0 && m[1] != 0 && m[2] == 0 { // empatou a terceira
			switch m[0] {
			case 1:
				return time01, true
			case 2:
				return time02, true
			} // quem ganhou a primeira leva

		} else if len(m) == 2 && m[1] == 0 { // empatou a segunda
			switch m[0] {
			case 1:
				return time01, true
			case 2:
				return time02, true
			} // quem ganhou a primeira leva

		}

		// se ninguem ganhou ainda
		return nil, false
	}

	if time02Pnts > time01Pnts {
		return time02, true
	}
	return time01, true
}

// Atribui os tentos ao time vencedor, caso chegue a 30 o time é declarado vencedor
func AtribuirPontoTime(e *models.Equipe, pnts int, s *models.Sala) {
	e.Pontos += pnts

	if pnts == 30 || e.Pontos >= 30 {
		e.Pontos = 30
		AcabouPartida(e, s)
	}
}

// Notifica o websocket que a mão foi encerrada
func NotificarMaoFinalizada(sala *models.Sala, timeVencedor *models.Equipe, pontosGanhos int) {
	payload := models.MaoFinalizada{
		Type:          "MAO_FINALIZADA",
		TimeVencedor:  timeVencedor.Jogadores[0].Time,
		PontosGanhos:  pontosGanhos,
		Placar:        MostrarPlacar(sala),
		CartasJogadas: CartasNaMesa(sala.Jogo.Rodada), // Adiciona as cartas jogadas na rodada
	}

	data, _ := json.Marshal(payload)

	// Envia a mensagem para todos os jogadores na sala
	for _, jogador := range sala.Jogadores {
		jogador.Conn.WriteMessage(websocket.TextMessage, data)
	}
}

// Encerra a partida
func AcabouPartida(e *models.Equipe, s *models.Sala) {
	s.Jogo.Estado = "PARTIDA_FINALIZADA"

	payload := models.PartidaFinalizada{
		Type:   "PARTIDA_FINALIZADA",
		Placar: MostrarPlacar(s),
	}

	for _, j := range s.Jogadores {
		if j == e.Jogadores[0] {
			payload.Mensagem = "VOCE_GANHOU"
		} else {
			payload.Mensagem = "VOCE_PERDEU"
		}

		data, _ := json.Marshal(payload)
		j.Conn.WriteMessage(websocket.TextMessage, data)
	}
}

// Verifica se a carta jogada é válida
func VerificarCartaJogada(vezJogador *models.Jogador, payload models.FazerJogada) (models.CartaJogada, bool) {
	cartaJogada := models.Cartas{
		Valor: payload.CartaJogada.Valor,
		Naipe: payload.CartaJogada.Naipe,
		Forca: payload.CartaJogada.Forca,
	}

	if slices.Contains(vezJogador.Mao, cartaJogada) {
		return models.CartaJogada{
			Jogador: vezJogador,
			Carta:   &cartaJogada,
		}, false
	}

	return models.CartaJogada{}, true
}

// Notifica os jogadores do status da partida
func NotificarJogadores(sala *models.Sala) {
	rodadaAtual := RodadaAtual(sala)

	for _, jogador := range sala.Jogadores {
		if jogador.Conn != rodadaAtual.VezJogador.Conn {
			payload := models.StatusRodada{
				Type:              "STATUS_PARTIDA",
				CartasJogadas:     CartasNaMesa(rodadaAtual),
				ApostasDiponiveis: ApostasAtivas(rodadaAtual),
				Placar:            MostrarPlacar(sala),
			}

			data, _ := json.Marshal(payload)
			jogador.Conn.WriteMessage(websocket.TextMessage, data)
		}
	}
}

// Verifica se é a vez do jogador
func VerificarVezJogadorRodada(sala *models.Sala, conn *websocket.Conn) bool {
	if !VerificarJogadorNaSala(sala, conn) {
		responderErro(conn, "O jogador não está na partida")
		return false
	}
	if RodadaAtual(sala).VezJogador.Conn != conn {
		responderErro(conn, "Não é a vez do jogador")
		return false
	}
	return true
}

// Verifica se um jogador está na sala
func VerificarJogadorNaSala(sala *models.Sala, conn *websocket.Conn) bool {
	for _, jogador := range sala.Jogadores {
		if jogador.Conn == conn {
			return true
		}
	}

	return false
}

// Retorna a rodada atual
func RodadaAtual(sala *models.Sala) *models.Rodada {
	return sala.Jogo.Rodada
}

// Verifica se a sala existe
func VerificarSalaExiste(idSala string, conn *websocket.Conn) *models.Sala {
	sala, ok := models.Salas[idSala]

	if !ok {
		responderErro(conn, "A sala com o ID %s não foi encontrada.", idSala)
		return nil
	}

	if sala.Status != "EM_ANDAMENTO" {
		responderErro(conn, "A sala com o ID %s não está em andamento.", idSala)
		return nil
	}

	return sala
}

// Responde erro
func responderErro(conn *websocket.Conn, msg string, args ...interface{}) {
	resposta := models.Resposta{
		Type: "error",
		Msg:  fmt.Sprintf(msg, args...),
	}
	data, _ := json.Marshal(resposta)
	conn.WriteMessage(websocket.TextMessage, data)
}

// Retorna as cartas jogadas na mesa
func CartasNaMesa(r *models.Rodada) []models.Jogada {
	lista := []models.Jogada{}

	for _, cartas := range r.CartasJogadas {
		carta := models.CartaResposta{
			Naipe: cartas.Carta.Naipe,
			Valor: cartas.Carta.Valor,
			Forca: cartas.Carta.Forca,
		}

		cartaJogada := models.Jogada{
			IDEquipe:    cartas.Jogador.Time,
			JogadorNome: cartas.Jogador.Nome,
			CartaJogada: carta,
		}

		lista = append(lista, cartaJogada)
	}

	return lista
}

// Cria o baralho do jogo
func CriarBaralho() []models.Cartas {
	naipes := []string{"Copas", "Espadas", "Paus", "Ouros"}
	valores := []int{1, 2, 3, 4, 5, 6, 7, 10, 11, 12}
	baralho := make([]models.Cartas, 0, 40)

	for _, naipe := range naipes {
		for _, valor := range valores {
			carta := models.Cartas{Valor: valor, Naipe: naipe}

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

// Canta Truco
func CantarTruco(m []byte, conn *websocket.Conn) {
	var payload models.IDSala

	json.Unmarshal(m, &payload)

	sala := VerificarSalaExiste(payload.IDSala, conn)

	if sala == nil {
		return
	}

	rodadaAtual := RodadaAtual(sala)
	if !rodadaAtual.Truco {
		responderErro(conn, "Não é possível pedir Truco")
	}

	jogadorDoTruco := BuscarJogador(sala, conn)

	time := jogadorDoTruco.Time

	if time == Time01 && sala.Jogo.Time01.Pontos == 29 {
		sala.Jogo.Time01.Pontos = 15
	} else if time == Time02 && sala.Jogo.Time02.Pontos == 29 {
		sala.Jogo.Time02.Pontos = 15
	}

	payload.Type = strings.ReplaceAll(payload.Type, "CHAMAR_", "")

	switch payload.Type {
	case TipoTruco:
		if !rodadaAtual.Truco {
			responderErro(conn, "IMPOSSÍVEL DE PEDIR TRUCO")
			return
		}
	case TipoRetruco:
		if !rodadaAtual.Retruco || jogadorDoTruco.Time != rodadaAtual.ApostaAtual.ParaTime {
			responderErro(conn, "IMPOSSÍVEL DE PEDIR RETRUCO")
			return
		}
	case TipoValeQuatro:
		if !rodadaAtual.ValeQuatro || jogadorDoTruco.Time != rodadaAtual.ApostaAtual.ParaTime {
			responderErro(conn, "IMPOSSÍVEL DE PEDIR VALE QUATRO")
			return
		}
	}

	switch time {
	case Time01:
		rodadaAtual.ApostaAtual = models.Aposta{
			Tipo:     payload.Type,
			Estado:   "AGUARDANDO_RESPOSTA",
			ParaTime: Time02,
		}
		EnviarAposta(Time02, sala, payload.Type)
	case Time02:
		rodadaAtual.ApostaAtual = models.Aposta{
			Tipo:     payload.Type,
			Estado:   "AGUARDANDO_RESPOSTA",
			ParaTime: Time01,
		}
		EnviarAposta(Time01, sala, payload.Type)
	}

	sala.Jogo.Estado = "AGUARDANDO_RESPOSTA_APOSTA"
}

// Canta Envido
func CantarEnvido(m []byte, conn *websocket.Conn) {
	var payload models.IDSala

	json.Unmarshal(m, &payload)

	sala := VerificarSalaExiste(payload.IDSala, conn)

	if sala == nil {
		return
	}

	rodadaAtual := RodadaAtual(sala)
	if !rodadaAtual.Envido {
		responderErro(conn, "Não é possível pedir Envido")
	}

	jogadorDoEnvido := BuscarJogador(sala, conn)
	time := jogadorDoEnvido.Time

	if time == Time01 && sala.Jogo.Time01.Pontos == 29 {
		sala.Jogo.Time01.Pontos = 15
	} else if time == Time02 && sala.Jogo.Time02.Pontos == 29 {
		sala.Jogo.Time02.Pontos = 15
	}

	// Tira o CHAMAR_ do tipo da aposta
	payload.Type = strings.ReplaceAll(payload.Type, "CHAMAR_", "")

	switch time {
	case Time01:
		rodadaAtual.ApostaAtual = models.Aposta{
			Tipo:     payload.Type,
			Estado:   "AGUARDANDO_RESPOSTA",
			ParaTime: Time02,
		}
		EnviarAposta(Time02, sala, payload.Type)
	case Time02:
		rodadaAtual.ApostaAtual = models.Aposta{
			Tipo:     payload.Type,
			Estado:   "AGUARDANDO_RESPOSTA",
			ParaTime: Time01,
		}
		EnviarAposta(Time01, sala, payload.Type)
	}

	rodadaAtual.CadeiaEnvido = append(rodadaAtual.CadeiaEnvido, payload.Type)
	sala.Jogo.Estado = "AGUARDANDO_RESPOSTA_APOSTA"
}

// Canta Flor
func CantarFlor(m []byte, conn *websocket.Conn) {
	var payload models.IDSala

	json.Unmarshal(m, &payload)

	sala := VerificarSalaExiste(payload.IDSala, conn)

	if sala == nil {
		return
	}

	rodadaAtual := RodadaAtual(sala)
	if !rodadaAtual.Flor {
		responderErro(conn, "Não é possível pedir Flor")
	}

	jogadorDaFlor := BuscarJogador(sala, conn)
	var outroJogador *models.Jogador

	for _, jogador := range sala.Jogadores {
		if jogador != jogadorDaFlor {
			outroJogador = jogador
		}
	}

	if !outroJogador.TemFlor {
		time := jogadorDaFlor.Time

		if time == Time01 {
			sala.Jogo.Time01.Pontos += 3
			AvisarFlorBoa(conn, time, sala, false)
		} else {
			sala.Jogo.Time02.Pontos += 3
			AvisarFlorBoa(conn, time, sala, false)
		}

	} else {
		payload.Type = strings.ReplaceAll(payload.Type, "CHAMAR_", "")

		switch jogadorDaFlor.Time {
		case Time01:
			rodadaAtual.ApostaAtual = models.Aposta{
				Tipo:     payload.Type,
				Estado:   "AGUARDANDO_RESPOSTA",
				ParaTime: Time02,
			}
			AvisarFlorAdversario(Time02, sala, true)
		case Time02:
			rodadaAtual.ApostaAtual = models.Aposta{
				Tipo:     payload.Type,
				Estado:   "AGUARDANDO_RESPOSTA",
				ParaTime: Time01,
			}
			AvisarFlorAdversario(Time01, sala, true)
		}

		rodadaAtual.EstadoFlor = payload.Type
		sala.Jogo.Estado = "AGUARDANDO_RESPOSTA_APOSTA"

		rodadaAtual.ContraFlor = true
		rodadaAtual.ContraFlorAlResto = true
	}
}

func AvisarFlorBoa(conn *websocket.Conn, time string, s *models.Sala, c bool) {
	payload := models.RespostaFlor{
		Type: "BOA",
	}

	data, _ := json.Marshal(payload)

	conn.WriteMessage(websocket.TextMessage, data)

	AvisarFlorAdversario(time, s, c)
}

func AvisarFlorAdversario(time string, s *models.Sala, c bool) {
	payload := models.RespostaFlor{
		Type:             "FLOR_CANTADA",
		RespostaParaFlor: c,
	}

	data, _ := json.Marshal(payload)

	switch time {
	case Time01:
		s.Jogo.Time02.Jogadores[0].Conn.WriteMessage(websocket.TextMessage, data)
	case Time02:
		s.Jogo.Time01.Jogadores[0].Conn.WriteMessage(websocket.TextMessage, data)
	}
}

// resolve a disputa de Truco
func AvaliarTruco(sala *models.Sala, r *models.Rodada, time string, quero bool, tipoAposta string) {
	resposta := models.RespostaAposta{
		Type:       "RESPOSTA_APOSTA",
		TipoAposta: TipoTruco,
		Quero:      quero,
	}

	if quero {
		r.ApostaAtual.Estado = EstadoQuero
		switch tipoAposta {
		case TipoTruco:
			r.ValorDaMao = 2

			r.Truco = false
			r.Retruco = true

		case TipoRetruco:
			r.ValorDaMao = 3

			r.Retruco = false
			r.ValeQuatro = true

		case TipoValeQuatro:
			r.ValorDaMao = 4
			r.ValeQuatro = false
		}

	} else {
		// Caso o Truco seja recusado, atribui o valor da mão
		r.ApostaAtual.Estado = "RECUSADO"
		switch time {
		case Time01:
			AtribuirPontoTime(&sala.Jogo.Time02, r.ValorDaMao, sala)
		case Time02:
			AtribuirPontoTime(&sala.Jogo.Time01, r.ValorDaMao, sala)
		}

		r.Truco = false
		r.Retruco = false
		r.ValeQuatro = false

		var timeVencedor *models.Equipe
		if time == Time01 {
			timeVencedor = &sala.Jogo.Time02
		} else {
			timeVencedor = &sala.Jogo.Time01
		}

		NotificarMaoFinalizada(sala, timeVencedor, r.ValorDaMao)
		NotificarRespostaAposta(sala, resposta, time)
		IniciarRodada(sala)
		return
	}

	NotificarRespostaAposta(sala, resposta, time)
}

// resolve a disputa de Envido
func AvaliarEnvido(sala *models.Sala, r *models.Rodada, time string, quero bool, tipoAposta string) {
	resposta := models.RespostaAposta{
		Type:       "RESPOSTA_APOSTA",
		TipoAposta: tipoAposta,
		Quero:      quero,
	}

	timeAposta := time
	falta := false

	var pontosEnvido int

	if quero {
		for _, canto := range r.CadeiaEnvido {
			switch canto {
			case "ENVIDO":
				pontosEnvido += 2
			case "REAL_ENVIDO":
				pontosEnvido += 3
			case "FALTA_ENVIDO":
				falta = true
			}
		}

		time01 := ContarPontosEnvido(sala.Jogo.Time01.Jogadores[0])
		time02 := ContarPontosEnvido(sala.Jogo.Time02.Jogadores[0])

		if time01 > time02 {
			time = "TIME_01"
		} else if time02 > time01 {
			time = "TIME_02"
		} else {
			time = r.TimeDaMao
		}

	} else {
		for i, canto := range r.CadeiaEnvido {
			switch canto {
			case "ENVIDO":
				pontosEnvido += 1
			case "REAL_ENVIDO":
				pontosEnvido += 1
			case "FALTA_ENVIDO":
				switch pontosEnvido {
				case 0:
					pontosEnvido += 1
				case 1:
					switch r.CadeiaEnvido[i-1] {
					case "ENVIDO":
						pontosEnvido += 1
					case "REAL_ENVIDO":
						pontosEnvido += 2
					}
				case 2:
					pontosEnvido += 3
				}
			}
		}
		switch time {
		case Time01:
			time = Time02
		case Time02:
			time = Time01
		}
	}

	switch time {
	case Time01:
		if falta {
			AtribuirPontoTime(&sala.Jogo.Time01, 30-sala.Jogo.Time02.Pontos, sala)
			return
		} else {
			AtribuirPontoTime(&sala.Jogo.Time01, pontosEnvido, sala)
		}
	case Time02:
		if falta {
			AtribuirPontoTime(&sala.Jogo.Time02, 30-sala.Jogo.Time01.Pontos, sala)
			return
		} else {
			AtribuirPontoTime(&sala.Jogo.Time02, pontosEnvido, sala)
		}
	}

	NotificarRespostaAposta(sala, resposta, timeAposta)

	if quero {
		NotificarPontosEnvido(sala, time)
	}

	r.Envido = false
	r.RealEnvido = false
	r.FaltaEnvido = false

	sala.Jogo.Estado = "EM_ANDAMENTO"
}

// determina vencedor da disputa de Flor
func AvaliarFlor(sala *models.Sala, r *models.Rodada, time string, quero bool, tipoAposta string) {
	resposta := models.RespostaAposta{
		Type:       "RESPOSTA_APOSTA",
		TipoAposta: TipoFlor,
		Quero:      quero,
	}

	pontosFlor := 3
	resto := false

	if quero && resposta.TipoAposta == TipoFlor && r.EstadoFlor == "FLOR" { // FLOR e FLOR
		sala.Jogo.Time02.Pontos += 3
		sala.Jogo.Time01.Pontos += 3

	} else if quero {
		switch r.EstadoFlor {
		case "CONTRA_FLOR":
			pontosFlor = 6
		case "CONTRA_FLOR_AL_RESTO":
			resto = true
		}

		time01 := ContarPontosFlor(sala.Jogo.Time01.Jogadores[0].Mao)
		time02 := ContarPontosFlor(sala.Jogo.Time02.Jogadores[0].Mao)

		if time01 > time02 {
			time = Time01
		} else if time02 > time01 {
			time = Time02
		} else {
			time = r.TimeDaMao
		}
	} else {
		switch r.EstadoFlor {
		case TipoContraFlor:
			pontosFlor = 3
		case TipoContraFlorAlResto:
			pontosFlor = 6
		}

		if time == Time01 {
			time = Time02
		} else {
			time = Time01
		}
	}

	switch time {
	case Time01:
		if resto {
			AtribuirPontoTime(&sala.Jogo.Time01, 30-sala.Jogo.Time02.Pontos, sala)
		} else {
			AtribuirPontoTime(&sala.Jogo.Time01, pontosFlor, sala)
		}
	case Time02:
		if resto {
			AtribuirPontoTime(&sala.Jogo.Time02, 30-sala.Jogo.Time01.Pontos, sala)
		} else {
			AtribuirPontoTime(&sala.Jogo.Time02, pontosFlor, sala)
		}
	}

	if quero {
		NotificarPontosEnvido(sala, time)
	}

	r.Flor = false
	r.ContraFlor = false
	r.ContraFlorAlResto = false

	NotificarRespostaAposta(sala, resposta, time)
}

// Manda o canto para o adversário
func EnviarAposta(time string, sala *models.Sala, tipoAposta string) {
	aposta := models.EnviarAposta{
		Type:         "APOSTA",
		TipoDeAposta: tipoAposta,
	}
	var data []byte
	data, _ = json.Marshal(aposta)

	// Envia a aposta para o time adversário
	for _, jogador := range sala.Jogadores {
		if jogador.Time == time {
			jogador.Conn.WriteMessage(websocket.TextMessage, data)
		}
	}
}

// Responde as apostas
func ResponderAposta(m []byte, conn *websocket.Conn) {
	var payload models.ResponderAposta

	json.Unmarshal(m, &payload)

	sala := VerificarSalaExiste(payload.IDSala, conn)

	if sala == nil {
		return
	}

	j := BuscarJogador(sala, conn)
	r := RodadaAtual(sala)

	if sala.Jogo.Estado != StatusAguardandoAposta ||
		r.ApostaAtual.ParaTime != j.Time ||
		r.ApostaAtual.Tipo != payload.TipoAposta {
		responderErro(conn, "Foi impossível de aceitar a aposta.")
		return
	}
	// só entra aqui se for tudo válido
	switch payload.TipoAposta {
	case TipoTruco, TipoRetruco, TipoValeQuatro:
		AvaliarTruco(sala, r, r.ApostaAtual.ParaTime, payload.Aceitar, payload.TipoAposta)
	case TipoEnvido, TipoRealEnvido, TipoFaltaEnvido:
		AvaliarEnvido(sala, r, r.ApostaAtual.ParaTime, payload.Aceitar, payload.TipoAposta)
	case TipoFlor, TipoContraFlor, TipoContraFlorAlResto:
		AvaliarFlor(sala, r, r.ApostaAtual.ParaTime, payload.Aceitar, payload.TipoAposta)
	}
}

// Notifica se o jogador quis, não quis ou aumentou
func NotificarRespostaAposta(sala *models.Sala, resposta models.RespostaAposta, time string) {
	data, _ := json.Marshal(resposta)

	var adversarios []*models.Jogador
	if time == Time01 {
		adversarios = sala.Jogo.Time02.Jogadores
	} else {
		adversarios = sala.Jogo.Time01.Jogadores
	}

	// Notificar Resposta da Aposta
	for _, jogador := range adversarios {
		jogador.Conn.WriteMessage(websocket.TextMessage, data)

	}

	sala.Jogo.Estado = "EM_ANDAMENTO"
}

// conta os pontos do Envido de um jogador
func ContarPontosEnvido(jogador *models.Jogador) int {
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
		maiorCarta = maiorPontuacao
	}

	jogador.PontosEnvido = maiorCarta

	return maiorCarta
}

// Conta os pontos da FLor
func ContarPontosFlor(m []models.Cartas) int {
	valorFlor := 0

	for _, carta := range m {
		switch carta.Valor {
		case 10, 11, 12:
			// não somam pontos
		default:
			valorFlor += carta.Valor
		}
	}

	valorFlor += 20 // bônus da Flor
	return valorFlor
}

// Procura jogador na sala
func BuscarJogador(sala *models.Sala, conn *websocket.Conn) *models.Jogador {
	for _, jogador := range sala.Jogadores {
		if jogador.Conn == conn {
			return jogador
		}
	}

	return nil
}

// IrAoMazo é quando um jogador desiste da rodada
func IrAoMazo(m []byte, conn *websocket.Conn) {
	var payload models.IDSala
	if err := json.Unmarshal(m, &payload); err != nil {
		log.Printf("Erro ao decodificar payload para IrAoMazo: %v", err)
		return
	}

	sala := VerificarSalaExiste(payload.IDSala, conn)
	if sala == nil {
		return
	}

	// Verifica se a partida está em um estado que permite desistir
	if sala.Jogo.Estado != "EM_ANDAMENTO" {
		responderErro(conn, "Não é possível ir ao maço durante uma aposta.")
		return
	}

	// Identifica o jogador que desistiu e a rodada atual
	jogadorQueDesistiu := BuscarJogador(sala, conn)
	rodadaAtual := RodadaAtual(sala)

	// Determina qual time ganhou os pontos
	var timeVencedor *models.Equipe
	if jogadorQueDesistiu.Time == Time01 {
		timeVencedor = &sala.Jogo.Time02
	} else {
		timeVencedor = &sala.Jogo.Time01
	}

	// Atribui os pontos da mão atual para o time vencedor
	pontosGanhos := rodadaAtual.ValorDaMao
	AtribuirPontoTime(timeVencedor, pontosGanhos, sala)
	NotificarMaoFinalizada(sala, timeVencedor, pontosGanhos)

	// Notifica os jogadores sobre a desistência
	resposta := models.Resposta{
		Type: "JOGADOR_FOI_AO_MAZO",
		Msg:  fmt.Sprintf("O jogador %s foi ao mazo. A equipe %s ganha %d ponto(s).", jogadorQueDesistiu.Nome, timeVencedor.Jogadores[0].Time, pontosGanhos),
	}
	data, _ := json.Marshal(resposta)
	for _, jogador := range sala.Jogadores {
		jogador.Conn.WriteMessage(websocket.TextMessage, data)
	}

	// Se a partida não acabou, inicia a próxima rodada.
	if sala.Jogo.Estado != "PARTIDA_FINALIZADA" {
		IniciarRodada(sala)
	}
}

// Toca um áudio para todos os jogadores da sala
func TocarAudio(m []byte, conn *websocket.Conn) {
	var payload models.ComandoAudio
	if err := json.Unmarshal(m, &payload); err != nil {
		log.Printf("Erro ao decodificar payload para ComandoAudio: %v", err)
		return
	}

	sala := VerificarSalaExiste(payload.IDSala, conn)
	if sala == nil {
		return
	}

	mensagemParaJogadores := models.TocarAudio{
		Type:      "TOCAR_AUDIO",
		NomeAudio: payload.NomeAudio,
	}

	data, _ := json.Marshal(mensagemParaJogadores)
	for _, jogador := range sala.Jogadores {
		jogador.Conn.WriteMessage(websocket.TextMessage, data)
	}
}
