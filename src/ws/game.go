package ws

import (
	"encoding/json"
	"fmt"
	"log"
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
	if !ok {
		resposta := models.Resposta{
			Type: "Err",
			Msg:  fmt.Sprintf("A sala com o ID %s não foi encontrada", payload.IdSala),
		}

		data, _ := json.Marshal(resposta)
		conn.WriteMessage(websocket.TextMessage, data)
		return
	}

	if len(sala.Jogadores) >= 2 {
		resposta := models.Resposta{
			Type: "Err",
			Msg:  fmt.Sprintf("A sala com o ID %s já está lotada", payload.IdSala),
		}

		data, _ := json.Marshal(resposta)
		conn.WriteMessage(websocket.TextMessage, data)
		return
	}

	jogador := models.NovoJogador(payload.Nome, conn)

	log.Println(len(sala.Jogo.Time01.Jogadores))

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
		sala.Jogo.EscolherEquipe(payload.TimeEscolhido, jogador)

		resposta := models.Resposta{
			Type: "ok",
			Msg:  "Você entrou no time com sucesso",
		}

		data, _ := json.Marshal(resposta)

		conn.WriteMessage(websocket.TextMessage, data)
	}

	for _, jogador := range sala.Jogo.Time01.Jogadores {
		fmt.Printf("TIME 01: JOGADOR: %s \n", jogador.Nome)
	}

	for _, jogador := range sala.Jogo.Time02.Jogadores {
		fmt.Printf("TIME 02: JOGADOR: %s \n", jogador.Nome)
	}

	// CASO NÃO ACHE O JOGADOR (ADICIONAR COD)
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
