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

	// Dinâmicas da sala
	case "CRIAR_SALA":
		CriarSala(message, conn)
	case "ENTRAR_SALA":
		EntrarSala(message, conn)
	case "ENTRAR_EQUIPE":
		EscolherTime(message, conn)
	case "LISTAR_SALAS":
		ListarSalas(conn)

	// Jogar carta
	case "FAZER_JOGADA":
		FazerJogada(message, conn)

	// Truco e aumentos
	case "CHAMAR_TRUCO":
		CantarTruco(message, conn)
	case "CHAMAR_RETRUCO":
		CantarTruco(message, conn)
	case "CHAMAR_VALE_QUATRO":
		CantarTruco(message, conn)

	// Envidos
	case "CHAMAR_ENVIDO":
		CantarEnvido(message, conn)
	case "CHAMAR_REAL_ENVIDO":
		CantarEnvido(message, conn)
	case "CHAMAR_FALTA_ENVIDO":
		CantarEnvido(message, conn)

	// Flor e Contra-Flor
	case "CANTAR_FLOR":
		CantarFlor(message, conn)
	case "CANTAR_CONTRA_FLOR":
		CantarFlor(message, conn)
	case "CANTAR_CONTRA_FLOR_AL_RESTO":
		CantarFlor(message, conn)

	// Resposta das apostas
	case "RESPONDER_APOSTA":
		ResponderAposta(message, conn)

	// Ir ao mazo
	case "IR_AO_MAZO":
		IrAoMazo(message, conn)

	// Tocar áudio
	case "TOCAR_AUDIO":
		TocarAudio(message, conn)
	}
}
