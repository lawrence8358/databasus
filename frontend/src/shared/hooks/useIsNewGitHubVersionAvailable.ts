import { useEffect, useState } from 'react';

import { APP_VERSION } from '../../constants';
import { compareVersions, parseGitHubTag } from '../lib';

// eslint-disable-next-line i18next/no-literal-string -- localStorage key
const CACHE_KEY_PREFIX = 'githubLatestVersion';
const CACHE_TTL_MS = 15 * 60 * 1000;

interface CachedVersion {
  version: string;
  fetchedAt: number;
}

function readCachedVersion(cacheKey: string): string | null {
  try {
    const raw = localStorage.getItem(cacheKey);
    if (!raw) {
      return null;
    }

    const cached = JSON.parse(raw) as CachedVersion;

    if (Date.now() - cached.fetchedAt < CACHE_TTL_MS) {
      return cached.version;
    }

    return null;
  } catch {
    return null;
  }
}

async function fetchLatestVersion(repository: string, cacheKey: string): Promise<string | null> {
  try {
    const response = await fetch(`https://api.github.com/repos/${repository}/releases/latest`, {
      headers: { Accept: 'application/vnd.github+json' },
    });

    if (!response.ok) {
      return null;
    }

    const release = (await response.json()) as { tag_name?: string };
    if (!release.tag_name) {
      return null;
    }

    const version = parseGitHubTag(release.tag_name);

    localStorage.setItem(
      cacheKey,
      JSON.stringify({ version, fetchedAt: Date.now() } satisfies CachedVersion),
    );

    return version;
  } catch {
    return null;
  }
}

// Checks the releases of the repository the interface links to (see getGitHubRepository).
export function useIsNewGitHubVersionAvailable(repository: string): boolean {
  const [latestVersion, setLatestVersion] = useState<string | null>(null);

  useEffect(() => {
    if (APP_VERSION === 'dev') {
      return;
    }

    const cacheKey = `${CACHE_KEY_PREFIX}:${repository}`;
    const cachedVersion = readCachedVersion(cacheKey);
    if (cachedVersion) {
      setLatestVersion(cachedVersion);
      return;
    }

    let isActive = true;

    fetchLatestVersion(repository, cacheKey).then((version) => {
      if (isActive && version) {
        setLatestVersion(version);
      }
    });

    return () => {
      isActive = false;
    };
  }, [repository]);

  return latestVersion !== null && compareVersions(latestVersion, APP_VERSION) > 0;
}
