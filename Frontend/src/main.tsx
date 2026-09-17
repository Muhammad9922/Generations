import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import "@radix-ui/themes/styles.css";
import './index.css'
import { Theme } from "@radix-ui/themes";
import App from './App.tsx'
import { BrowserRouter } from 'react-router';
import { KBarProvider } from 'kbar';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Theme>
      <BrowserRouter>
        <KBarProvider>
          <App />
        </KBarProvider>
      </BrowserRouter>
    </Theme>
  </StrictMode>,
)
