/* eslint-disable no-console */
const WebSocket = require('ws');
const { SERVER_URL, EVENTS } = require('./config');

const PLAYER_NAME = 'Jogador2';
const TEAM = 'TIME_02';

const ws = new WebSocket(SERVER_URL);

ws.on('open', () => {
  console.log(`[${PLAYER_NAME}] Conectado ao servidor`);

  // aguarda Jogador1 criar a sala e cole o id manualmente
  setTimeout(() => {
    const salaId = '2020';
    ws.send(JSON.stringify({
      type: EVENTS.ENTRAR_SALA,
      nome: PLAYER_NAME,
      idSala: salaId,
    }));
  }, 500);
});

ws.on('message', raw => {
  const msg = JSON.parse(raw);
  console.log(`[${PLAYER_NAME}] Recebido:`, msg);

  switch (msg.type) {
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
      const jogadas = msg.cartasJogadas.length;
      if (msg.vez === PLAYER_NAME && jogadas < 3) {
        ws.send(JSON.stringify({ type: 'FAZER_JOGADA', indiceCarta: jogadas }));
      }
      break;
    }

    case 'error':
      console.error(`[${PLAYER_NAME}] Erro:`, msg.message);
      break;

    default:
      // ignorar
  }
});

ws.on('close', () => console.log(`[${PLAYER_NAME}] Conexão fechada`));
