import { getApplicationServer } from '../../../constants';
import { apiHelper } from '../../../shared/api/apiHelper';
import RequestOptions from '../../../shared/api/RequestOptions';
import type { Availability } from '../model/Availability';
import type { VersionResponse } from '../model/VersionResponse';

export const systemApi = {
  async getVersion(): Promise<VersionResponse> {
    const response = await fetch(`${getApplicationServer()}/api/v1/system/version`, {
      cache: 'no-store',
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch version: ${response.status}`);
    }

    return response.json();
  },

  async getAvailability(): Promise<Availability> {
    return apiHelper.fetchGetJson<Availability>(
      `${getApplicationServer()}/api/v1/system/availability`,
      new RequestOptions(),
      true,
    );
  },
};
