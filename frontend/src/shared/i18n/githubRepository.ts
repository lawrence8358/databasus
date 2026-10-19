/* eslint-disable i18next/no-literal-string -- repository names */
import type { Locale } from './dictionaries';

const UPSTREAM_REPOSITORY = 'databasus/databasus';

// The Traditional Chinese interface ships with the Windows fork, so its GitHub links lead there.
const FORK_REPOSITORY = 'lawrence8358/databasus';

export const getGitHubRepository = (locale: Locale): string =>
  locale === 'zh' ? FORK_REPOSITORY : UPSTREAM_REPOSITORY;

export const getGitHubRepositoryUrl = (locale: Locale): string =>
  `https://github.com/${getGitHubRepository(locale)}`;
