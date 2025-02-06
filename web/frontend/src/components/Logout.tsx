// src/components/Logout.tsx
import React from "react";
import { logout } from "../services/authService";
import { useNavigate } from "react-router-dom";

const Logout: React.FC = () => {
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate("/login");
  };

  return (
    <button onClick={handleLogout} className="text-blue-500 hover:underline">
      Logout
    </button>
  );
};

export default Logout;