import { describe, expect, it } from 'vitest';

import { shouldWarnAboutOlderRestoreTarget } from './shouldWarnAboutOlderRestoreTarget';

const olderThanClient = (targetMajorVersion: number) => ({
  isTargetOlderThanDumpClient: true,
  dumpClientMajorVersion: 18,
  targetMajorVersion,
});

describe('shouldWarnAboutOlderRestoreTarget', () => {
  it('warns when the target is older than the dump client but not older than the source server', () => {
    expect(shouldWarnAboutOlderRestoreTarget(olderThanClient(16), '16')).toBe(true);
    expect(shouldWarnAboutOlderRestoreTarget(olderThanClient(16), '12')).toBe(true);
  });

  it('does not warn when the backend will refuse the restore because the source server is newer', () => {
    expect(shouldWarnAboutOlderRestoreTarget(olderThanClient(16), '18')).toBe(false);
  });

  it('does not warn when the target is not older than the dump client', () => {
    expect(
      shouldWarnAboutOlderRestoreTarget(
        { isTargetOlderThanDumpClient: false, dumpClientMajorVersion: 18, targetMajorVersion: 18 },
        '18',
      ),
    ).toBe(false);
  });

  it('warns when the source server version is unknown', () => {
    expect(shouldWarnAboutOlderRestoreTarget(olderThanClient(16), undefined)).toBe(true);
  });
});
