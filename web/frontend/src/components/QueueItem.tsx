// src/components/QueueItem.tsx
import React from "react";
import { Mode, Song } from "../types";
import QueueItemType from "./QueueItemType";

interface QueueItemProps {
  song: Song;
  itemType: QueueItemType;
  onUpvote: (id: string) => void;
  onDownvote: (id: string) => void;
  onDelete: (id: string) => void;
  onSkip: (id: string) => void;
  mode: Mode;
  userIsModerator: boolean;
}

const formatTime = (nanoseconds: number): string => {
    const seconds = Math.floor(nanoseconds / 1000000000);
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    return `${minutes}:${remainingSeconds.toString().padStart(2, "0")}`;
  };

const QueueItem: React.FC<QueueItemProps> = ({ song, itemType, onUpvote, onDownvote, onDelete, onSkip, mode, userIsModerator }) => {
  const modeIsVoting = mode !== Mode.Simple;
  const canSkipAndDelete = mode === Mode.Simple || mode === Mode.Voting || userIsModerator;

  return (
    <div className="p-4 bg-gray-600 rounded-lg shadow-sm hover:bg-gray-500 transition-colors">
      <div className="flex flex-col">
        <h4 className="text-lg font-bold text-gray-100">{song.name}</h4>
        <p className="text-sm text-gray-300">Added by: {song.addedby}</p>
      </div>
        {itemType == QueueItemType.PLAY && (
            <div className="mt2 flex flex-col space-y-2">
            <progress value={song.position} max={song.length} className="w-full" />
            <p className="text-center text-sm text-gray-100">
                {formatTime(song.position)} / {formatTime(song.length)}
            </p>
            {canSkipAndDelete && (
            <button
                onClick={() => onSkip(song.id)}
                className="bg-yellow-500 hover:bg-yellow-700 text-white font-bold py-1 px-1 rounded"
            >
                Skip
            </button>
            )}
            </div>
        )}
        {itemType == QueueItemType.QUEUE && (
            <div className="mt2 flex flex-col space-y-2">
             { modeIsVoting && (
            <span className="text-md font-semibold text-gray-200">Votes: {song.points}</span>
              )
            }
            <div className="flex space-x-2">
            { modeIsVoting && (
              <>
              <button
                  onClick={() => onUpvote(song.id)}
                  className="bg-green-500 hover:bg-green-600 text-white rounded px-2 py-1 transition-colors"
                  title="Upvote"
              > ▲
              </button>
              <button
                  onClick={() => onDownvote(song.id)}
                  className="bg-red-500 hover:bg-red-600 text-white rounded px-2 py-1 transition-colors"
                  title="Downvote"
              >
                  ▼
              </button>
              </>
            ) }
            {canSkipAndDelete && (
            <button
                onClick={() => onDelete(song.id)}
                className="bg-gray-700 hover:bg-gray-600 text-white rounded px-2 py-1 transition-colors"
                title="Delete"
            >
                Delete
            </button>)}

            </div>
            </div>
        )}      
    </div>
  );
};

export default QueueItem;