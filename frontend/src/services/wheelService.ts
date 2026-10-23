import {
  Donate,
  Reset,
  Snapshot,
  Spin,
} from "../../bindings/github.com/Dmitro44/IiSM/internal/wheel/wheelservice";
import type {
  Game,
  Spin as SpinModel,
  State,
} from "../../bindings/github.com/Dmitro44/IiSM/internal/wheel/models";
import type { WheelGame, WheelSpin, WheelState } from "../types/wheel";

function toGame(game: Game): WheelGame {
  return {
    name: game.Name,
    amount: game.Amount,
    probability: game.Probability,
    wins: game.Wins,
  };
}

function toState(state: State): WheelState {
  return {
    games: (state.Games ?? []).map(toGame),
    total: state.Total,
    spins: state.Spins,
  };
}

export async function donate(name: string, amount: number): Promise<WheelState> {
  return toState(await Donate(name, amount));
}

export async function spinWheel(): Promise<WheelSpin> {
  const result: SpinModel = await Spin();

  return { winner: result.Winner, state: toState(result.State) };
}

export async function wheelSnapshot(): Promise<WheelState> {
  return toState(await Snapshot());
}

export async function resetWheel(): Promise<WheelState> {
  return toState(await Reset());
}
