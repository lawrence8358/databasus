import { useEffect, useState } from 'react';

import { systemApi } from '../api/systemApi';
import type { Availability } from './Availability';
import type { AvailabilityEngine } from './AvailabilityEngine';
import type { AvailabilityFeature } from './AvailabilityFeature';

// Availability is fixed for the lifetime of the backend process, so one request serves every
// component; a failed request is dropped so the next mount retries it.
let availabilityRequest: Promise<Availability> | undefined;

const loadAvailability = (): Promise<Availability> => {
  if (!availabilityRequest) {
    availabilityRequest = systemApi.getAvailability().catch((e) => {
      availabilityRequest = undefined;
      throw e;
    });
  }

  return availabilityRequest;
};

// Until the answer arrives everything counts as available: that is what every Docker instance
// reports, and the backend refuses an unavailable engine or feature anyway.
export const useAvailability = () => {
  const [availability, setAvailability] = useState<Availability | undefined>();

  const isEngineAvailable = (engine: AvailabilityEngine): boolean =>
    availability?.engines.find((item) => item.engine === engine)?.isAvailable ?? true;

  const isFeatureAvailable = (feature: AvailabilityFeature): boolean =>
    availability?.features.find((item) => item.feature === feature)?.isAvailable ?? true;

  useEffect(() => {
    let isMounted = true;

    loadAvailability()
      .then((loadedAvailability) => {
        if (isMounted) setAvailability(loadedAvailability);
      })
      .catch(() => {});

    return () => {
      isMounted = false;
    };
  }, []);

  return { isEngineAvailable, isFeatureAvailable };
};
