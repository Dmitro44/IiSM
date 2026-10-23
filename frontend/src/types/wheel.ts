export interface WheelGame {
  name: string;
  /** Пожертвования за игру в белорусских рублях. */
  amount: number;
  probability: number;
  wins: number;
}

export interface WheelState {
  games: WheelGame[];
  total: number;
  spins: number;
}

export interface WheelSpin {
  winner: string;
  state: WheelState;
}
