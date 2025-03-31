// src/components/Queue.tsx
import React, { useState, useEffect } from "react";
import { fetchAllQueues, addToQueue, skipSong, deleteSong, fetchCurrentSong, upvoteSong, downvoteSong } from "../services/queueService";
import { QueueResponse, Song, UserRight } from "../types";
import QueueItem from "./QueueItem";
import QueueItemType from "./QueueItemType";
import { Mode } from "../types";
import { useAuth } from "../contexts/AuthContext";

interface QueueProps {
  mode: Mode;
}

const Queue: React.FC<QueueProps> = ({mode}) => {
  const [queues, setQueues] = useState<QueueResponse| null>(null);
  const [currentSong, setCurrentSong] = useState<Song | null>(null);
  const [error, setError] = useState<string | null>(null);
  const { user } = useAuth();
  let userIsModerator = false
  if (user) {
    userIsModerator = (user.right === UserRight.ADMIN || user.right === UserRight.MODERATOR);
  }

  // Function to load the queue from the backend
  const loadQueues = async () => {
    try {
      const data = await fetchAllQueues();
      // If the API returns null, default to an empty array.
      setQueues(data || null);
    } catch (error) {
      console.error("Failed to load queue", error);
      setError("Failed to load queue: " + error); // Set an error message
      setQueues(null); // Reset to empty array on error.
    }
  };

  const loadCurrentSong = async () => {
    try {
      const data = await fetchCurrentSong();
      setCurrentSong(data);
    } catch (error) {
      console.error("Failed to load current song", error);
      setError("Failed to load current song: " + error);
      setCurrentSong(null);
    }
};

  useEffect(() => {
    loadQueues(); // Initial load
    const intervalId = setInterval(() => {
      loadQueues();
    }, 5000);
    return () => clearInterval(intervalId);
  }, []);

  useEffect(() => {
    loadCurrentSong();
    const songInterval = setInterval(() => {
      loadCurrentSong();
    }, 1000);
    return () => clearInterval(songInterval);
  }, []);

  const { loadUserInfo } = useAuth();

  const handleAddSong = async () => {
    const song = prompt("Enter song name:");
    if (song) {
        try {
            await addToQueue(song);
            loadQueues();
            loadUserInfo();
        }
        catch (error) {
            console.error("Failed to add song", error);
            setError("Failed to add song: " + error);
        }
    }
  };

  const handleSkipSong = async () => {
    try {
      await skipSong();
      loadQueues();
    } catch (error) {
        console.error("Failed to skip song", error);
        setError("Failed to skip song: " + error);
    }
  };

  const handleDeleteSong = async (songId: string) => {
    try {
      await deleteSong(songId);
      loadQueues();
    } catch (error) {
        console.error("Failed to delete song", error);
        setError("Failed to delete song: " + error);
    }
  };


  const handleUpvote = async (id: string) => {
    try {
      await upvoteSong(id);
      loadQueues();
      loadUserInfo();
    } catch (error) {
        console.error("Failed to upvote song", error);
        setError("Failed to upvote song: " + error);
    }
  };

  const handleDownvote = async (id: string) => {
    try {
      await downvoteSong(id);
      loadQueues();
      loadUserInfo();
    } catch (error) {
        console.error("Failed to downvote song", error);
        setError("Failed to downvote song: " + error);
    }
  };

  return (
    <div className="mt-8 p-6 max-w-lg mx-auto bg-gray-800 shadow-md rounded-lg">
          {/* Error Message */}
            {error && (
            <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative mb-4">
                <span>{error}</span>
                {/* A close button to clear the error */}
                <span
                className="absolute top-0 bottom-0 right-0 px-4 py-3 cursor-pointer"
                onClick={() => setError(null)}
                >
                X
                </span>
            </div>
            )}

      {/* Controls (always visible) */}
      <div className="flex mb-4">
        <button
          onClick={handleAddSong}
          className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded mr-2"
        >
          Add Song
        </button>
      </div>

      <h3 className="text-l font-bold text-white mb-4">Current Song</h3>
      <div className="mt-4 space-y-2 min-h-[50px]">
      {(currentSong && currentSong.name != "" )? (
            <div>
                <QueueItem
                    song={currentSong}
                    itemType={QueueItemType.PLAY}
                    onUpvote={() => {}}
                    onDownvote={() => {}}
                    onDelete={() => {}}
                    onSkip={handleSkipSong}
                    mode = {mode}
                    userIsModerator = {userIsModerator}
                />
            </div>
        ) : (
            <p className="text-center text-gray-400">No current song</p>
        )}
      </div>


      {/* List container with a minimum height to avoid layout collapse */}
      <h3 className="text-l font-bold text-white mb-4">Play Queue</h3>
      <ul className="mt-4 space-y-2 min-h-[50px]">
        {queues?.playqueue.length === 0 ? (
          <li className="text-center text-gray-400">Queue is empty.</li>
        ) : (
          queues?.playqueue.map((song) => (
            <QueueItem
              key={song.id}
              song={song}
              itemType={QueueItemType.QUEUE}
              onUpvote={handleUpvote}
              onDownvote={handleDownvote}
              onDelete={handleDeleteSong}
              onSkip={() => {}}
              mode = {mode}
              userIsModerator = {userIsModerator}
            />
          ))
        )}
      </ul>

      <h3 className="text-l font-bold text-white mb-4">Download Queue</h3>
      <ul className="mt-4 space-y-2 min-h-[50px]">
        {queues?.downloadqueue.length === 0 ? (
          <li className="text-center text-gray-400">Queue is empty.</li>
        ) : (
            queues?.downloadqueue.map((song) => (
            <QueueItem
              key={song.id}
              song={song}
              itemType={QueueItemType.DOWNLOAD}
              onUpvote={() => {}}
              onDownvote={() => {}}
              onDelete={() => {}}
              onSkip={() => {}}
              mode = {mode}
              userIsModerator = {userIsModerator}
            />
          ))
        )}
      </ul>

      <h3 className="text-l font-bold text-white mb-4">Prepare Queue</h3>
      <ul className="mt-4 space-y-2 min-h-[50px]">
        {queues?.preparequeue.length === 0 ? (
          <li className="text-center text-gray-400">Queue is empty.</li>
        ) : (
            queues?.preparequeue.map((song) => (
            <QueueItem
              key={song.id}
              song={song}
              itemType={QueueItemType.PREPARE}
              onUpvote={() => {}}
              onDownvote={() => {}}
              onDelete={() => {}}
              onSkip={() => {}}
              mode = {mode}
              userIsModerator = {userIsModerator}
            />
          ))
        )}
      </ul>
    </div>
  );
};

export default Queue;
