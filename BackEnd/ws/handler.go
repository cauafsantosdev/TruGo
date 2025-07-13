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
	case "ENTRAR_EQUIPE":
		EscolherTime(message, conn)
	case "LISTAR_SALAS":
		ListarSalas(conn)
	case "FAZER_JOGADA":
		FazerJogada(message, conn)
	case "CHAMAR_TRUCO":
		ChamarTruco(message, conn)
	case "CANTAR_FLOR":
		CantarFlor(message, conn)
	case "CHAMAR_ENVIDO":
		ChamarEnvido(message, conn)
	case "CANTAR_CONTRA_FLOR":
		CantarContraFlor(message, conn)
	case "CHAMAR_REAL_ENVIDO":
		ChamarRealEnvido(message, conn)
	case "CHAMAR_FALTA_ENVIDO":
		ChamarFaltaEnvido(message, conn)
	case "CHAMAR_RETRUCO":
		ChamarRetruco(message, conn)
	case "CHAMAR_VALE_QUATRO":
		ChamarValeQuatro(message, conn)
	case "ACEITAR_APOSTA":
		AceitarAposta(message, conn)
	}
}
