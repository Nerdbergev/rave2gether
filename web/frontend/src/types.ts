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
  
export enum UserRight {
    USER,
    MODERATOR,
    ADMIN
}

export interface User {
    username: string;
    rights: UserRight;
    coins: number;
}