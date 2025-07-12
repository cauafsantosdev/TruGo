import './index.css';

function App() {
  return (
    // Um container que centraliza tudo na tela com fundo escuro
    <div className="flex h-screen items-center justify-center bg-gray-900">
      
      {/* Um título grande, branco, em negrito e com um sublinhado azul */}
      <h1 className="text-4xl font-bold text-white underline decoration-blue-500">
        React com Tailwind é show!
      </h1>

    </div>
  );
}

export default App;