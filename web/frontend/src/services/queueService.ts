// src/services/queueService.ts
import api from "../api";
import { Song } from "../types";

export const fetchQueue = async (): Promise<Song[]> => {
  const response = await api.get("/queue");
  return response.data;
};

export const fetchDownloadQueue = async (): Promise<Song[]> => {
  const response = await api.get("/queue/download");
  return response.data;
}

export const addToQueue = async (song: string): Promise<void> => {
  await api.post("/queue", { "queries": [song] });
};

export const skipSong = async (): Promise<void> => {
  await api.post("/queue/skip");
};

export const deleteSong = async (songId: string): Promise<void> => {
  await api.delete(`/queue/${songId}`);
};

export const fetchCurrentSong = async (): Promise<Song | null> => {
  const response = await api.get("/queue/current");
  return response.data;
};

export const upvoteSong = async (id: string): Promise<void> => {
  await api.post(`/queue/${id}/vote`, { upvote: true });
};

export const downvoteSong = async (id: string): Promise<void> => {
  await api.post(`/queue/${id}/vote`, { upvote: false });
};