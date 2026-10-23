import SimpleEvent from "./components/SimpleEvent.vue";
import ComplexEvents from "./components/ComplexEvents.vue";
import DependentEvents from "./components/DependentEvents.vue";
import FullGroupEvents from "./components/FullGroupEvents.vue";
import WheelOfFortune from "./components/WheelOfFortune.vue";
import ContinuousVariable from "./components/ContinuousVariable.vue";
import DiscreteVariable from "./components/DiscreteVariable.vue";
import type { LabSection } from "./types/navigation";

export const sections: LabSection[] = [
  {
    id: "events",
    label: "События",
    description:
      "Простые, независимые и зависимые события, полная группа и колесо фортуны.",
    icon: "events",
    tasks: [
      { id: "simple", label: "Простое событие", component: SimpleEvent },
      { id: "complex", label: "Независимые события", component: ComplexEvents },
      {
        id: "dependent",
        label: "Зависимые события",
        component: DependentEvents,
      },
      { id: "fullGroup", label: "Полная группа", component: FullGroupEvents },
      { id: "wheel", label: "Колесо фортуны", component: WheelOfFortune },
    ],
  },
