import type { EngineAvailability } from './EngineAvailability';
import type { FeatureAvailability } from './FeatureAvailability';

export interface Availability {
  engines: EngineAvailability[];
  features: FeatureAvailability[];
}
