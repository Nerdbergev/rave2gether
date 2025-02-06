// src/services/authService.ts
import api from "../api";

export const login = async (username: string, password: string): Promise<string> => {
  const response = await api.post("/token", {}, {
    auth: {
      username,
      password,
    },
  });
  const token = response.data.token;
  localStorage.setItem("token", token);
  return token;
};

export const logout = (): void => {
    localStorage.removeItem("token");
    // You can add further cleanup or redirection here
  };
  
  export const refreshToken = async (): Promise<string> => {
    // Make an API call to refresh the token.
    // This endpoint should validate your refresh token (stored securely) and return a new access token.
    const response = await api.post("/token/refresh");
    const newToken = response.data.token;
    return newToken;
  };

export const isAuthenticated = (): boolean => !!localStorage.getItem("token");
