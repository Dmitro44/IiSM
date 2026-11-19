export interface DensityBin {
  lower: number;
  upper: number;
  density: number;
  theoreticalDensity: number;
  count: number;
}

export interface CurvePoint {
  x: number;
  y: number;
}

export interface JointCellPoint {
  centerX: number;
  centerY: number;
  count: number;
  expectedCount: number;
}

export interface HypothesisTestView {
  name: string;
  statistic: number;
  degreesOfFreedom: number;
  pValue: number;
  alpha: number;
  accepted: boolean;
  valid: boolean;
  note: string;
}
