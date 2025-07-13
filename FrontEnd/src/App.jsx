import Logo from './components/logo.jsx';
import Devs from './components/desenvolvedores.jsx';
import Salas from './components/salas.jsx';

function App() {
  return (
    <>
      <div className=' flex justify-between'>
        <Logo />
        <Salas />
      </div>
      <Devs />
    </>
  );
}

export default App;