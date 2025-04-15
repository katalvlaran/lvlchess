/*
  index.js:
  - The entry point for the React application in the "frontend" folder.
  - ReactDOM.createRoot attaches the <App/> component to "root" in index.html.
*/

import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';

// Grab the <div id="root"></div> from public/index.html
const root = ReactDOM.createRoot(document.getElementById('root'));

// Render <App/> inside that root,
// wrapped in <React.StrictMode> for highlighting potential issues.
root.render(
    <React.StrictMode>
        <App />
    </React.StrictMode>
);
