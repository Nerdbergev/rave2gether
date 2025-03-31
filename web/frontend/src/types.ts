// src/types.ts
export interface Song {
    id:         string;
    name:       string;
    url:        string    
	hash:       string    
	addedby:    string    
	addedAt:    string 
	playedAt:   string 
	points:     number
    position:  number
    length:     number   
  }

export interface QueueResponse {
    preparequeue: Song[];
    downloadqueue: Song[];
    playqueue: Song[];
}

export enum UserRight {
    USER,
    MODERATOR,
    ADMIN
}

export interface User {
    username: string;
    right: UserRight;
    coins: number;
}

export enum Mode {
    Simple = 0,
    Voting = 1,
    UserVoting = 2,
    UserCoin = 3,
  }

export interface AppConfig {
    mode: Mode;
}
  