/*
  App.js:
  - Main functional component for your React application
  - Renders a simple "Hello from lvlChess React" message
  - Typically, you'll expand this with your chess board UI,
    or route it to more pages, etc.
*/

import React from 'react';

function App() {
    return (
        <div style={{ margin: '20px', fontFamily: 'sans-serif' }}>
            <h1>Hello from lvlChess React</h1>
            <p>This is a minimal example!</p>
            {/*
         Possible expansions:
           - fetch data from your Go backend
           - show a chessboard (e.g., with a library or custom code)
           - integrate Telegram WebApp if needed
      */}
        </div>
    );
}

export default App;
