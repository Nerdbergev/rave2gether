// src/App.tsx
import React from "react";
import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router-dom";
import Header from './components/Header';
import Queue from "./components/Queue";
import Login from "./components/Login";
import Logout from "./components/Logout";

const App: React.FC = () => {
  return (
    <Router>
      <Header />
      <Routes>
        <Route path="/queue" element={<Queue /> } />
        <Route path="/login" element={<Login />} />
        <Route path="/logout" element={<Logout />} />
        <Route path="*" element={<Navigate to="/queue" />} />
      </Routes>
    </Router>
  );
};

export default App;
