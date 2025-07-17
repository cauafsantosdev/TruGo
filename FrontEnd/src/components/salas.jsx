import React, { useEffect, useRef, useState } from "react";
import { useParams, useNavigate } from 'react-router-dom';

function Salas({ telaAtual = "salas" }) {
    const { salaId } = useParams();
    const navigate = useNavigate();
    const [salas, setSalas] = useState([]);
    const [novaSala, setNovaSala] = useState("");
    const [showDialog, setShowDialog] = useState(false);
    const [nome, setNome] = useState("");
    const [time, setTime] = useState("");
    const [selectedSala, setSelectedSala] = useState(null);
    const ws = useRef(null);

    useEffect(() => {
        ws.current = new WebSocket("wss://trugo.onrender.com/ws");
        ws.current.onopen = listarSalas;
        ws.current.onmessage = handleWebSocketMessage;
        ws.current.onclose = () => console.log("Conexão com o WebSocket fechada.");

        return () => ws.current && ws.current.close();
    }, []);

    const handleWebSocketMessage = (event) => {
        try {
            const data = JSON.parse(event.data);
            if (data.salasDisponiveis) {
                setSalas(Object.entries(data.salasDisponiveis));
            } else if (data.type === "ok") {
                listarSalas();
            } else if (data.type === "error") {
                alert(data.msg);
            }
        } catch (e) {
            console.error("Erro ao processar mensagem:", e);
        }
    };

    const listarSalas = () => {
        if (ws.current?.readyState === WebSocket.OPEN) {
            ws.current.send(JSON.stringify({ type: "LISTAR_SALAS" }));
        }
    };

    const criarSala = (e) => {
        e.preventDefault();
        if (novaSala.trim() && ws.current?.readyState === WebSocket.OPEN) {
            ws.current.send(JSON.stringify({ type: "CRIAR_SALA", id: novaSala }));
            setNovaSala("");
        }
    };

    const handleEntrarSala = (id, vagas) => {
        if (vagas === 0) {
            alert("Não é possível entrar na sala, pois não há vagas disponíveis.");
            return;
        }
        setSelectedSala(id);
        setShowDialog(true);
    };

    const confirmarEntradaSala = () => {
        if (!nome.trim() || !time.trim()) {
            alert("Por favor, preencha todos os campos.");
            return;
        }

        const timePadrao = time === 'time1' ? 'TIME_01' : 'TIME_02';
        localStorage.setItem('nome', nome);
        localStorage.setItem('time', timePadrao);
        localStorage.setItem('salaId', selectedSala);
        localStorage.setItem('entryData', JSON.stringify({ nome, time: timePadrao, salaId: selectedSala }));
        setShowDialog(false);
        setNome("");
        setTime("");
        navigate(`/jogo/${selectedSala}`);
    };

    const renderSalas = () => (
        <ul className="w-full max-h-100 overflow-y-auto rounded p-1 no-scrollbar relative text-gray-300">
            {salas.length === 0 ? (
                <li className="text-gray-300 text-center">Nenhuma sala disponível</li>
            ) : (
                salas.map(([id, vagas]) => (
                    <li
                        key={id}
                        onClick={() => handleEntrarSala(id, vagas)}
                        className="flex justify-between items-center border-b border-amber-600 last:border-b-0 py-2 px-2 hover:bg-yellow-900 rounded transition-colors cursor-pointer text-gray-300"
                    >
                        <span className="font-semibold text-gray-300">ID: {id}</span>
                        <span className="text-sm text-gray-300">Vagas: {vagas}</span>
                    </li>
                ))
            )}
        </ul>
    );

    const renderDialog = () => (
        <div className="fixed inset-0 flex items-center justify-center backdrop-blur-sm">
            <div className="bg-yellow-950 bg-opacity-0 p-6 rounded-lg shadow-lg text-gray-300 w-96">
                <h3 className="text-xl font-bold mb-4">Entrar na sala {selectedSala}</h3>
                <form className="flex flex-col gap-4">
                    <input
                        type="text"
                        placeholder="Seu nome"
                        value={nome}
                        onChange={(e) => setNome(e.target.value)}
                        className="border rounded px-2 py-1 text-gray-300"
                    />
                    <select
                        value={time}
                        onChange={(e) => setTime(e.target.value)}
                        className="border rounded px-2 py-1 text-gray-300"
                    >
                        <option value="" disabled className="text-black">Selecione seu time</option>
                        <option value="time1" className="text-black">Time 1</option>
                        <option value="time2" className="text-black">Time 2</option>
                    </select>
                    <div className="flex gap-4">
                        <button
                            type="button"
                            onClick={() => setShowDialog(false)}
                            className="bg-red-600 hover:bg-red-500 text-white px-4 py-2 rounded transition-colors flex-1"
                        >
                            Fechar
                        </button>
                        <button
                            type="button"
                            onClick={confirmarEntradaSala}
                            className="bg-[#00923F] hover:bg-[#007A34] text-white px-4 py-2 rounded transition-colors flex-1"
                        >
                            Confirmar
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );

    return (
        telaAtual === "salas" ? (
            <div className="flex flex-col items-center justify-center w-full h-screen">
                <div className="relative p-8 rounded-lg shadow-lg min-w-[450px] min-h-[600px] flex flex-col items-center text-white">
                    <div className="absolute inset-0 bg-yellow-950 opacity-50 rounded-lg pointer-events-none"></div>
                    <h2 className="text-2xl font-bold mb-4 text-center relative text-gray-300">Salas Disponíveis</h2>
                    <form onSubmit={criarSala} className="flex gap-2 mb-4 w-full justify-center relative text-gray-300">
                        <input
                            type="text"
                            placeholder="ID da nova sala"
                            value={novaSala}
                            onChange={(e) => setNovaSala(e.target.value)}
                            className="border rounded px-2 py-1 flex-1 text-gray-300"
                        />
                        <button type="submit" className="bg-[#FFD700] hover:bg-[#E5C200] text-white px-3 py-1 rounded transition-colors">Criar</button>
                    </form>
                    <button
                        onClick={listarSalas}
                        className="mb-4 bg-[#00923F] hover:bg-[#007A34] text-gray-200 px-4 py-2 rounded transition-colors relative w-full"
                    >
                        Atualizar Lista
                    </button>
                    {renderSalas()}
                </div>
                {showDialog && renderDialog()}
            </div>
        ) : null
    );
}

export default Salas;