// src/App.tsx
import React from "react";
import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router-dom";
import Queue from "./components/Queue";
import Login from "./components/Login";
import Logout from "./components/Logout";
import { isAuthenticated } from "./services/authService";

const App: React.FC = () => {
  return (
    <Router>
      <Routes>
        <Route path="/queue" element={isAuthenticated() ? <Queue /> : <Navigate to="/login" />} />
        <Route path="/login" element={<Login />} />
        <Route path="/logout" element={<Logout />} />
        <Route path="*" element={<Navigate to="/queue" />} />
      </Routes>
    </Router>
  );
};

export default App;
