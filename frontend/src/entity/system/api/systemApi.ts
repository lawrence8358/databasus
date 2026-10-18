import { getApplicationServer } from '../../../constants';
import { apiHelper } from '../../../shared/api/apiHelper';
import RequestOptions from '../../../shared/api/RequestOptions';
import type { Availability } from '../model/Availability';
import type { BackupsFolder } from '../model/BackupsFolder';
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

  async getBackupsFolder(): Promise<BackupsFolder> {
    return apiHelper.fetchGetJson<BackupsFolder>(
      `${getApplicationServer()}/api/v1/system/backups-folder`,
      new RequestOptions(),
    );
  },

  async changeBackupsFolder(folder: string): Promise<BackupsFolder> {
    const requestOptions = new RequestOptions();
    requestOptions.setBody(JSON.stringify({ folder }));

    return apiHelper.fetchPutJson<BackupsFolder>(
      `${getApplicationServer()}/api/v1/system/backups-folder`,
      requestOptions,
    );
  },
};
