import type { Component } from "vue";

export type SectionId = "events" | "variables" | "bivariate" | "queueing" | "petri";

export type SectionIconName = SectionId;

export interface LabTask {
  id: string;
  label: string;
  component: Component;
  props?: Record<string, string>;
}

export interface LabSection {
  id: SectionId;
  label: string;
  description: string;
  icon: SectionIconName;
  tasks: LabTask[];
}
