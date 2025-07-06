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
		log.Println("Sala já existe!")
	}

	sala := models.Sala{}

	models.Salas[payload.ID] = &sala

	resposta := models.Resposta{
		Type: "ok",
		Msg:  fmt.Sprintf("Sala criada com sucesso, ID: %s", payload.ID),
	}

	data, _ := json.Marshal(resposta)

	conn.WriteMessage(websocket.TextMessage, data)
	log.Println(fmt.Sprintf("Sala criada com sucesso, ID: %s", payload.ID))
}

func EntrarSala(m []byte, conn *websocket.Conn) {
	var payload models.EntrarSala

	if err := json.Unmarshal(m, &payload); err != nil {
		log.Println(err)
		return
	}

	sala, ok := models.Salas[payload.IdSala]
	if !ok {
		log.Println("Sala não encontrada")
		return
	}

	jogador := models.NovoJogador(payload.Nome, conn)

	sala.Jogadores = append(sala.Jogadores, jogador)
	resposta := models.Resposta{
		Type: "Ok",
		Msg:  fmt.Sprintf("Você entrou na sala com o ID %s", payload.IdSala),
	}

	data, _ := json.Marshal(resposta)
	conn.WriteMessage(websocket.TextMessage, data)

	log.Println(fmt.Sprintf("O jogador %s entrou na partida %s", payload.Nome, payload.IdSala))
}
