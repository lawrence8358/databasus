import type { AvailabilityFeature } from './AvailabilityFeature';

export interface FeatureAvailability {
  feature: AvailabilityFeature;
  isAvailable: boolean;
  unavailableReason?: string;
}
