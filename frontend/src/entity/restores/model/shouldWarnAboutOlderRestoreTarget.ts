import type { RestoreCompatibility } from './RestoreCompatibility';

// The older-server warning offers to restore anyway, so it only makes sense when the backend
// would accept the restore. A target older than the backed-up server itself is always refused,
// and the restore request then reports that reason instead.
export const shouldWarnAboutOlderRestoreTarget = (
  compatibility: RestoreCompatibility,
  sourceServerMajorVersion: string | undefined,
): boolean => {
  if (!compatibility.isTargetOlderThanDumpClient) {
    return false;
  }

  const sourceMajor = Number(sourceServerMajorVersion);
  if (!Number.isFinite(sourceMajor) || sourceMajor <= 0) {
    return true;
  }

  return sourceMajor <= compatibility.targetMajorVersion;
};
