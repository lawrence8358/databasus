export interface RestoreCompatibility {
  isTargetOlderThanDumpClient: boolean;
  dumpClientMajorVersion: number;
  targetMajorVersion: number;
}
