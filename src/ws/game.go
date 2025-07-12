package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"trugo/models"

	"github.com/gorilla/websocket"
)

const (
	StatusAguardandoAposta = "AGUARDANDO_RESPOSTA_APOSTA"
	EstadoAceito           = "ACEITO"

	TipoTruco      = "TRUCO"
	TipoRetruco    = "RETRUCO"
	TipoValeQuatro = "VALE_QUATRO"

	TipoEnvido      = "ENVIDO"
	TipoRealEnvido  = "REAL_ENVIDO"
	TipoFaltaEnvido = "FALTA_ENVIDO"

	Time01 = "TIME_01"
	Time02 = "TIME_02"
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
			Type: "error",
			Msg:  fmt.Sprintf("A sala com o ID %s não foi encontrada", payload.IdSala),
		}

		data, _ := json.Marshal(resposta)
		conn.WriteMessage(websocket.TextMessage, data)
		return
	}

	if len(sala.Jogadores) >= 2 { // (EXCEPTION) SALA LOTADA
		resposta := models.Resposta{
			Type: "error",
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

	var resposta models.Resposta

	if !ok {

		resposta = models.Resposta{
			Type: "error",
			Msg:  "Sala com esse ID não foi encontrada",
		}

		data, _ := json.Marshal(resposta)
		conn.WriteMessage(websocket.TextMessage, data)
	}

	if jogador := ProcurarJogador(sala.Jogadores, conn); jogador != nil {
		entrouEquipe := sala.Jogo.EscolherEquipe(payload.TimeEscolhido, jogador)

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
	sala.Jogo.IdxJogador = 0

	// Cria o baralho e atribuir ao Estado do Jogo
	sala.Jogo.Baralho = CriarBaralho()

	// Adicinar um for caso haja mais de 2 jogadores
	sala.Jogo.JogadorVez = sala.Jogo.Time01.Jogadores[0]
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
	// src := rand.NewSource(time.Now().UnixNano())
	// r := rand.New(src)

	// r.Shuffle(len(sala.Jogo.Baralho), func(i, j int) {
	// 	sala.Jogo.Baralho[i], sala.Jogo.Baralho[j] = sala.Jogo.Baralho[j], sala.Jogo.Baralho[i]
	// })

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

	EnviarMaosAosJogadores(sala)
	sala.Jogo.Estado = "EM_ANDAMENTO"

	rodada := models.Rodada{
		Flor:        true,
		Envido:      true,
		Truco:       true,
		ContraFlor:  true,
		RealEnvido:  true,
		FaltaEnvido: true,
		Retruco:     false,
		ValeQuatro:  false,
		ValorDaMao:  1,
		// AJUSTAR ISSO
		VezJogador: AlternarVezJogador(sala),
	}

	sala.Jogo.Rodadas = append(sala.Jogo.Rodadas, &rodada)

	AvisarJogadorVez(rodada.VezJogador, &rodada, sala)
	NotificarJogadores(sala)
}

func EnviarMaosAosJogadores(s *models.Sala) {
	for _, jogador := range s.Jogadores {
		payload := models.MaoDaRodada{
			Type: "MAO_RODADA",
		}

		for _, c := range jogador.Mao {
			payload.Mao = append(payload.Mao, models.CartaResposta(c))
		}

		if data, err := json.Marshal(payload); err == nil {
			jogador.Conn.WriteMessage(websocket.TextMessage, data)
		}
	}
}

func AlternarVezJogador(s *models.Sala) *models.Jogador {
	if s.Jogo.IdxJogador == 0 {
		s.Jogo.IdxJogador ^= 1 // XOR 0 ↔ 1
		return s.Jogo.Time01.Jogadores[0]
	}
	s.Jogo.IdxJogador ^= 1 // XOR 0 ↔ 1
	return s.Jogo.Time02.Jogadores[0]
}

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

func MostrarPlacar(s *models.Sala) map[string]int {
	placar := make(map[string]int)

	placar[Time01] = s.Jogo.Time01.Pontos
	placar[Time02] = s.Jogo.Time02.Pontos

	return placar
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

	rodadaAtual := RodadaAtual(salaExiste)

	// LÓGICA DA JOGADA AQUI
	// VERIFICAR SE O JOGADOR CHAMOU ALGUMA APOSTA... # payload.ApostaPedida

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

	// VERIFICAR SE O JOGADOR JOGOU UMA CARTA QUE ESTÁ NA SUA MÃO... # salaExiste.Jogo.JogadorVez.Mao (fazer um FOR)
	cartaJogada, err := VerificarCartaJogada(rodadaAtual.VezJogador, payload)
	if err {
		return
	}

	// ADICIONAR A CARTA JOGADA AO # RodadaAtual(salaExiste).CartasJogada (usar append)
	rodadaAtual.CartasJogada = append(RodadaAtual(salaExiste).CartasJogada, cartaJogada)

	// CHAMAR FUNÇÃO DE VERIFICAR O ESTADO DA RODADA, CASO O JOGADOR FOR O ÚLTIMO A JOGAR (se ganhou a mão ou se perdeu)
	if rodadaAtual.IdxJogador == 1 {
		jogadorGanhouMao := VerificarMao(rodadaAtual.CartasJogada)

		// PASSAR A VEZ PARA O PRÓXIMO JOGADOR (o jogador que ganhou a mão)
		if jogadorGanhouMao == nil {
			// rodadaAtual.VezJogador = salaExiste.Jogadores[0] Adicionar se escalar para 4 jogadores
			rodadaAtual.Rodada = append(rodadaAtual.Rodada, 0)
		} else if jogadorGanhouMao.Time == "TIME_01" {
			rodadaAtual.VezJogador = jogadorGanhouMao
			rodadaAtual.Rodada = append(rodadaAtual.Rodada, 1)
		} else if jogadorGanhouMao.Time == "TIME_02" {
			rodadaAtual.VezJogador = jogadorGanhouMao
			rodadaAtual.Rodada = append(rodadaAtual.Rodada, 2)
		}

		NotificarJogadores(salaExiste)
		rodadaAtual.CartasJogada = []models.CartaJogada{}
		rodadaAtual.IdxJogador = 0

	} else {
		switch rodadaAtual.VezJogador.Time {
		case "TIME_01":
			rodadaAtual.VezJogador = salaExiste.Jogo.Time02.Jogadores[0]
		case "TIME_02":
			rodadaAtual.VezJogador = salaExiste.Jogo.Time01.Jogadores[0]
		}
		// Controla o número de jogadas em uma mão (2 no max)

		rodadaAtual.IdxJogador = 1
		NotificarJogadores(salaExiste)

	}

	// NOTIFICAR TODOS OS JOGADORES SOBRE A JOGADA FEITA (função NotificarJogadores)

	// VERIFICA SE ACABOU A RODADA E PASSA PARA O JOGADOR SEGUINTE (next na lista)
	equipe, fimDaMao := TimeGanhadorMao(rodadaAtual.Rodada, &salaExiste.Jogo.Time01, &salaExiste.Jogo.Time02)
	if fimDaMao {
		log.Println("UAI ZE")
		AtribuirPontoTime(equipe, rodadaAtual.ValorDaMao)
		IniciarRodada(salaExiste)
		return
	}

	AvisarJogadorVez(rodadaAtual.VezJogador, rodadaAtual, salaExiste)
}

func TimeGanhadorMao(m []int, time01, time02 *models.Equipe) (*models.Equipe, bool) {
	time01Pnts := 0
	time02Pnts := 0

	for _, pnt := range m {
		if pnt == 1 {
			time01Pnts++
		} else if pnt == 2 {
			time02Pnts++
		} else {
			time01Pnts++
			time02Pnts++
		}
	}

	if time02Pnts < 2 && time01Pnts < 2 {
		return nil, false
	}

	if time02Pnts > time01Pnts {
		return time02, true
	}
	return time01, true
}

func AtribuirPontoTime(e *models.Equipe, pnts int) {
	e.Pontos += pnts
}

func VerificarMao(cartasJogada []models.CartaJogada) *models.Jogador {
	if cartasJogada[0].Carta.Forca > cartasJogada[1].Carta.Forca {
		return cartasJogada[0].Jogador
	} else if cartasJogada[1].Carta.Forca > cartasJogada[0].Carta.Forca {
		return cartasJogada[1].Jogador
	}

	return nil
}

func VerificarCartaJogada(vezJogador *models.Jogador, payload models.FazerJogada) (models.CartaJogada, bool) {
	cartaJogada := models.Cartas{
		Valor: payload.CartaJogada.Valor,
		Naipe: payload.CartaJogada.Naipe,
		Forca: payload.CartaJogada.Forca,
	}

	for _, carta := range vezJogador.Mao {
		if carta == cartaJogada {
			return models.CartaJogada{
				Jogador: vezJogador,
				Carta:   &cartaJogada,
			}, false
		}
	}

	return models.CartaJogada{}, true
}

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
		Type: "error",
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

// ONGOING
func AceitarAposta(m []byte, conn *websocket.Conn) {
	var payload models.AceitarAposta

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
		return
	}
	// só entra aqui se for tudo válido
	switch payload.TipoAposta {
	case TipoTruco:
		AvaliarTruco(sala, r, r.ApostaAtual.ParaTime, payload.Aceitar)
	case TipoRetruco:
		AvaliarRetruco(sala, r, r.ApostaAtual.ParaTime, payload.Aceitar)
	case TipoValeQuatro:
		AvaliarValeQuatro(sala, r, r.ApostaAtual.ParaTime, payload.Aceitar)
	}

}

// ANDAMENTO

func AvaliarValeQuatro(sala *models.Sala, r *models.Rodada, time string, aceitou bool) {
	resposta := models.RespostaAposta{
		Type:       "RESPOSTA_APOSTA",
		TipoAposta: TipoValeQuatro,
		Aceito:     aceitou,
	}

	if aceitou {
		r.ApostaAtual.Estado = EstadoAceito
		r.ValorDaMao = 4
	} else {
		// Caso o Retruco seja recusado, atribui o valor da mão
		r.ApostaAtual.Estado = "RECUSADO"
		switch time {
		case Time01:
			AtribuirPontoTime(&sala.Jogo.Time02, r.ValorDaMao)
		case Time02:
			AtribuirPontoTime(&sala.Jogo.Time01, r.ValorDaMao)
		}

		NotificarRespostaAposta(sala, resposta, time)
		IniciarRodada(sala)
	}

	NotificarRespostaAposta(sala, resposta, time)

}

func AvaliarRetruco(sala *models.Sala, r *models.Rodada, time string, aceitou bool) {
	resposta := models.RespostaAposta{
		Type:       "RESPOSTA_APOSTA",
		TipoAposta: TipoRetruco,
		Aceito:     aceitou,
	}

	if aceitou {
		r.ApostaAtual.Estado = EstadoAceito
		r.ValorDaMao = 3
	} else {
		// Caso o Retruco seja recusado, atribui o valor da mão
		r.ApostaAtual.Estado = "RECUSADO"
		switch time {
		case Time01:
			AtribuirPontoTime(&sala.Jogo.Time02, r.ValorDaMao)
		case Time02:
			AtribuirPontoTime(&sala.Jogo.Time01, r.ValorDaMao)
		}

		NotificarRespostaAposta(sala, resposta, time)
		IniciarRodada(sala)
	}

	NotificarRespostaAposta(sala, resposta, time)
}

func AvaliarTruco(sala *models.Sala, r *models.Rodada, time string, aceitou bool) {
	resposta := models.RespostaAposta{
		Type:       "RESPOSTA_APOSTA",
		TipoAposta: TipoTruco,
		Aceito:     aceitou,
	}

	if aceitou {
		r.ApostaAtual.Estado = EstadoAceito
		r.ValorDaMao = 2
	} else {
		// Caso o Truco seja recusado, atribui o valor da mão
		r.ApostaAtual.Estado = "RECUSADO"
		switch time {
		case Time01:
			AtribuirPontoTime(&sala.Jogo.Time02, r.ValorDaMao)
		case Time02:
			AtribuirPontoTime(&sala.Jogo.Time01, r.ValorDaMao)
		}

		NotificarRespostaAposta(sala, resposta, time)
		IniciarRodada(sala)
	}

	NotificarRespostaAposta(sala, resposta, time)
}

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

func ChamarTruco(m []byte, conn *websocket.Conn) {
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

	if rodadaAtual.VezJogador != jogadorDoTruco {
		responderErro(conn, "Não é a vez do jogador")
		return
	}

	switch jogadorDoTruco.Time {
	case Time01:
		rodadaAtual.ApostaAtual = models.Aposta{
			Tipo:     TipoTruco,
			Estado:   "AGUARDANDO_RESPOSTA",
			ParaTime: Time02,
		}
		EnviarAposta(Time02, sala, TipoTruco)
	case Time02:
		rodadaAtual.ApostaAtual = models.Aposta{
			Tipo:     TipoTruco,
			Estado:   "AGUARDANDO_RESPOSTA",
			ParaTime: Time01,
		}
		EnviarAposta(Time01, sala, TipoTruco)
	}

	sala.Jogo.Estado = "AGUARDANDO_RESPOSTA_APOSTA"
}

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

func BuscarJogador(sala *models.Sala, conn *websocket.Conn) *models.Jogador {
	for _, jogador := range sala.Jogadores {
		if jogador.Conn == conn {
			return jogador
		}
	}

	return nil
}

func ChamarRetruco(m []byte, conn *websocket.Conn) {
	var payload models.IDSala

	json.Unmarshal(m, &payload)

	sala := VerificarSalaExiste(payload.IDSala, conn)

	if sala == nil {
		return
	}

	rodadaAtual := RodadaAtual(sala)
	if !rodadaAtual.Truco {
		responderErro(conn, "Não é possível pedir Retruco")
	}

	jogadorDoTruco := BuscarJogador(sala, conn)

	//if rodadaAtual.VezJogador != jogadorDoTruco {
	// NÃO É A VEZ DO JOGADOR
	//	return
	//}

	//if !(rodadaAtual.ApostaAtual.Tipo == TipoTruco) && !(jogadorDoTruco.Time == rodadaAtual.ApostaAtual.ParaTime) {
	//	responderErro(conn, "Pedido de Retruco inválido")
	//	return
	//}

	switch jogadorDoTruco.Time {
	case Time01:
		rodadaAtual.ApostaAtual = models.Aposta{
			Tipo:     TipoRetruco,
			Estado:   "AGUARDANDO_RESPOSTA",
			ParaTime: Time02,
		}
		EnviarAposta(Time02, sala, TipoRetruco)
	case Time02:
		rodadaAtual.ApostaAtual = models.Aposta{
			Tipo:     TipoRetruco,
			Estado:   "AGUARDANDO_RESPOSTA",
			ParaTime: Time01,
		}
		EnviarAposta(Time01, sala, TipoRetruco)
	}

	sala.Jogo.Estado = "AGUARDANDO_RESPOSTA_APOSTA"
}

func ChamarValeQuatro(m []byte, conn *websocket.Conn) {
	var payload models.IDSala

	json.Unmarshal(m, &payload)

	sala := VerificarSalaExiste(payload.IDSala, conn)

	if sala == nil {
		return
	}

	rodadaAtual := RodadaAtual(sala)
	if !rodadaAtual.Truco {
		responderErro(conn, "Não é possível pedir Retruco")
	}

	jogadorDoTruco := BuscarJogador(sala, conn)

	if rodadaAtual.VezJogador != jogadorDoTruco {
		responderErro(conn, "Não é a vez do jogador")
		return
	}

	if !(rodadaAtual.ApostaAtual.Tipo == TipoRetruco) && !(jogadorDoTruco.Time == rodadaAtual.ApostaAtual.ParaTime) {
		responderErro(conn, "Pedido de Vale Quatro inválido")
		return
	}

	switch jogadorDoTruco.Time {
	case Time01:
		rodadaAtual.ApostaAtual = models.Aposta{
			Tipo:     TipoValeQuatro,
			Estado:   "AGUARDANDO_RESPOSTA",
			ParaTime: Time02,
		}
		EnviarAposta(Time02, sala, TipoValeQuatro)
	case Time02:
		rodadaAtual.ApostaAtual = models.Aposta{
			Tipo:     TipoValeQuatro,
			Estado:   "AGUARDANDO_RESPOSTA",
			ParaTime: Time01,
		}
		EnviarAposta(Time01, sala, TipoValeQuatro)
	}

	sala.Jogo.Estado = "AGUARDANDO_RESPOSTA_APOSTA"
}

func ChamarEnvido(m []byte, conn *websocket.Conn) {}
func CantarFlor(m []byte, conn *websocket.Conn)   {}

func CantarContraFlor(m []byte, conn *websocket.Conn)  {}
func ChamarRealEnvido(m []byte, conn *websocket.Conn)  {}
func ChamarFaltaEnvido(m []byte, conn *websocket.Conn) {}
