package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"
	"trugo/models"

	"github.com/gorilla/websocket"
)

func CriarSala(m []byte, conn *websocket.Conn) {
	var payload models.CriarSalaID

	if err := json.Unmarshal(m, &payload); err != nil {
		log.Println(err)
		return
	}

	_, ok := models.Salas[payload.ID]

	if ok {
		resposta := models.Resposta{
			Type: "error",
			Msg:  "Já há uma sala com esse ID",
		}

		data, _ := json.Marshal(resposta)
		conn.WriteMessage(websocket.TextMessage, data)
		return
	}

	sala := models.Sala{}
	sala.PrepararJogo()

	models.Salas[payload.ID] = &sala

	resposta := models.Resposta{
		Type: "ok",
		Msg:  fmt.Sprintf("Sala criada com sucesso, ID: %s", payload.ID),
	}

	data, _ := json.Marshal(resposta)

	conn.WriteMessage(websocket.TextMessage, data)
}

func EntrarSala(m []byte, conn *websocket.Conn) {
	var payload models.EntrarSala

	if err := json.Unmarshal(m, &payload); err != nil {
		log.Println(err)
		return
	}

	sala, ok := models.Salas[payload.IdSala]
	if !ok { // (EXCEPTION) ID DA SALA NÃO ENCONTRADO
		resposta := models.Resposta{
			Type: "Err",
			Msg:  fmt.Sprintf("A sala com o ID %s não foi encontrada", payload.IdSala),
		}

		data, _ := json.Marshal(resposta)
		conn.WriteMessage(websocket.TextMessage, data)
		return
	}

	if len(sala.Jogadores) >= 2 { // (EXCEPTION) SALA LOTADA
		resposta := models.Resposta{
			Type: "Err",
			Msg:  fmt.Sprintf("A sala com o ID %s já está lotada", payload.IdSala),
		}

		data, _ := json.Marshal(resposta)
		conn.WriteMessage(websocket.TextMessage, data)
		return
	}

	jogador := models.NovoJogador(payload.Nome, conn)

	sala.Jogadores = append(sala.Jogadores, jogador)
	resposta := models.EntrouSalaResposta{
		Type:          "ok",
		ID:            payload.IdSala,
		Equipe01Vagas: 1 - len(sala.Jogo.Time01.Jogadores),
		Equipe02Vagas: 1 - len(sala.Jogo.Time02.Jogadores),
	}

	data, _ := json.Marshal(resposta)
	conn.WriteMessage(websocket.TextMessage, data)
}

func EscolherTime(m []byte, conn *websocket.Conn) {
	var payload models.EscolherEquipe

	if err := json.Unmarshal(m, &payload); err != nil {
		log.Println(err)
		return
	}

	sala, ok := models.Salas[payload.ID]

	if !ok {
		// SE NÃO ACHAR A SALA (ADICIONAR COD)
	}

	if jogador := ProcurarJogador(sala.Jogadores, conn); jogador != nil {
		entrouEquipe := sala.Jogo.EscolherEquipe(payload.TimeEscolhido, jogador)

		var resposta models.Resposta

		if entrouEquipe {
			resposta = models.Resposta{
				Type: "ok",
				Msg:  "Você entrou no time com sucesso",
			}
		} else {
			resposta = models.Resposta{
				Type: "error",
				Msg:  "O time selecionado não há vagas disponiveis",
			}
		}

		data, _ := json.Marshal(resposta)
		conn.WriteMessage(websocket.TextMessage, data)

	} // CASO NÃO ACHE O JOGADOR (ADICIONAR COD) else (EXCEPTION)

	if len(sala.Jogo.Time01.Jogadores)+len(sala.Jogo.Time02.Jogadores) == 2 {
		ComecarPartida(sala)
	}
}

func ListarSalas(conn *websocket.Conn) {
	salasDisponiveis := make(map[string]int)

	for chave, sala := range models.Salas {
		salasDisponiveis[chave] = 2 - len(sala.Jogadores)
	}

	var payload models.SalasDisponiveis
	payload.SalasDisponiveis = salasDisponiveis

	data, _ := json.Marshal(payload)

	conn.WriteMessage(websocket.TextMessage, data)
}

func ComecarPartida(sala *models.Sala) {
	// Inicia a partida
	sala.Status = "EM_ANDAMENTO"

	// Cria o baralho e atribuir ao Estado do Jogo
	sala.Jogo.Baralho = CriarBaralho()

	// Adicinar um for caso haja mais de 2 jogadores
	sala.Jogo.JogadorVez = sala.Jogadores[1]
	IniciarRodada(sala)
}

// REFATORAR DEPOIS
func ProcurarJogador(listaJogadores []*models.Jogador, conn *websocket.Conn) *models.Jogador {
	for _, jogador := range listaJogadores {
		if jogador.Conn == conn {
			return jogador
		}
	}

	return nil
}

// <<BLOCO ONGOING>> {
func IniciarRodada(sala *models.Sala) {
	// Embaralha o baralho
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	r.Shuffle(len(sala.Jogo.Baralho), func(i, j int) {
		sala.Jogo.Baralho[i], sala.Jogo.Baralho[j] = sala.Jogo.Baralho[j], sala.Jogo.Baralho[i]
	})

	// Limpa as mãos dos jogadores antes de atribuir as cartas
	for _, jogador := range sala.Jogadores {
		jogador.Mao = []models.Cartas{}
	}

	// Atribui as cartas aos jogadores
	idxBaralho := 0
	for i := 0; i < 3; i++ {
		for _, jogador := range sala.Jogadores {
			jogador.Mao = append(jogador.Mao, sala.Jogo.Baralho[idxBaralho])
			idxBaralho++
		}
	}

	rodada := models.Rodada{
		Flor:        true,
		Envido:      true,
		Truco:       true,
		ContraFlor:  true,
		RealEnvido:  true,
		FaltaEnvido: true,
		Retruco:     true,
		ValeQuatro:  true,
		// AJUSTAR ISSO
		VezJogador: sala.Jogadores[1],
	}

	sala.Jogo.Rodadas = append(sala.Jogo.Rodadas, &rodada)

	AvisarJogadorVez(sala.Jogo.JogadorVez, &rodada)
}

func AvisarJogadorVez(j *models.Jogador, r *models.Rodada) {
	payload := models.StatusRodada{
		Type:              "SUA_VEZ",
		CartasJogadas:     CartasNaMesa(r),
		ApostasDiponiveis: ApostasAtivas(r),
	}

	data, _ := json.Marshal(payload)
	j.Conn.WriteMessage(websocket.TextMessage, data)
}

func ApostasAtivas(r *models.Rodada) map[string]bool {
	return map[string]bool{
		"Flor":        r.Flor,
		"Envido":      r.Envido,
		"Truco":       r.Truco,
		"ContraFlor":  r.ContraFlor,
		"RealEnvido":  r.RealEnvido,
		"FaltaEnvido": r.FaltaEnvido,
		"Retruco":     r.Retruco,
		"ValeQuatro":  r.ValeQuatro,
	}
}

// <<BLOCO ONGOING>> }

func FazerJogada(m []byte, conn *websocket.Conn) {
	var payload models.FazerJogada
	json.Unmarshal(m, &payload)

	salaExiste := VerificarSalaExiste(payload.IDSala, conn)
	if salaExiste == nil {
		return
	}
	if !VerificarVezJogadorRodada(salaExiste, conn) {
		return
	}

	log.Println(payload)
	// LÓGICA DA JOGADA AQUI
	// VERIFICAR SE O JOGADOR CHAMOU ALGUMA APOSTA... # payload.ApostaPedida
	// VERIFICAR SE O JOGADOR JOGOU UMA CARTA QUE ESTÁ NA SUA MÃO... # salaExiste.Jogo.JogadorVez.Mao (fazer um FOR)
	// ADICIONAR A CARTA JOGADA AO # RodadaAtual(salaExiste).CartasJogada (usar append)
	// CHAMAR FUNÇÃO DE VERIFICAR O ESTADO DA RODADA, CASO O JOGADOR FOR O ÚLTIMO A JOGAR (se ganhou a mão ou se perdeu)
	// PASSAR A VEZ PARA O PRÓXIMO JOGADOR (o jogador que ganhou a rodada)
	// NOTIFICAR TODOS OS JOGADORES SOBRE A JOGADA FEITA (função NotificarJogadores)
	// VERIFICA SE ACABOU A RODADA E PASSA PARA O JOGADOR SEGUINTE (next na lista)
}

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

func VerificarJogadorNaSala(sala *models.Sala, conn *websocket.Conn) bool {
	for _, jogador := range sala.Jogadores {
		if jogador.Conn == conn {
			return true
		}
	}

	return false
}

func RodadaAtual(sala *models.Sala) *models.Rodada {
	return sala.Jogo.Rodadas[len(sala.Jogo.Rodadas)-1]
}

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

func responderErro(conn *websocket.Conn, msg string, args ...interface{}) {
	resposta := models.Resposta{
		Type: "err",
		Msg:  fmt.Sprintf(msg, args...),
	}
	data, _ := json.Marshal(resposta)
	conn.WriteMessage(websocket.TextMessage, data)
}

func CartasNaMesa(r *models.Rodada) []models.Jogada {
	lista := []models.Jogada{}

	for _, cartas := range r.CartasJogada {
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
