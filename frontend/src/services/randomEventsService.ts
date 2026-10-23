import {
  SimulateComplex,
  SimulateDependent,
  SimulateFullGroup,
  SimulateSimple,
} from "../../bindings/github.com/Dmitro44/IiSM/internal/events/eventservice";
import type {
  Outcome,
  Result,
} from "../../bindings/github.com/Dmitro44/IiSM/internal/events/models";
import type { EventOutcome, SimulationResult } from "../types/randomEvents";

function toOutcomes(
  outcomes: Outcome[],
  labels: string[],
): EventOutcome[] {
  if (outcomes.length !== labels.length) {
    throw new Error("service returned unexpected count of outcomes.");
  }

  return outcomes.map((outcome: Outcome, index: number) => ({
    label: labels[index],
    count: outcome.Count,
    frequency: outcome.Frequency,
    probability: outcome.Probability,
  }));
}

function present(
  title: string,
  result: Result,
  labels: string[],
): SimulationResult {
  return {
    title,
    sampleSize: result.SampleSize,
    outcomes: toOutcomes(result.Outcomes ?? [], labels),
  };
}

// Строка index — маска: бит i означает, что событие i+1 произошло; порядок как в Result.
function jointLabels(k: number): string[] {
  return Array.from({ length: 1 << k }, (_, index) =>
    Array.from(
      { length: k },
      (_, bit) => `${(index >> bit) & 1 ? "" : "¬"}A${bit + 1}`,
    ).join(""),
  );
}

export async function simulateSimpleEvent(
  p: number,
  n: number,
): Promise<SimulationResult> {
  return present("Простое случайное событие A", await SimulateSimple(p, n), [
    "A произошло",
    "A не произошло",
  ]);
}

export async function simulateComplexEvents(
  probabilities: number[],
  n: number,
): Promise<SimulationResult> {
  return present(
    "Независимые случайные события",
    await SimulateComplex(probabilities, n),
    jointLabels(probabilities.length),
  );
}

export async function simulateDependentEvents(
  pA: number,
  pBA: number,
  n: number,
): Promise<SimulationResult> {
  return present(
    "Зависимые события A и B",
    await SimulateDependent(pA, pBA, n),
    ["AB (0)", "A¬B (1)", "¬AB (2)", "¬A¬B (3)"],
  );
}

export async function simulateFullGroupEvents(
  probabilities: number[],
  n: number,
): Promise<SimulationResult> {
  const labels = probabilities.map((_, index) => `A${index + 1}`);
  return present(
    "Полная группа событий",
    await SimulateFullGroup(probabilities, n),
    labels,
  );
}
