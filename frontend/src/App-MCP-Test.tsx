import React from 'react';
import { Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { CanvasProvider } from '@workday/canvas-kit-react/common';
import { SimpleTestPage } from './SimpleTestPage';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      refetchOnWindowFocus: false,
    },
  },
});

function App() {
  return (
    <Routes>
      <Route path="/" element={<SimpleTestPage />} />
    </Routes>
  );
}

export default App;