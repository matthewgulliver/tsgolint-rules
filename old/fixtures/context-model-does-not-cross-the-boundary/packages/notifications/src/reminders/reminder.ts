// expect: context-model-does-not-cross-the-boundary
import type { Occasion } from "../../../gifting/src/occasions/occasion"
export const remindAbout = (occasion: Occasion): string => occasion.id
