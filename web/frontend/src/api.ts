// src/api.ts
import axios from "axios";
import { jwtDecode } from "jwt-decode";
import { logout, refreshToken } from "./services/authService";

const API_URL = "/api";

const getToken = (): string | null => localStorage.getItem("token");

// Helper to check if token is expired using its 'exp' claim.
const isTokenExpired = (token: string): boolean => {
  try {
    const decoded: { exp: number } = jwtDecode(token);
    // 'exp' is in seconds, so compare with current time in seconds.
    return decoded.exp < Date.now() / 1000;
  } catch (error) {
    console.error("Failed to decode token", error);
    return true;
  }
};

const api = axios.create({
  baseURL: API_URL,
});

api.interceptors.request.use(
  async (config) => {
    // Optionally, skip token logic for endpoints that should be public.
    // For example, if the URL includes '/token' (for login or refresh), skip:
    if (config.url && (config.url.includes("/token") || config.url.includes("/public"))) {
      return config;
    }

    const token = getToken();
    // If there is no token, just return the config (e.g., during initial login).
    if (!token) {
      return config;
    }

    // If the token exists, check if it’s expired.
    if (isTokenExpired(token)) {
      try {
        // Try to refresh the token.
        const newToken = await refreshToken(); // Implement refreshToken in your auth service.
        localStorage.setItem("token", newToken);
        config.headers.Authorization = `Bearer ${newToken}`;
      } catch (error) {
        // If refreshing fails, log out and redirect.
        console.error("Token expired and refresh failed", error);
        logout();
        window.location.href = "/login";
        return Promise.reject("Token expired and refresh failed");
      }
    } else {
      // Token is valid; attach it.
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

export default api;
