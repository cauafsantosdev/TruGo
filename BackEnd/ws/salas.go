package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"trugo/models"

	"github.com/gorilla/websocket"
)

// Cria sala
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

// Entra na sala
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

// Escolhe o time
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

	} else {
		responderErro(conn, "Jogador não faz parte da sala")
		return
	}

	if len(sala.Jogo.Time01.Jogadores)+len(sala.Jogo.Time02.Jogadores) == 2 {
		ComecarPartida(sala)
	}
}

// Mostra todas as salas disponiveis
func ListarSalas(conn *websocket.Conn) {
	salasDisponiveis := make(map[string]int)

	for chave, sala := range models.Salas {
		if len(sala.Jogadores) < 2 {
			salasDisponiveis[chave] = 2 - len(sala.Jogadores)
		}
	}

	var payload models.SalasDisponiveis
	payload.SalasDisponiveis = salasDisponiveis

	data, _ := json.Marshal(payload)

	conn.WriteMessage(websocket.TextMessage, data)
}

// Função para procurar um jogador numa lista de jogadores
func ProcurarJogador(listaJogadores []*models.Jogador, conn *websocket.Conn) *models.Jogador {
	for _, jogador := range listaJogadores {
		if jogador.Conn == conn {
			return jogador
		}
	}

	return nil
}
