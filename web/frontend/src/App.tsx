// src/App.tsx
import React, { useEffect, useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router-dom";
import Header from './components/Header';
import Queue from "./components/Queue";
import Login from "./components/Login";
import Logout from "./components/Logout";
import { AppConfig } from "./types";
import api from "./api";

const App: React.FC = () => {
  const [config, setConfig] = useState<AppConfig | null>(null);
  const [loading, setLoading] = useState<boolean>(true);

  useEffect(() => {
    // Adjust the endpoint as needed (for example, /api/config or /api/self)
    const response = api.get("/mode");
    response.then((response) => {
      setConfig(response.data);
      setLoading(false);
    }).catch((error) => {
      console.error("Failed to load configuration", error);
      setLoading(false);
    });
  }, []);

  if (loading) {
    return <div>Loading configuration...</div>;
  }

  if (!config) {
    return <div>Error loading configuration.</div>;
  }

  return (
    <Router>
      <Header mode={config.mode} />
      <Routes>
        <Route path="/queue" element={<Queue mode={config.mode}/> } />
        <Route path="/login" element={<Login />} />
        <Route path="/logout" element={<Logout />} />
        <Route path="*" element={<Navigate to="/queue" />} />
      </Routes>
    </Router>
  );
};

export default App;
