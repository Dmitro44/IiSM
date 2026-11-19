import {
  SimulateContinuous,
  SimulateDiscrete,
} from "../../bindings/github.com/Dmitro44/IiSM/internal/bivariate/bivariateservice";
import type {
  ContinuousRequest,
  ContinuousResult,
  DiscreteRequest,
  DiscreteResult,
} from "../../bindings/github.com/Dmitro44/IiSM/internal/bivariate/models";

export type {
  ContinuousRequest,
  ContinuousResult,
  DiscreteRequest,
  DiscreteResult,
};

export function simulateContinuous(
  request: ContinuousRequest,
): Promise<ContinuousResult> {
  return SimulateContinuous(request);
}

export function simulateDiscrete(
  request: DiscreteRequest,
): Promise<DiscreteResult> {
  return SimulateDiscrete(request);
}
