/* eslint-disable no-console */
const WebSocket = require('ws');
const { SERVER_URL, EVENTS } = require('./config');

const PLAYER_NAME = 'Jogador1';
const ROOM_ID = "2020";       // undefined => cria nova sala
const TEAM = 'TIME_01';

const ws = new WebSocket(SERVER_URL);

ws.on('open', () => {
  console.log(`[${PLAYER_NAME}] Conectado ao servidor`);

  // cria sala ou entra em sala existente
  const payload = !ROOM_ID
    ? { type: EVENTS.CRIAR_SALA, nomeJogador: PLAYER_NAME }
    : { type: EVENTS.ENTRAR_SALA, nome: PLAYER_NAME, idSala: ROOM_ID };

  ws.send(JSON.stringify(payload));
});

ws.on('message', raw => {
  const msg = JSON.parse(raw);
  console.log(`[${PLAYER_NAME}] Recebido:`, msg);

  switch (msg.type) {
    case 'SALA_CRIADA':
    case 'ENTRAR_SALA_SUCESSO':
      ws.send(JSON.stringify({
        type: EVENTS.ENTRAR_EQUIPE,
        idSala: msg.idSala,
        time: TEAM,
      }));
      break;

    case EVENTS.MAO_RODADA:
      ws.send(JSON.stringify({ type: 'FAZER_JOGADA', indiceCarta: 0 }));
      break;

    case EVENTS.STATUS_PARTIDA: {
      const proxima = msg.cartasJogadas.length;
      if (msg.vez === PLAYER_NAME && proxima < 3) {
        ws.send(JSON.stringify({ type: 'FAZER_JOGADA', indiceCarta: proxima }));
      }
      break;
    }

    case 'error':
      console.error(`[${PLAYER_NAME}] Erro:`, msg.message);
      break;

    default:
      // silencioso
  }
});

ws.on('close', () => console.log(`[${PLAYER_NAME}] Conexão fechada`));
