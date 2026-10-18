import { describe, expect, it } from 'vitest';

import { getGitHubRepository, getGitHubRepositoryUrl } from './githubRepository';

describe('getGitHubRepositoryUrl', () => {
  it('links the Chinese interface to the fork', () => {
    expect(getGitHubRepositoryUrl('zh')).toBe('https://github.com/lawrence8358/databasus');
  });

  it('links every other language to the upstream repository', () => {
    expect(getGitHubRepository('en')).toBe('databasus/databasus');
    expect(getGitHubRepository('ru')).toBe('databasus/databasus');
  });
});
