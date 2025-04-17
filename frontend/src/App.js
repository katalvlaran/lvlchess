/*
  App.js:
  - Main functional component for your React application
  - Renders a simple "Hello from lvlChess React" message
  - Typically, you'll expand this with your chess board UI,
    or route it to more pages, etc.
*/

import React from 'react';
import TelegramGame from './TelegramGame';

function App() {
    return <TelegramGame />;
}

export default App;
