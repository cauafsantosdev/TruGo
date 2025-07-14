import React, { useEffect, useRef, useState } from "react";

function Salas() {
    const [salas, setSalas] = useState([]);
    const [novaSala, setNovaSala] = useState("");
    const ws = useRef(null);

    useEffect(() => {
        // Conecta ao WebSocket do backend
        ws.current = new WebSocket("ws://localhost:8080/ws");

        ws.current.onopen = () => {
            console.log("Conectado ao WebSocket");
            // Busca a lista de salas assim que a conexão é estabelecida
            listarSalas();
        };

        ws.current.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                // CORREÇÃO: A chave no JSON do backend é "salasDisponiveis" (camelCase)
                if (data.salasDisponiveis) {
                    setSalas(Object.entries(data.salasDisponiveis));
                } else if (data.type === "ok") {
                    // Se uma sala for criada com sucesso, atualiza a lista
                    listarSalas();
                } else if (data.type === "error") {
                    // Exibe um alerta em caso de erro
                    alert(data.msg);
                }
            } catch (e) {
                console.error("Erro ao processar mensagem:", e);
            }
        };

        ws.current.onclose = () => {
            console.log("Conexão com o WebSocket fechada.");
        };

        // Fecha a conexão ao desmontar o componente
        return () => {
            ws.current && ws.current.close();
        };
    }, []);

    // Função para enviar a requisição de listar salas
    function listarSalas() {
        if (ws.current && ws.current.readyState === WebSocket.OPEN) {
            ws.current.send(JSON.stringify({ type: "LISTAR_SALAS" }));
        }
    }

    // Função para criar uma nova sala
    function criarSala(e) {
        e.preventDefault();
        if (!novaSala.trim()) return;
        if (ws.current && ws.current.readyState === WebSocket.OPEN) {
            ws.current.send(JSON.stringify({ type: "CRIAR_SALA", id: novaSala }));
        }
        setNovaSala("");
    }

    // Função para lidar com o clique em uma sala
    function handleEntrarSala(id) {
        // Apenas exibe o alerta, sem mudar o estado
        alert(`Entrou na sala de id ${id}`);
        // Aqui você adicionaria a lógica para realmente entrar na sala,
        // por exemplo: ws.current.send(JSON.stringify({ type: "ENTRAR_SALA", idSala: id }));
    }

    return (
        <div className="flex flex-col items-center justify-center w-full h-screen">
            <div className="relative p-8 rounded-lg shadow-lg min-w-[450px] min-h-[600px] flex flex-col items-center text-white">
                {/* Pseudo-elemento para opacidade no fundo */}
                <div className="absolute inset-0 bg-yellow-950 opacity-50 rounded-lg pointer-events-none"></div>

                <h2 className="text-2xl font-bold mb-4 text-center relative text-gray-300">Salas Disponíveis</h2>
                
                {/* Formulário para criar sala */}
                <form onSubmit={criarSala} className="flex gap-2 mb-4 w-full justify-center relative text-gray-300">
                    <input
                        type="text"
                        placeholder="ID da nova sala"
                        value={novaSala}
                        onChange={e => setNovaSala(e.target.value)}
                        className="border rounded px-2 py-1 flex-1 text-gray-300"
                    />
                    <button type="submit" className="bg-[#FFD700] hover:bg-[#E5C200] text-white px-3 py-1 rounded transition-colors">Criar</button>
                </form>

                {/* Botão para atualizar a lista */}
                <button
                    onClick={listarSalas}
                    className="mb-4 bg-[#00923F] hover:bg-[#007A34] text-gray-200 px-4 py-2 rounded transition-colors relative w-full"
                >
                    Atualizar Lista
                </button>

                {/* Lista de salas */}
                <ul className="w-full max-h-100 overflow-y-auto rounded p-1 no-scrollbar relative text-gray-300">
                    {salas.length === 0 ? (
                        <li className="text-gray-300 text-center">Nenhuma sala disponível</li>
                    ) : (
                        salas.map(([id, vagas]) => (
                            <li
                                key={id}
                                onClick={() => handleEntrarSala(id)}
                                className="flex justify-between items-center border-b border-amber-600 last:border-b-0 py-2 px-2 hover:bg-yellow-900 rounded transition-colors cursor-pointer text-gray-300"
                            >
                                <span className="font-semibold text-gray-300">ID: {id}</span>
                                <span className="text-sm text-gray-300">Vagas: {vagas}</span>
                            </li>
                        ))
                    )}
                </ul>
            </div>
        </div>
    );
}

export default Salas;