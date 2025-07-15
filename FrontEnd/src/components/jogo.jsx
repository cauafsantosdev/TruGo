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
    const [mensagem, setMensagem] = useState('');
    const [conexaoStatus, setConexaoStatus] = useState('Desconectado');
    const wsRef = useRef(null);
    const reconnectTimeoutRef = useRef(null);
    const reconnectAttemptsRef = useRef(0);
    const maxReconnectAttempts = 10;

    const sincronizarCartasMesa = useCallback((cartasJogadas, forcarAtualizacao = false) => {
        if (!cartasJogadas || cartasJogadas.length === 0) {
            if (forcarAtualizacao) {
                setCartasMesa([]);
            }
            return;
        }
        const cartasNormalizadas = cartasJogadas
            .filter(carta => {
                const temCarta = carta && (carta.cartaJogada || carta.carta);
                const temJogador = carta && carta.jogador;
                return temCarta && temJogador;
            })
            .map((carta, index) => {
                const cartaJogada = carta.cartaJogada || carta.carta;
                return {
                    jogador: carta.jogador,
                    cartaJogada: cartaJogada,
                    rodada: carta.rodada || 1,
                    timestamp: carta.timestamp || Date.now(),
                    id: `${carta.jogador}-${cartaJogada?.naipe}-${cartaJogada?.valor}-${index}`
                };
            });
        setCartasMesa(cartasNormalizadas);
        if (!forcarAtualizacao) {
            setJogador(j => {
                const minhasJogadas = cartasNormalizadas
                    .filter(p => p && p.jogador === nome && p.cartaJogada)
                    .map(p => p.cartaJogada);
                if (minhasJogadas.length === 0) return j;
                const maoAtualizada = j.mao.filter(c =>
                    !minhasJogadas.some(mj =>
                        mj && mj.naipe === c.naipe && mj.valor === c.valor
                    )
                );
                return { ...j, mao: maoAtualizada };
            });
        }
    }, [nome]);

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
                    const data = JSON.parse(event.data);
                    if (reconnectTimeoutRef.current) {
                        clearTimeout(reconnectTimeoutRef.current);
                    }
                    switch (data.type) {
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
                            setJogador(j => ({ ...j, mao: data.mao }));
                            if (cartasMesa.length > 0 && data.mao && data.mao.length > 0) {
                                setCartasMesa([]);
                            }
                            if (data.limparMesa || data.novaPartida) {
                                setCartasMesa([]);
                            }
                            break;
                        case 'SUA_VEZ':
                            setMostrarSuaVez(true);
                            setApostasDisponiveis(data.apostasDisponiveis || {});
                            setSala(s => ({ ...s, placar: data.placar || s.placar }));
                            const cartasParaSincronizar = data.cartasJogadas || data.cartasJogadasMesa || data.cartas || [];
                            if (cartasParaSincronizar.length > 0) {
                                sincronizarCartasMesa(cartasParaSincronizar);
                            }
                            break;
                        case 'STATUS_PARTIDA':
                            setApostasDisponiveis(data.apostasDisponiveis || {});
                            const cartasStatus = data.cartasJogadas || [];
                            sincronizarCartasMesa(cartasStatus, true);
                            setSala(s => ({ ...s, placar: data.placar || s.placar }));
                            break;
                        case 'CARTA_JOGADA':
                            if (data.cartaJogada && data.jogador && data.jogador !== nome) {
                                setCartasMesa(prev => {
                                    const cartaJaExiste = prev.find(c => 
                                        c.jogador === data.jogador && 
                                        c.cartaJogada?.naipe === data.cartaJogada.naipe &&
                                        c.cartaJogada?.valor === data.cartaJogada.valor
                                    );
                                    if (cartaJaExiste) {
                                        return prev;
                                    }
                                    const novaCarta = {
                                        jogador: data.jogador,
                                        cartaJogada: data.cartaJogada,
                                        rodada: data.rodada || 1,
                                        timestamp: Date.now(),
                                        id: `${data.jogador}-${data.cartaJogada.naipe}-${data.cartaJogada.valor}-${Date.now()}`
                                    };
                                    return [...prev, novaCarta];
                                });
                                setMensagem(`${data.jogador} jogou ${data.cartaJogada.valor} de ${data.cartaJogada.naipe}`);
                            }
                            break;
                        case 'APOSTA':
                            setMensagem(`Aposta: ${data.tipoAposta}`);
                            break;
                        case 'RESPOSTA_APOSTA':
                            setMensagem(`Aposta ${data.tipoAposta} ${data.aceito ? 'aceita' : 'recusada'}`);
                            break;
                        case 'RODADA_FINALIZADA':
                            setMensagem("Rodada finalizada!");
                            setCartasMesa([]);
                            break;
                        case 'NOVA_RODADA':
                            setCartasMesa([]);
                            setMensagem("Nova rodada iniciada!");
                            if (data.cartasJogadas && data.cartasJogadas.length > 0) {
                                sincronizarCartasMesa(data.cartasJogadas);
                            }
                            break;
                        case 'VENCEDOR_MAO':
                            setMensagem(`${data.vencedor} venceu a mão!`);
                            setCartasMesa([]);
                            setMostrarSuaVez(false);
                            break;
                        case 'LIMPAR_MESA':
                            setCartasMesa([]);
                            break;
                        case 'error':
                            setMensagem(data.msg);
                            break;
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

    const enviarEntrarEquipe = () => {
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
            setMensagem('Payload ENTRAR_EQUIPE enviado!');
        } else {
            setMensagem('Conexão não disponível. Tentando reconectar...');
            reconectarManualmente();
        }
    };

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
        setJogador(j => ({
            ...j,
            mao: j.mao.filter(c => !(c.naipe === carta.naipe && c.valor === carta.valor))
        }));
        const novaCartaNaMesa = {
            jogador: nome,
            cartaJogada: carta,
            rodada: 1,
            timestamp: Date.now(),
            id: `${nome}-${carta.naipe}-${carta.valor}-${Date.now()}`
        };
        setCartasMesa(prev => [...prev, novaCartaNaMesa]);
        setMostrarSuaVez(false);
        setMensagem(`Você jogou ${carta.valor} de ${carta.naipe}`);
        enviarAcao('FAZER_JOGADA', carta);
    };
    const pedirEnvido = (tipo) => {
        const tiposEnvido = {
            'Envido': 'CHAMAR_ENVIDO',
            'Real Envido': 'CHAMAR_REAL_ENVIDO',
            'Falta Envido': 'CHAMAR_FALTA_ENVIDO'
        };
        enviarAcao(tiposEnvido[tipo]);
    };
    const pedirTruco = () => enviarAcao('CHAMAR_TRUCO');
    const aceitarAposta = (tipoAposta) => enviarAcao('ACEITAR_APOSTA', { tipoAposta, aceitar: true });
    const recusarAposta = (tipoAposta) => enviarAcao('ACEITAR_APOSTA', { tipoAposta, aceitar: false });

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
            <button
                onClick={enviarEntrarEquipe}
                className="absolute top-4 left-4 z-50 bg-green-700 hover:bg-green-800 text-white px-6 py-2 rounded font-bold shadow-lg"
            >
                Entrar
            </button>
            <img 
                src="/Trugo_logo-removebg-preview(2).png" 
                alt="Logo TruGo" 
                className="absolute top-4 right-4 w-44 h-auto z-30 drop-shadow-lg select-none"
            />
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
                        const cartasJogadasAdversario = cartasMesa.filter(carta => carta.jogador !== nome).length;
                        const cartasRestantes = Math.max(0, 3 - cartasJogadasAdversario);
                        return Array.from({ length: cartasRestantes }).map((_, idx) => (
                            <div
                                key={`adv-hand-${idx}`}
                                className="w-36 h-56 rounded-lg shadow-lg bg-white"
                                style={{ 
                                    backgroundImage: `url(${cartabg})`, 
                                    backgroundSize: 'cover', 
                                    backgroundPosition: 'center' 
                                }}
                            />
                        ));
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
                    {jogador.mao.length > 0 ? (
                        jogador.mao.map((carta, idx) => {
                            const pasta = carta.naipe.toLowerCase();
                            const valor = carta.valor;
                            const imgPath = `/${pasta}/${valor}-${pasta}.png`;
                            const podeJogar = mostrarSuaVez;
                            return (
                                <div
                                    key={`${carta.naipe}-${carta.valor}`}
                                    onClick={() => podeJogar && jogarCarta(carta)}
                                    className={`w-36 h-56 rounded-lg shadow-lg transition-transform duration-200 mt-12 ${podeJogar ? 'cursor-pointer hover:scale-105 hover:shadow-2xl' : 'opacity-50 cursor-not-allowed'}`}
                                    style={{
                                        backgroundImage: `url(${imgPath})`,
                                        backgroundSize: '95% 95%',
                                        backgroundPosition: 'center',
                                        backgroundRepeat: 'no-repeat'
                                    }}
                                />
                            );
                        })
                    ) : (
                        Array.from({ length: 3 }).map((_, idx) => (
                            <div
                                key={idx}
                                className="w-36 h-56 rounded-lg shadow-lg flex items-center justify-center bg-white mt-12"
                                style={{
                                    backgroundImage: `url(${cartabg})`,
                                    backgroundSize: 'cover',
                                    backgroundPosition: 'center',
                                    backgroundRepeat: 'no-repeat',
                                    opacity: 0.8
                                }}
                            />
                        ))
                    )}
                </div>
            </div>
            {showEnvidoOptions && (
                <div className="absolute top-[46%] right-32 z-30 flex flex-col space-y-4 envido-options" style={{ transform: 'translateY(-50%)' }}>
                    {["Envido", "Real Envido", "Falta Envido"].map((option, idx) => (
                        <button
                            key={idx}
                            className="bg-yellow-950/60 text-gray-300 hover:bg-yellow-900/80 px-8 py-4 text-xl font-bold border-2 border-amber-600 rounded shadow-lg transition-colors"
                        >
                            {option}
                        </button>
                    ))}
                </div>
            )}
            <div className="absolute top-1/2 right-80 transform -translate-y-1/2 z-10">
                <div className="flex flex-col space-y-4">
                        <button
                            className={`bg-yellow-950/60 text-gray-300 px-8 py-4 text-xl font-bold border-2 border-amber-600 rounded shadow-lg transition-colors ${apostasDisponiveis['CHAMAR_TRUCO'] && mostrarSuaVez ? 'hover:bg-yellow-900/80' : 'opacity-50 cursor-not-allowed'}`}
                            onClick={apostasDisponiveis['CHAMAR_TRUCO'] && mostrarSuaVez ? pedirTruco : undefined}
                            disabled={!apostasDisponiveis['CHAMAR_TRUCO'] || !mostrarSuaVez}
                        >
                            TRUCO
                        </button>
                    <button
                        className={`bg-yellow-950/60 text-gray-300 px-8 py-4 text-xl font-bold border-2 border-amber-600 rounded shadow-lg transition-colors ${apostasDisponiveis['CHAMAR_ENVIDO'] && mostrarSuaVez ? 'hover:bg-yellow-900/80 envido-trigger' : 'opacity-50 cursor-not-allowed'}`}
                        onClick={apostasDisponiveis['CHAMAR_ENVIDO'] && mostrarSuaVez ? () => setShowEnvidoOptions(true) : undefined}
                        disabled={!apostasDisponiveis['CHAMAR_ENVIDO'] || !mostrarSuaVez}
                    >
                        ENVIDO
                    </button>
                    <button
                        className={`bg-yellow-950/60 text-gray-300 px-8 py-4 text-xl font-bold border-2 border-amber-600 rounded shadow-lg transition-colors ${apostasDisponiveis['CANTAR_FLOR'] && mostrarSuaVez ? 'hover:bg-yellow-900/80' : 'opacity-50 cursor-not-allowed'}`}
                        onClick={apostasDisponiveis['CANTAR_FLOR'] && mostrarSuaVez ? () => enviarAcao('CANTAR_FLOR') : undefined}
                        disabled={!apostasDisponiveis['CANTAR_FLOR'] || !mostrarSuaVez}
                    >
                        FLOR
                    </button>
                    <button
                        className={`bg-yellow-950/60 text-gray-300 px-8 py-4 text-xl font-bold border-2 border-amber-600 rounded shadow-lg transition-colors ${apostasDisponiveis['MAZO'] && mostrarSuaVez ? 'hover:bg-yellow-900/80' : 'opacity-50 cursor-not-allowed'}`}
                        onClick={apostasDisponiveis['MAZO'] && mostrarSuaVez ? () => enviarAcao('MAZO') : undefined}
                        disabled={!apostasDisponiveis['MAZO'] || !mostrarSuaVez}
                    >
                        MAZO
                    </button>
                </div>
            </div>
        </div>
    );
}

export default Jogo;
