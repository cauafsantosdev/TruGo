package ws

import (
	"encoding/json"
	"log"
	"trugo/models"

	"github.com/gorilla/websocket"
)

func EscolhaType(message []byte, conn *websocket.Conn) {
	var payload models.Payload
	if err := json.Unmarshal(message, &payload); err != nil {
		log.Println(err)
		return
	}

	switch payload.Type {
	case "CRIAR_SALA":
		CriarSala(message, conn)
	case "ENTRAR_SALA":
		EntrarSala(message, conn)
	}
}
