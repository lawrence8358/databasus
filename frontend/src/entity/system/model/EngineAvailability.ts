import type { AvailabilityEngine } from './AvailabilityEngine';

export interface EngineAvailability {
  engine: AvailabilityEngine;
  isAvailable: boolean;
  unavailableReason?: string;
}
