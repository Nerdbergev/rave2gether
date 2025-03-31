// src/services/queueService.ts
import api from "../api";
import { QueueResponse, Song } from "../types";

export const fetchAllQueues = async (): Promise<QueueResponse> => {
  const response = await api.get("/queue/all");
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