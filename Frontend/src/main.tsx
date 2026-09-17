import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import "@radix-ui/themes/styles.css";
import './index.css'

import App from './App.tsx'
import { BrowserRouter } from 'react-router';

// Mount into the root element supplied by index.html. Load Radix styles before
// application styles so the latter can customize the palette's presentation.
// StrictMode checks effect cleanup in development; BrowserRouter supplies route
// hooks to App and all pages. Theme and Kbar providers are owned inside App.
createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </StrictMode>,
)
