import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import Logo from './components/logo.jsx';
import Devs from './components/desenvolvedores.jsx';
import Salas from './components/salas.jsx';
import Jogo from './components/jogo.jsx';

function App() {
  return (
    <Router>
      <Routes>
        <Route
          path="/"
          element={(
            <div className='flex justify-between'>
              <Logo />
              <Salas />
              <Devs />
            </div>
          )}
        />
        <Route path="/jogo/:salaId" element={<Jogo />} />
      </Routes>
    </Router>
  );
}

export default App;