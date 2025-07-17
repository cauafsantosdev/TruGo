import React, { useState, useRef, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import cartabg from '../assets/cartabg.jpg';

function Jogo() {
    const navigate = useNavigate();
    const nome = localStorage.getItem('nome') || '';
    const time = localStorage.getItem('time') || '';
    const salaId = localStorage.getItem('salaId') || '';

    const [showEnvidoOptions, setShowEnvidoOptions] = useState(false);
    const [mostrarSuaVez, setMostrarSuaVez] = useState(false);
    const [jogador, setJogador] = useState({ nome, time, mao: [] });
    const [sala, setSala] = useState({ id: salaId, status: '', placar: { TIME_01: 0, TIME_02: 0 } });
    const [cartasMesa, setCartasMesa] = useState([]);
    const [apostasDisponiveis, setApostasDisponiveis] = useState({});
    const [rodadaAtual, setRodadaAtual] = useState(1);
    const [mensagem, setMensagem] = useState('');
    const [apostaDialog, setApostaDialog] = useState(null);
    const [alertaVencedor, setAlertaVencedor] = useState(null);
    const [conexaoStatus, setConexaoStatus] = useState('Desconectado');
    const [maoFinalizada, setMaoFinalizada] = useState(false);
    const [chamouTruco, setChamouTruco] = useState(false);
    const wsRef = useRef(null);
    const reconnectTimeoutRef = useRef(null);
    const reconnectAttemptsRef = useRef(0);
    const maxReconnectAttempts = 10;

    const sincronizarCartasMesa = useCallback((cartasJogadas) => {
        if (!cartasJogadas || cartasJogadas.length === 0) {
            setCartasMesa([]);
            return;
        }
        const cartasNormalizadas = cartasJogadas
            .filter(carta => carta && carta.jogador && (carta.cartaJogada || carta.carta))
            .map((carta, index) => ({
                jogador: carta.jogador,
                cartaJogada: carta.cartaJogada || carta.carta,
                id: `${carta.jogador}-${index}`
            }));
        setCartasMesa(cartasNormalizadas);
    }, []);

    const conectarWebSocket = useCallback(() => {
        try {
            setConexaoStatus('Conectando...');
            if (wsRef.current && wsRef.current.readyState !== WebSocket.CLOSED) {
                wsRef.current.close();
            }
            wsRef.current = new WebSocket("ws://192.168.2.101:8080/ws");

            wsRef.current.onopen = () => {
                setConexaoStatus('Conectado');
                setMensagem("Conectado ao servidor!");
                reconnectAttemptsRef.current = 0;
                const entryJson = localStorage.getItem('entryData');
                if (entryJson) {
                    try {
                        const { nome: nomeEntry, salaId: salaEntry } = JSON.parse(entryJson);
                        wsRef.current.send(JSON.stringify({ type: 'ENTRAR_SALA', nome: nomeEntry, idSala: salaEntry }));
                    } catch (parseError) {
                        console.error("Erro ao processar entryData:", parseError);
                    }
                }
            };

            wsRef.current.onmessage = (event) => {
                try {
                    const rawData = JSON.parse(event.data);
                    const data = { ...rawData, type: (String(rawData.type) || '').toUpperCase().trim() };
                    console.log('WS event recebido:', data.type, rawData);
                    if (reconnectTimeoutRef.current) {
                        clearTimeout(reconnectTimeoutRef.current);
                    }
                    switch (data.type) {
                        case 'PONTOS_ENVIDO': {
                            const equipePayload = data.Equipe || data.equipe || {};
                            const pontosT1 = typeof equipePayload.TIME_01 === 'number' ? equipePayload.TIME_01 : 0;
                            const pontosT2 = typeof equipePayload.TIME_02 === 'number' ? equipePayload.TIME_02 : 0;
                            let vencedor = '';
                            if (pontosT1 > pontosT2) vencedor = 'Time 1';
                            else if (pontosT2 > pontosT1) vencedor = 'Time 2';
                            else vencedor = pontosT1 === 0 && pontosT2 === 0 ? '-' : 'Empate';

                            if (data.placar) {
                                setSala(s => ({ ...s, placar: data.placar }));
                            }

                            // Verificar se é resultado de Flor ou Envido baseado no contexto
                            const tipoAposta = apostaDialog?.tipoAposta === 'FLOR' || apostaDialog?.tipoAposta?.includes('FLOR') ? 'FLOR' : 'ENVIDO';
                            
                            setApostaDialog({
                                tipoAposta: tipoAposta,
                                aguardandoResposta: false,
                                pontosT1,
                                pontosT2,
                                vencedor,
                                msgAguardo: `Time 1: ${pontosT1}, Time 2: ${pontosT2}. Vencedor: ${vencedor}`
                            });
                            setTimeout(() => setApostaDialog(null), 3500);
                            break;
                        }
                        case 'PONTOS_ENVIDO_GANHOS':
                            if (data.placar) {
                                setSala(s => ({ ...s, placar: data.placar }));
                            }
                            break;
                        case 'SALA_CRIADA':
                        case 'ENTRAR_SALA_SUCESSO': {
                            const entryJson = localStorage.getItem('entryData');
                            let timeEntry = time;
                            if (entryJson) {
                                const entry = JSON.parse(entryJson);
                                timeEntry = entry.time || time;
                            }
                            wsRef.current.send(JSON.stringify({ type: 'ENTRAR_EQUIPE', idSala: data.idSala, timeEscolhido: timeEntry }));
                            setSala(s => ({ ...s, id: data.idSala }));
                            break;
                        }
                        case 'OK': {
                            const entryJson = localStorage.getItem('entryData');
                            let timeEntry = time;
                            let salaEntry = salaId;
                            if (entryJson) {
                                const entry = JSON.parse(entryJson);
                                timeEntry = entry.time || time;
                                salaEntry = entry.salaId || salaId;
                            }
                            wsRef.current.send(JSON.stringify({ type: 'ENTRAR_EQUIPE', idSala: data.idSala || salaEntry, timeEscolhido: timeEntry }));
                            setSala(s => ({ ...s, id: data.idSala || salaEntry }));
                            break;
                        }
                        case 'MAO_RODADA':
                            setMensagem('Sua mão foi distribuída!');
                            setTimeout(() => {
                                setJogador(j => ({ ...j, mao: data.mao || [] }));
                                setCartasMesa([]);
                            }, 1500);
                            break;
                        case 'SUA_VEZ':
                            console.log('SUA_VEZ recebido para:', nome);
                            setMostrarSuaVez(true);
                            setApostasDisponiveis(data.apostasDisponiveis || {});
                            setSala(s => ({ ...s, placar: data.placar || s.placar }));
                            if (data.cartasJogadas && Array.isArray(data.cartasJogadas)) {
                                const jogadores = {};
                                data.cartasJogadas.forEach(cj => {
                                    if (cj.jogador) {
                                        jogadores[cj.jogador] = (jogadores[cj.jogador] || 0) + 1;
                                    }
                                });
                                const maxCartas = Math.max(0, ...Object.values(jogadores));
                                setRodadaAtual(maxCartas + 1);
                            } else {
                                setRodadaAtual(1);
                            }
                            if (data.cartasJogadas?.length > 0) {
                                sincronizarCartasMesa(data.cartasJogadas);
                            }
                            break;
                        case 'STATUS_PARTIDA':
                            setApostasDisponiveis(data.apostasDisponiveis || {});
                            if (data.cartasJogadas && Array.isArray(data.cartasJogadas)) {
                                const jogadores = {};
                                data.cartasJogadas.forEach(cj => {
                                    if (cj.jogador) {
                                        jogadores[cj.jogador] = (jogadores[cj.jogador] || 0) + 1;
                                    }
                                });
                                const maxCartas = Math.max(0, ...Object.values(jogadores));
                                setRodadaAtual(maxCartas + 1);
                            } else {
                                setRodadaAtual(1);
                            }
                            if (data.cartasJogadas && data.cartasJogadas.length > 0) {
                                sincronizarCartasMesa(data.cartasJogadas);
                            }
                            setSala(s => ({ ...s, placar: data.placar || s.placar }));
                            break;
                        case 'CARTA_JOGADA':
                            if (data.cartaJogada && data.jogador) {
                                if (data.jogador !== nome) {
                                    setCartasMesa(prev => {
                                        const cartaExiste = prev.find(c => 
                                            c.jogador === data.jogador && 
                                            c.cartaJogada?.naipe === data.cartaJogada.naipe &&
                                            c.cartaJogada?.valor === data.cartaJogada.valor
                                        );
                                        if (cartaExiste) return prev;
                                        return [...prev, {
                                            jogador: data.jogador,
                                            cartaJogada: data.cartaJogada,
                                            id: `${data.jogador}-${Date.now()}`
                                        }];
                                    });
                                    setMensagem(`${data.jogador} jogou ${data.cartaJogada.valor} de ${data.cartaJogada.naipe}`);
                                } else {
                                    setCartasMesa(prev => {
                                        const cartaExiste = prev.find(c => 
                                            c.jogador === nome && 
                                            c.cartaJogada?.naipe === data.cartaJogada.naipe &&
                                            c.cartaJogada?.valor === data.cartaJogada.valor
                                        );
                                        if (cartaExiste) return prev;
                                        return [...prev, {
                                            jogador: nome,
                                            cartaJogada: data.cartaJogada,
                                            id: `${nome}-${Date.now()}`
                                        }];
                                    });
                                    console.log('Minha carta confirmada:', data.cartaJogada);
                                }
                            }
                            break;
                        case 'APOSTA': {
                            console.log('APOSTA recebida:', data);
                            const tipoAposta = data.TipoDeAposta || data.tipoAposta || data.aposta || data.type;
                            // Se estava aguardando resposta, fecha o diálogo
                            if (apostaDialog && apostaDialog.aguardandoResposta) {
                                setApostaDialog(null);
                            }
                            if (tipoAposta) {
                                console.log('Definindo apostaDialog com tipoAposta:', tipoAposta);
                                
                                // Tratamento específico para apostas de Flor
                                if (tipoAposta === 'FLOR' || tipoAposta === 'CONTRA_FLOR' || tipoAposta === 'CONTRA_FLOR_AL_RESTO') {
                                    setApostaDialog({ 
                                        tipoAposta,
                                        aguardandoResposta: false,
                                        opcoesResposta: tipoAposta === 'FLOR' ? ['ACEITAR', 'RECUSAR', 'CONTRA_FLOR', 'CONTRA_FLOR_AL_RESTO'] : ['ACEITAR', 'RECUSAR']
                                    });
                                } else {
                                    setApostaDialog({ 
                                        tipoAposta,
                                        aguardandoResposta: false
                                    });
                                }
                            } else {
                                console.log('Nenhum tipoAposta encontrado no payload APOSTA');
                            }
                            break;
                        }
                        case 'RESPOSTA_APOSTA':
                            setMensagem(`Aposta ${data.TipoDeAposta || data.tipoAposta} ${data.Quero || data.aceito ? 'aceita' : 'recusada'}`);
                            setApostaDialog(null);
                            if (data.apostasDisponiveis) {
                                setApostasDisponiveis(data.apostasDisponiveis);
                            }
                            if ((data.Quero || data.aceito) && data.proximoJogador === nome) {
                                setMostrarSuaVez(true);
                            }
                            break;
                        case 'BOA': {
                            // Flor aceita/cantada com sucesso - 3 pontos diretos
                            setMensagem('Flor boa! Você ganhou 3 pontos');
                            if (data.placar) {
                                setSala(s => ({ ...s, placar: data.placar }));
                            }
                            setApostaDialog(null);
                            setMostrarSuaVez(false);
                            break;
                        }
                        case 'FLOR_CANTADA': {
                            // Adversário cantou flor - verificar se você também tem flor
                            if (data.RespostaParaFlor || data.apostaFlor) {
                                // Ambos têm flor - abre diálogo de aposta
                                setApostaDialog({
                                    tipoAposta: 'FLOR',
                                    aguardandoResposta: true,
                                    msgAguardo: 'Adversário cantou Flor! Você também tem flor. Responder?',
                                    podeAumentar: true,
                                    opcoesResposta: ['ACEITAR', 'RECUSAR', 'CONTRA_FLOR', 'CONTRA_FLOR_AL_RESTO']
                                });
                            } else {
                                // Só o adversário tem flor - mostra mensagem
                                setMensagem('Adversário cantou Flor e ganhou 3 pontos');
                                if (data.placar) {
                                    setSala(s => ({ ...s, placar: data.placar }));
                                }
                            }
                            break;
                        }
                        case 'MAZO':
                        case 'JOGADOR_FOI_AO_MAZO': {
                            const jogadorMazo = data.jogador || data.nome || 'Jogador';
                            setMensagem(`${jogadorMazo} foi ao mazo - desistiu da rodada`);
                            setMostrarSuaVez(false);
                            setApostaDialog(null);
                            setMaoFinalizada(true);

                            // Parseando a mensagem para extrair informações
                            const regex = /A equipe (TIME_\d{2}) ganha (\d+) ponto/;
                            const match = data.message.match(regex);
                            let timeVencedor = match ? match[1] : null;
                            let pontosGanhos = match ? parseInt(match[2], 10) : 1;

                            let timeQueFoiAoMazo = null;
                            if (!timeVencedor) {
                                if (time === 'TIME_01') {
                                    timeVencedor = 'TIME_02';
                                } else if (time === 'TIME_02') {
                                    timeVencedor = 'TIME_01';
                                } else {
                                    timeVencedor = 'TIME_01';
                                }
                            }
                            if (timeVencedor === 'TIME_01') {
                                timeQueFoiAoMazo = 'TIME_02';
                            } else if (timeVencedor === 'TIME_02') {
                                timeQueFoiAoMazo = 'TIME_01';
                            } else {
                                timeQueFoiAoMazo = 'Adversário';
                            }

                            setAlertaVencedor({
                                time: timeVencedor,
                                pontos: pontosGanhos,
                                foiAoMazo: true,
                                timeQueFoiAoMazo: timeQueFoiAoMazo
                            });
                            setTimeout(() => setAlertaVencedor(null), 2500);
                            if (data.placar) {
                                setSala(s => ({ ...s, placar: data.placar }));
                            }
                            setTimeout(() => {
                                setCartasMesa([]);
                                setJogador(j => ({ ...j, mao: [] }));
                                setMaoFinalizada(false);
                            }, 1500);
                            break;
                        }
                        case 'APOSTA_ACEITA':
                        case 'APOSTA_RECUSADA':
                            setMensagem(`Aposta ${data.TipoDeAposta || data.tipoAposta || ''} ${data.type === 'APOSTA_ACEITA' ? 'aceita' : 'recusada'}`);
                            setApostaDialog(null);
                            if (data.apostasDisponiveis) {
                                setApostasDisponiveis(data.apostasDisponiveis);
                            }
                            if (data.placar) {
                                setSala(s => ({ ...s, placar: data.placar }));
                            }
                            break;
                            break;
                        case 'RODADA_FINALIZADA':
                        case 'MAO_FINALIZADA':
                        case 'NOVA_RODADA':
                        case 'VENCEDOR_MAO':
                        case 'LIMPAR_MESA':
                            if (data.type === 'MAO_FINALIZADA') {
                                setMensagem("Mão finalizada!");
                                setMaoFinalizada(true);
                                if (data.timeVencedor && typeof data.pontosGanhos === 'number') {
                                    setAlertaVencedor({
                                        time: data.timeVencedor,
                                        pontos: data.pontosGanhos,
                                        foiAoMazo: false
                                    });
                                    setTimeout(() => setAlertaVencedor(null), 2500);
                                }
                                if (data.cartasJogadas && data.cartasJogadas.length > 0) {
                                    sincronizarCartasMesa(data.cartasJogadas);
                                    setTimeout(() => {
                                        setCartasMesa([]);
                                        setJogador(j => ({ ...j, mao: [] }));
                                        setMaoFinalizada(false);
                                    }, 1500);
                                } else {
                                    setJogador(j => ({ ...j, mao: [] }));
                                    setMaoFinalizada(false);
                                }
                            } else if (data.type === 'VENCEDOR_MAO') {
                                setMensagem(`${data.vencedor} venceu a mão!`);
                                setMostrarSuaVez(false);
                                if (data.cartasJogadas && data.cartasJogadas.length > 0) {
                                    sincronizarCartasMesa(data.cartasJogadas);
                                    setTimeout(() => setCartasMesa([]), 1500);
                                }
                            } else if (data.type === 'NOVA_RODADA') {
                                setMensagem('Nova rodada iniciada!');
                                setTimeout(() => setCartasMesa([]), 1500);
                            } else {
                                setTimeout(() => setCartasMesa([]), 1500);
                            }
                            break;
                        case 'error':
                            setMensagem(data.msg);
                            break;
                        case 'PARTIDA_FINALIZADA': {
                            const mensagemFinal = data.message || data.Mensagem;
                            const placarFinal = data.placar || data.Placar;

                            setApostaDialog({
                                tipoAposta: null,
                                aguardandoResposta: false,
                                msgAguardo: null,
                                mensagemFinal,
                                placarFinal
                            });
                            break;
                        }
                    }
                } catch (e) {
                    console.error("Erro ao processar mensagem:", e);
                }
            };

            wsRef.current.onclose = (event) => {
                setConexaoStatus('Desconectado');
                const closeReasons = {
                    1000: 'Fechamento normal',
                    1006: 'Conexão perdida',
                    1011: 'Erro interno do servidor'
                };
                const reason = closeReasons[event.code] || `Código: ${event.code}`;
                setMensagem(`Conexão fechada: ${reason}`);
                if (event.code !== 1000 && reconnectAttemptsRef.current < maxReconnectAttempts) {
                    reconnectAttemptsRef.current++;
                    const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current - 1), 30000);
                    setMensagem(`Reconectando... (${reconnectAttemptsRef.current}/${maxReconnectAttempts})`);
                    reconnectTimeoutRef.current = setTimeout(() => {
                        conectarWebSocket();
                    }, delay);
                } else if (event.code !== 1000) {
                    setMensagem("Falha ao conectar. Máximo de tentativas atingido.");
                }
            };

            wsRef.current.onerror = (error) => {
                setConexaoStatus('Erro');
                setMensagem("Erro de conexão");
                if (navigator.onLine === false) {
                    setConexaoStatus('Sem rede');
                    setMensagem("Sem conexão com a internet");
                }
            };

        } catch (error) {
            setConexaoStatus('Erro');
            console.error("Erro ao conectar WebSocket:", error);
            if (reconnectAttemptsRef.current < maxReconnectAttempts) {
                reconnectAttemptsRef.current++;
                const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current - 1), 30000);
                setMensagem(`Erro ao conectar. Tentativa ${reconnectAttemptsRef.current}/${maxReconnectAttempts}`);
                reconnectTimeoutRef.current = setTimeout(() => {
                    conectarWebSocket();
                }, delay);
            } else {
                setMensagem("Falha crítica ao estabelecer conexão");
            }
        }
    }, []);

    const reconectarManualmente = useCallback(() => {
        if (wsRef.current) {
            wsRef.current.close();
        }
        if (reconnectTimeoutRef.current) {
            clearTimeout(reconnectTimeoutRef.current);
        }
        reconnectAttemptsRef.current = 0;
        conectarWebSocket();
    }, [conectarWebSocket]);

    const handleHideEnvidoOptions = (e) => {
        if (
            !e.target.closest('.envido-trigger') &&
            !e.target.closest('.envido-options')
        ) {
            setShowEnvidoOptions(false);
        }
    };

    const resetarMesaManual = () => {
        setCartasMesa([]);
        setMostrarSuaVez(false);
        setMensagem("Mesa resetada manualmente");
        console.log("🧹 Reset manual da mesa executado");
    };

    const enviarEntrarEquipeAutomatico = useCallback(() => {
        const entryJson = localStorage.getItem('entryData');
        let timeEntry = time;
        let salaEntry = salaId;
        if (entryJson) {
            const entry = JSON.parse(entryJson);
            timeEntry = entry.time || time;
            salaEntry = entry.salaId || salaId;
        }
        if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
            wsRef.current.send(JSON.stringify({ type: 'ENTRAR_EQUIPE', idSala: salaEntry, timeEscolhido: timeEntry }));
            setMensagem('Payload ENTRAR_EQUIPE enviado automaticamente!');
        }
    }, [time, salaId]);

    useEffect(() => {
        conectarWebSocket();
        return () => {
            if (reconnectTimeoutRef.current) { 
                clearTimeout(reconnectTimeoutRef.current);
            }
            if (wsRef.current) {
                wsRef.current.close(1000, 'Component unmounting');
            }
        };
    }, [conectarWebSocket]);

    useEffect(() => {
        if (conexaoStatus === 'Conectado') {
            enviarEntrarEquipeAutomatico();
        }
    }, [conexaoStatus, enviarEntrarEquipeAutomatico]);
    useEffect(() => {
        const handleOnline = () => {
            if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
                reconectarManualmente();
            }
        };

        const handleOffline = () => {
            setConexaoStatus('Sem rede');
            setMensagem("Conexão de rede perdida");
        };

        window.addEventListener('online', handleOnline);
        window.addEventListener('offline', handleOffline);

        return () => {
            window.removeEventListener('online', handleOnline);
            window.removeEventListener('offline', handleOffline);
        };
    }, [reconectarManualmente]);

    useEffect(() => {
        if (showEnvidoOptions) {
            document.addEventListener('mousedown', handleHideEnvidoOptions);
        } else {
            document.removeEventListener('mousedown', handleHideEnvidoOptions);
        }
        return () => {
            document.removeEventListener('mousedown', handleHideEnvidoOptions);
        };
    }, [showEnvidoOptions]);

    useEffect(() => {
        const handleKeyPress = (event) => {
            if (event.key.toLowerCase() === 'j') {
                if (event.target.tagName.toLowerCase() !== 'input' && 
                    event.target.tagName.toLowerCase() !== 'textarea') {
                    resetarMesaManual();
                }
            }
        };

        document.addEventListener('keydown', handleKeyPress);
        return () => {
            document.removeEventListener('keydown', handleKeyPress);
        };
    }, []);

    const enviarAcao = (tipo, valor = null) => {
        if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
            try {
                let payload = { type: tipo, idSala: salaId };
                if (tipo === 'FAZER_JOGADA' && valor) {
                    payload = { ...payload, cartaJogada: valor };
                } else if (tipo === 'ACEITAR_APOSTA' && valor) {
                    payload = { ...payload, tipoAposta: valor.tipoAposta, aceitar: valor.aceitar };
                } else if (valor !== null) {
                    payload = { ...payload, ...valor };
                }
                wsRef.current.send(JSON.stringify(payload));
            } catch (error) {
                console.error("Erro ao enviar ação:", error);
                setMensagem("Erro ao enviar ação");
                reconectarManualmente();
            }
        } else {
            setMensagem("Conexão perdida. Reconectando...");
            reconectarManualmente();
        }
    };

    const jogarCarta = (carta) => {
        if (maoFinalizada) return;
        setJogador(j => ({
            ...j,
            mao: j.mao.filter(c => !(c.naipe === carta.naipe && c.valor === carta.valor))
        }));
        setCartasMesa(prev => [...prev, {
            jogador: nome,
            cartaJogada: carta,
            id: `${nome}-${Date.now()}`
        }]);
        setMostrarSuaVez(false);
        setMensagem(`Você jogou ${carta.valor} de ${carta.naipe}`);
        enviarAcao('FAZER_JOGADA', carta);
    };

    // Flor só é verdadeira se todas as cartas da mão forem do mesmo naipe
    const temFlor = useCallback(() => {
        if (!jogador.mao || jogador.mao.length < 3) return false;
        const naipes = jogador.mao.map(carta => carta.naipe);
        const primeiroNaipe = naipes[0];
        return naipes.every(naipe => naipe === primeiroNaipe);
    }, [jogador.mao]);

    const obterProximoTruco = useCallback(() => {
        if (apostasDisponiveis.ValeQuatro && chamouTruco) return 'CHAMAR_VALE_QUATRO';
        if (apostasDisponiveis.Retruco && !chamouTruco) return 'CHAMAR_RETRUCO';
        if (apostasDisponiveis.Truco) return 'CHAMAR_TRUCO';
        return null;
    }, [apostasDisponiveis, chamouTruco]);

    const obterTextoTruco = useCallback(() => {
        if (apostasDisponiveis.ValeQuatro) return 'VALE QUATRO';
        if (apostasDisponiveis.Retruco) return 'RETRUCO';
        if (apostasDisponiveis.Truco) return 'TRUCO';
        return 'TRUCO';
    }, [apostasDisponiveis]);

    const pedirAposta = (tipoAposta) => {
        if (!mostrarSuaVez) {
            console.log('Não é sua vez! mostrarSuaVez:', mostrarSuaVez);
            return;
        }

        let tipoFinal = tipoAposta;

        const payloadAposta = {
            type: tipoFinal,
            idSala: sala.id
        };

        if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
            wsRef.current.send(JSON.stringify(payloadAposta));
            console.log('Enviando aposta:', payloadAposta);

            if (tipoFinal === 'CHAMAR_TRUCO') {
                setChamouTruco(true);
            }

            if (tipoFinal === 'CANTAR_FLOR') {
                setMensagem('Flor cantada! Adicionando 3 pontos ao placar.');
                setSala(s => ({
                    ...s,
                    placar: {
                        ...s.placar,
                        [jogador.time]: s.placar[jogador.time] + 3
                    }
                }));
                setMostrarSuaVez(false);
            } else {
                setApostaDialog({
                    tipoAposta: tipoFinal,
                    aguardandoResposta: true,
                    msgAguardo: `Aguardando resposta do ${tipoFinal.replace('CHAMAR_', '').replace('_', ' ')}...`
                });
            }
        } else {
            console.log('WebSocket não está conectado!');
        }
    };

    const mapearTipoAposta = (tipoCompleto) => {
        const mapeamento = {
            'CHAMAR_TRUCO': 'TRUCO',
            'CHAMAR_RETRUCO': 'RETRUCO', 
            'CHAMAR_VALE_QUATRO': 'VALE_QUATRO',
            'CHAMAR_ENVIDO': 'ENVIDO',
            'CHAMAR_REAL_ENVIDO': 'REAL_ENVIDO',
            'CHAMAR_FALTA_ENVIDO': 'FALTA_ENVIDO',
            'CANTAR_FLOR': 'FLOR',
            'CHAMAR_FLOR': 'FLOR',
            'CANTAR_CONTRA_FLOR': 'CONTRA_FLOR',
            'CANTAR_CONTRA_FLOR_AL_RESTO': 'CONTRA_FLOR_AL_RESTO'
        };
        return mapeamento[tipoCompleto] || tipoCompleto;
    };

    const responderAposta = (aceitar) => {
        if (!apostaDialog || !apostaDialog.tipoAposta) return;
        const payload = {
            type: 'RESPONDER_APOSTA',
            tipoAposta: mapearTipoAposta(apostaDialog.tipoAposta),
            aceitar: aceitar,
            idSala: sala.id
        };
        if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
            wsRef.current.send(JSON.stringify(payload));
            console.log('Enviando resposta de aposta:', payload);
        }
        setApostaDialog(null);
    };

    const aumentarAposta = (novoTipoAposta) => {
        const payload = {
            type: novoTipoAposta,
            idSala: sala.id
        };
        if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
            wsRef.current.send(JSON.stringify(payload));
            console.log('Enviando aumento de aposta:', payload);
            setApostaDialog({ 
                tipoAposta: novoTipoAposta,
                aguardandoResposta: true,
                msgAguardo: `Aguardando resposta do ${novoTipoAposta.replace('CHAMAR_', '').replace('_', ' ')}...`
            });
        } else {
            console.log('WebSocket não está conectado!');
        }
    };

    return (
        <div 
            className="w-full h-screen flex flex-col items-center justify-center relative"
            style={{
                backgroundImage: "url('/background.jpg')",
                backgroundSize: 'cover',
                backgroundPosition: 'center',
                backgroundRepeat: 'no-repeat'
            }}
        >
            <img 
                src="/Trugo_logo-removebg-preview(2).png" 
                alt="Logo TruGo" 
                className="absolute top-4 right-4 w-44 h-auto z-30 drop-shadow-lg select-none"
            />
            {apostaDialog && (
                <div className="fixed inset-0 z-50 flex items-center justify-center" style={{ backgroundColor: 'rgba(0, 0, 0, 0.7)', backdropFilter: 'blur(4px)' }}>
                    <div className="border-2 border-amber-600 rounded-xl shadow-2xl px-10 py-8 flex flex-col items-center w-[340px]" style={{ backgroundColor: 'rgba(92, 70, 0, 0.9)', backdropFilter: 'blur(8px)' }}>
                        {apostaDialog.aguardandoResposta ? (
                            <>
                                <h2 className="text-2xl font-bold text-amber-300 mb-4">Aposta enviada!</h2>
                                <span className="text-lg text-gray-200 mb-6 text-center">{apostaDialog.msgAguardo}</span>
                                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-amber-300 mb-4"></div>
                            </>
                        ) : (apostaDialog.tipoAposta === 'ENVIDO' || apostaDialog.tipoAposta === 'FLOR') && typeof apostaDialog.pontosT1 === 'number' && typeof apostaDialog.pontosT2 === 'number' ? (
                            <>
                                <h2 className="text-2xl font-bold text-amber-300 mb-4">Resultado do {apostaDialog.tipoAposta}</h2>
                                <span className="text-lg text-gray-200 mb-4">Time 1: {apostaDialog.pontosT1} | Time 2: {apostaDialog.pontosT2}</span>
                                <span className="text-lg text-amber-300 mb-2">Vencedor: {apostaDialog.vencedor}</span>
                            </>
                        ) : apostaDialog.tipoAposta === 'FLOR' && apostaDialog.opcoesResposta ? (
                            <>
                                <h2 className="text-2xl font-bold text-amber-300 mb-4">Flor Cantada!</h2>
                                <span className="text-lg text-gray-200 mb-6">{apostaDialog.msgAguardo || 'Adversário cantou Flor!'}</span>
                                <div className="flex flex-col space-y-3 w-full">
                                    <button
                                        className="bg-green-700 text-white font-bold py-2 rounded shadow hover:bg-green-800 transition-colors"
                                        onClick={() => responderAposta(true)}
                                    >Quero (Flor)</button>
                                    <button
                                        className="bg-red-700 text-white font-bold py-2 rounded shadow hover:bg-red-800 transition-colors"
                                        onClick={() => responderAposta(false)}
                                    >Não quero</button>
                                    {apostasDisponiveis.ContraFlor && (
                                        <button
                                            className="bg-yellow-700 text-white font-bold py-2 rounded shadow hover:bg-yellow-800 transition-colors"
                                            onClick={() => aumentarAposta('CANTAR_CONTRA_FLOR')}
                                        >Contra Flor</button>
                                    )}
                                    {apostasDisponiveis.ContraFlorAlResto && (
                                        <button
                                            className="bg-purple-700 text-white font-bold py-2 rounded shadow hover:bg-purple-800 transition-colors"
                                            onClick={() => aumentarAposta('CANTAR_CONTRA_FLOR_AL_RESTO')}
                                        >Contra Flor al Resto</button>
                                    )}
                                </div>
                            </>
                        ) : (
                            <>
                                <h2 className="text-2xl font-bold text-amber-300 mb-4">Aposta recebida!</h2>
                                <span className="text-lg text-gray-200 mb-6">{apostaDialog.tipoAposta ? apostaDialog.tipoAposta.replace(/_/g, ' ').replace('CHAMAR ', '').replace('CANTAR ', '') : 'Aposta'}</span>
                                <div className="flex flex-col space-y-3 w-full">
                                    <button
                                        className="bg-green-700 text-white font-bold py-2 rounded shadow hover:bg-green-800 transition-colors"
                                        onClick={() => responderAposta(true)}
                                    >Quero</button>
                                    <button
                                        className="bg-red-700 text-white font-bold py-2 rounded shadow hover:bg-red-800 transition-colors"
                                        onClick={() => responderAposta(false)}
                                    >Não quero</button>
                                    {rodadaAtual === 1 && apostaDialog.tipoAposta === 'ENVIDO' && apostasDisponiveis.RealEnvido && (
                                        <button
                                            className="bg-yellow-700 text-white font-bold py-2 rounded shadow hover:bg-yellow-800 transition-colors"
                                            onClick={() => aumentarAposta('CHAMAR_REAL_ENVIDO')}
                                        >Aumentar para Real Envido</button>
                                    )}
                                    {rodadaAtual === 1 && apostaDialog.tipoAposta === 'ENVIDO' && apostasDisponiveis.FaltaEnvido && (
                                        <button
                                            className="bg-yellow-900 text-white font-bold py-2 rounded shadow hover:bg-yellow-800 transition-colors"
                                            onClick={() => aumentarAposta('CHAMAR_FALTA_ENVIDO')}
                                        >Aumentar para Falta Envido</button>
                                    )}
                                    {rodadaAtual === 1 && apostaDialog.tipoAposta === 'REAL_ENVIDO' && apostasDisponiveis.FaltaEnvido && (
                                        <button
                                            className="bg-yellow-900 text-white font-bold py-2 rounded shadow hover:bg-yellow-800 transition-colors"
                                            onClick={() => aumentarAposta('CHAMAR_FALTA_ENVIDO')}
                                        >Aumentar para Falta Envido</button>
                                    )}
                                </div>
                            </>
                        )}
                    </div>
                </div>
            )}
            {alertaVencedor && (
                <div className="absolute left-[12%] z-30 flex flex-col items-center w-64"
                    style={{ top: 'calc(50% - 210px)' }}>
                    <div className="bg-yellow-950/60 text-gray-300 px-8 py-2 rounded-xl shadow-lg border-2 border-amber-600 font-extrabold text-base text-center animate-fadeIn w-full">
                        {alertaVencedor.foiAoMazo ? (
                            <>
                                {alertaVencedor.timeQueFoiAoMazo === 'TIME_01' ? 'Time 1' : alertaVencedor.timeQueFoiAoMazo === 'TIME_02' ? 'Time 2' : 'Adversário'} foi ao mazo!<br />
                                <span className="text-amber-300">{alertaVencedor.time === 'TIME_01' ? 'Time 1' : alertaVencedor.time === 'TIME_02' ? 'Time 2' : alertaVencedor.time} ganhou +{alertaVencedor.pontos} ponto{alertaVencedor.pontos > 1 ? 's' : ''}</span>
                            </>
                        ) : (
                            <>
                                {alertaVencedor.time === 'TIME_01' ? 'Time 1' : alertaVencedor.time === 'TIME_02' ? 'Time 2' : alertaVencedor.time} venceu a mão!<br />
                                <span className="text-amber-300">+{alertaVencedor.pontos} ponto{alertaVencedor.pontos > 1 ? 's' : ''}</span>
                            </>
                        )}
                    </div>
                </div>
            )}
            <div className="absolute left-[12%] top-1/2 transform -translate-y-1/2 z-20 flex flex-col items-center">
                <div className="bg-yellow-950/60 text-gray-300 px-8 py-6 rounded-xl mb-6 shadow-lg flex flex-col items-center w-64 border-2 border-amber-600">
                    <h3 className="text-2xl font-extrabold border-b-2 border-amber-600 pb-2 mb-4 text-center">Placar</h3>
                    <div className="flex justify-between w-full mb-2">
                        <span className="text-lg font-bold text-gray-300">Time 1</span>
                        <span className="text-3xl font-extrabold text-gray-300">{sala.placar.TIME_01}</span>
                    </div>
                    <div className="flex justify-between w-full mb-4">
                        <span className="text-lg font-bold text-gray-300">Time 2</span>
                        <span className="text-3xl font-extrabold text-gray-300">{sala.placar.TIME_02}</span>
                    </div>
                </div>
            </div>
            {mostrarSuaVez && (
                <div className="fixed bottom-8 right-8 z-50 bg-green-600 text-white px-6 py-3 rounded-lg shadow-lg animate-pulse">
                    <span className="text-lg font-bold">🎯 Sua vez!</span>
                </div>
            )}
            <div className="relative z-10 flex flex-col items-center justify-center space-y-6">
                <div className="flex justify-center space-x-4 mb-8">
                    {(() => {
                        // Calcula quantas cartas o adversário já jogou
                        const cartasAdversarioJogadas = cartasMesa.filter(c => c.jogador !== nome).length;
                        const cartasRestantes = 3 - cartasAdversarioJogadas;
                        return Array.from({ length: 3 }).map((_, idx) => {
                            const mostrarBack = idx < cartasRestantes;
                            return (
                                <div
                                    key={`adv-hand-${idx}`}
                                    className="w-36 h-56 rounded-lg shadow-lg"
                                    style={{
                                        backgroundColor: mostrarBack ? 'white' : 'transparent',
                                        backgroundImage: mostrarBack ? `url(${cartabg})` : 'none',
                                        backgroundSize: 'cover',
                                        backgroundPosition: 'center'
                                    }}
                                />
                            );
                        });
                    })()}
                </div>
                <div className="flex justify-center items-center my-8" style={{ minHeight: '200px' }}>
                    <div className="flex flex-col items-center space-y-4">
                        <div className="flex justify-start space-x-24" style={{ width: '432px' }}>
                            {(() => {
                                const cartasAdversario = cartasMesa
                                    .filter(carta => carta.jogador !== nome)
                                    .slice(0, 3);
                                return Array.from({ length: 3 }).map((_, index) => {
                                    const carta = cartasAdversario[index];
                                    return (
                                        <div
                                            key={`adv-slot-${index}`}
                                            className={`w-20 h-32 rounded-lg shadow-lg border border-gray-400 ${
                                                carta ? 'bg-white' : 'bg-gray-200 opacity-50'
                                            }`}
                                            style={{ 
                                                backgroundImage: carta?.cartaJogada ? 
                                                    `url(/${carta.cartaJogada.naipe.toLowerCase()}/${carta.cartaJogada.valor}-${carta.cartaJogada.naipe.toLowerCase()}.png)` : 'none',
                                                backgroundSize: 'cover',
                                                backgroundPosition: 'center',
                                                backgroundRepeat: 'no-repeat'
                                            }}
                                        />
                                    );
                                });
                            })()}
                        </div>
                        <div className="flex justify-start space-x-24" style={{ width: '432px' }}>
                            {(() => {
                                const cartasJogador = cartasMesa
                                    .filter(carta => carta.jogador === nome)
                                    .slice(0, 3);
                                return Array.from({ length: 3 }).map((_, index) => {
                                    const carta = cartasJogador[index];
                                    return (
                                        <div
                                            key={`jog-slot-${index}`}
                                            className={`w-20 h-32 rounded-lg shadow-lg border ${
                                                carta ? 'bg-white border-blue-400' : 'bg-gray-200 border-gray-400 opacity-50'
                                            }`}
                                            style={{ 
                                                backgroundImage: carta?.cartaJogada ? 
                                                    `url(/${carta.cartaJogada.naipe.toLowerCase()}/${carta.cartaJogada.valor}-${carta.cartaJogada.naipe.toLowerCase()}.png)` : 'none',
                                                backgroundSize: 'cover',
                                                backgroundPosition: 'center',
                                                backgroundRepeat: 'no-repeat'
                                            }}
                                        />
                                    );
                                });
                            })()}
                        </div>
                    </div>
                </div>
                <div className="flex justify-center space-x-4 mb-8">
                    {Array.from({ length: 3 }).map((_, idx) => {
                        const carta = jogador.mao[idx];
                        return carta ? (
                            <div
                                key={`${carta.naipe}-${carta.valor}`}
                                onClick={() => mostrarSuaVez && jogarCarta(carta)}
                                className={`w-36 h-56 rounded-xl shadow-lg transition-transform duration-200 mt-12 bg-white ${mostrarSuaVez ? 'cursor-pointer hover:scale-105 hover:shadow-2xl' : 'opacity-50 cursor-not-allowed'}`}
                                style={{
                                    backgroundImage: `url(/${carta.naipe.toLowerCase()}/${carta.valor}-${carta.naipe.toLowerCase()}.png)`,
                                    backgroundSize: '90%',
                                    backgroundPosition: 'center',
                                    backgroundRepeat: 'no-repeat'
                                }}
                            />
                        ) : (
                            <div
                                key={idx}
                                className="w-36 h-56 rounded-xl shadow-lg mt-12"
                                style={{ backgroundColor: 'transparent' }}
                            />
                        );
                    })}
                </div>
            </div>
            {showEnvidoOptions && rodadaAtual === 1 && (
                <div className="absolute top-[46%] right-32 z-30 flex flex-col space-y-4 envido-options" style={{ transform: 'translateY(-50%)' }}>
                    {apostasDisponiveis.Envido && (
                        <button
                            className="bg-yellow-950/60 text-gray-300 hover:bg-yellow-900/80 px-8 py-4 text-xl font-bold border-2 border-amber-600 rounded shadow-lg transition-colors"
                            onClick={() => { pedirAposta('CHAMAR_ENVIDO'); setShowEnvidoOptions(false); }}
                        >Envido</button>
                    )}
                    {apostasDisponiveis.RealEnvido && (
                        <button
                            className="bg-yellow-950/60 text-gray-300 hover:bg-yellow-900/80 px-8 py-4 text-xl font-bold border-2 border-amber-600 rounded shadow-lg transition-colors"
                            onClick={() => { pedirAposta('CHAMAR_REAL_ENVIDO'); setShowEnvidoOptions(false); }}
                        >Real Envido</button>
                    )}
                    {apostasDisponiveis.FaltaEnvido && (
                        <button
                            className="bg-yellow-950/60 text-gray-300 hover:bg-yellow-900/80 px-8 py-4 text-xl font-bold border-2 border-amber-600 rounded shadow-lg transition-colors"
                            onClick={() => { pedirAposta('CHAMAR_FALTA_ENVIDO'); setShowEnvidoOptions(false); }}
                        >Falta Envido</button>
                    )}
                </div>
            )}
            <div className="absolute top-1/2 right-80 transform -translate-y-1/2 z-10">
                <div className="flex flex-col space-y-4">
                    <button
                        className={`bg-yellow-950/60 text-gray-300 px-8 py-4 text-xl font-bold border-2 border-amber-600 rounded shadow-lg transition-colors ${obterProximoTruco() && mostrarSuaVez ? 'hover:bg-yellow-900/80' : 'opacity-50 cursor-not-allowed'}`}
                        onClick={() => {
                            const proximoTruco = obterProximoTruco();
                            if (proximoTruco && mostrarSuaVez) {
                                pedirAposta(proximoTruco);
                            }
                        }}
                        disabled={!obterProximoTruco() || !mostrarSuaVez}
                    >{obterTextoTruco()}</button>
                    <button
                        className={`bg-yellow-950/60 text-gray-300 px-8 py-4 text-xl font-bold border-2 border-amber-600 rounded shadow-lg transition-colors ${apostasDisponiveis.Envido && mostrarSuaVez && rodadaAtual === 1 ? 'hover:bg-yellow-900/80 envido-trigger' : 'opacity-50 cursor-not-allowed'}`}
                        onClick={() => {
                            if (apostasDisponiveis.Envido && mostrarSuaVez && rodadaAtual === 1) {
                                setShowEnvidoOptions(true);
                            }
                        }}
                        disabled={!apostasDisponiveis.Envido || !mostrarSuaVez || rodadaAtual !== 1}
                    >ENVIDO</button>
                    <button
                        className={`bg-yellow-950/60 text-gray-300 px-8 py-4 text-xl font-bold border-2 border-amber-600 rounded shadow-lg transition-colors ${temFlor() && mostrarSuaVez && rodadaAtual === 1 ? 'hover:bg-yellow-900/80' : 'opacity-50 cursor-not-allowed'}`}
                        onClick={() => {
                            if (temFlor() && mostrarSuaVez && rodadaAtual === 1) {
                                pedirAposta('CHAMAR_FLOR');
                            }
                        }}
                        disabled={!temFlor() || !mostrarSuaVez || rodadaAtual !== 1}
                    >FLOR</button>
                    <button
                        className={`bg-yellow-950/60 text-gray-300 px-8 py-4 text-xl font-bold border-2 border-amber-600 rounded shadow-lg transition-colors ${mostrarSuaVez ? 'hover:bg-yellow-900/80' : 'opacity-50 cursor-not-allowed'}`}
                        onClick={() => {
                            if (mostrarSuaVez) {
                                pedirAposta('IR_AO_MAZO');
                            }
                        }}
                        disabled={!mostrarSuaVez}
                    >MAZO</button>
                </div>
            </div>
            {apostaDialog && apostaDialog.mensagemFinal && (
                <div className="fixed inset-0 z-50 flex items-center justify-center" style={{ backgroundColor: 'rgba(0, 0, 0, 0.7)', backdropFilter: 'blur(4px)' }}>
                    <div
                        className={`border-2 rounded-xl shadow-2xl px-10 py-8 flex flex-col items-center w-[340px] ${
                            apostaDialog.mensagemFinal === 'VOCE_GANHOU' ? 'bg-green-700 border-green-500' : 'bg-red-700 border-red-500'
                        }`}
                    >
                        <h2 className="text-2xl font-bold text-white mb-4">
                            {apostaDialog.mensagemFinal === 'VOCE_GANHOU' ? '🎉 Vitória!' : '😢 Derrota!'}
                        </h2>
                        <span className="text-lg text-gray-200 mb-6 text-center">
                            Placar Final: Time 1 - {apostaDialog.placarFinal?.TIME_01 || 0} | Time 2 - {apostaDialog.placarFinal?.TIME_02 || 0}
                        </span>
                        <button
                            className="bg-gray-800 text-white font-bold py-2 px-4 rounded shadow hover:bg-gray-900 transition-colors"
                            onClick={() => setApostaDialog(null)}
                        >
                            Fechar
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
}

export default Jogo;
