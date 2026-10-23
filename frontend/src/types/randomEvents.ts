export interface EventOutcome {
  label: string;
  count: number;
  frequency: number;
  probability: number;
}

export interface SimulationResult {
  title: string;
  sampleSize: number;
  outcomes: EventOutcome[];
}
