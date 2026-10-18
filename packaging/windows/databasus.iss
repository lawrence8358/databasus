; Offline installer for Databasus on Windows x64. Built by build.ps1, which stages the files and
; passes AppVersion and StageDir. Installs the files, registers the Databasus service through
; databasus-launcher.exe, starts it and waits until the Web UI answers.
;
; Silent install:   databasus-setup-<version>-x64.exe /VERYSILENT /SUPPRESSMSGBOXES
;                     [/DIR=<install dir>] [/DATADIR=<dir>] [/BACKUPSDIR=<dir>] [/PORT=<port>] [/LOG=<file>]
; Silent uninstall: "<install dir>\unins000.exe" /VERYSILENT /SUPPRESSMSGBOXES [/DELETEDATA]

#ifndef AppVersion
  #define AppVersion "0.0.0-dev"
#endif
#ifndef StageDir
  #define StageDir "out\stage"
#endif

#define ServiceName "Databasus"
#define SettingsKey "SOFTWARE\Databasus"
#define DefaultPort "4005"

[Setup]
AppId={{5BAD0061-73B2-4D9C-A2F3-CEF52B7FAAE1}
AppName=Databasus
AppVersion={#AppVersion}
AppVerName=Databasus {#AppVersion}
AppPublisher=Databasus
AppPublisherURL=https://github.com/lawrence8358/databasus
DefaultDirName={autopf}\Databasus
DisableProgramGroupPage=yes
PrivilegesRequired=admin
; x64os rejects 32-bit Windows and Windows on Arm, which are not supported.
ArchitecturesAllowed=x64os
ArchitecturesInstallIn64BitMode=x64os
; 10.0.14393 is Windows Server 2016 and Windows 10 version 1607.
MinVersion=10.0.14393
OutputBaseFilename=databasus-setup-{#AppVersion}-x64
Compression=lzma2/max
SolidCompression=yes
WizardStyle=modern
SetupLogging=yes
UninstallDisplayName=Databasus
UninstallDisplayIcon={app}\databasus-launcher.exe
; The service holds the program files open; [Code] stops it instead of Restart Manager.
CloseApplications=no
RestartApplications=no

[Messages]
WinVersionTooLowError=Databasus requires Windows Server 2016, Windows 10 version 1607 or a newer 64-bit Windows. This computer runs an older version, so Setup cannot continue.
OnlyOnTheseArchitectures=Databasus can only be installed on 64-bit Windows for x64 processors. This computer has a different processor architecture or runs 32-bit Windows, so Setup cannot continue.

[Files]
Source: "{#StageDir}\*"; Excludes: "\databasus-launcher.exe"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs
; Last entry: registering and starting the service needs every other file in place.
Source: "{#StageDir}\databasus-launcher.exe"; DestDir: "{app}"; Flags: ignoreversion; AfterInstall: InstallAndStartService

[Registry]
; Read back by the uninstaller and by later installers (upgrade keeps these choices).
Root: HKLM; Subkey: "{#SettingsKey}"; Flags: uninsdeletekey
Root: HKLM; Subkey: "{#SettingsKey}"; ValueType: string; ValueName: "InstallDir"; ValueData: "{app}"
Root: HKLM; Subkey: "{#SettingsKey}"; ValueType: string; ValueName: "DataDir"; ValueData: "{code:GetDataDir}"
Root: HKLM; Subkey: "{#SettingsKey}"; ValueType: string; ValueName: "BackupsDir"; ValueData: "{code:GetBackupsDir}"
Root: HKLM; Subkey: "{#SettingsKey}"; ValueType: string; ValueName: "WebPort"; ValueData: "{code:GetPort}"
Root: HKLM; Subkey: "{#SettingsKey}"; ValueType: string; ValueName: "WebListenHost"; ValueData: "{code:GetListenHost}"
Root: HKLM; Subkey: "{#SettingsKey}"; ValueType: string; ValueName: "Version"; ValueData: "{#AppVersion}"

[Run]
Filename: "{code:GetWebUiUrl}"; Description: "Open the Databasus Web UI"; Flags: shellexec postinstall skipifsilent nowait runasoriginaluser

[UninstallRun]
Filename: "{app}\databasus-launcher.exe"; Parameters: "uninstall"; Flags: runhidden waituntilterminated; RunOnceId: "RemoveService"

[UninstallDelete]
; Written by the launcher at every start.
Type: files; Name: "{app}\ui\build\runtime-config.js"

[Code]
const
  ListenHost = '127.0.0.1';
  ServiceStartTimeoutSeconds = 300;
  ServiceStopTimeoutSeconds = 180;

var
  FoldersPage: TInputDirWizardPage;
  PortPage: TInputQueryWizardPage;
  IsLauncherExtracted: Boolean;
  UninstallDataDir: String;
  UninstallBackupsDir: String;
  IsDeleteDataRequested: Boolean;

{ ---------- settings ---------- }

function GetPreviousSetting(const Name, Default: String): String;
begin
  if not RegQueryStringValue(HKLM, '{#SettingsKey}', Name, Result) or (Trim(Result) = '') then
    Result := Default;
end;

function GetDataDir(Param: String): String;
begin
  Result := RemoveBackslashUnlessRoot(Trim(FoldersPage.Values[0]));
end;

function GetBackupsDir(Param: String): String;
begin
  Result := RemoveBackslashUnlessRoot(Trim(FoldersPage.Values[1]));
end;

function GetPort(Param: String): String;
begin
  Result := Trim(PortPage.Values[0]);
end;

function GetListenHost(Param: String): String;
begin
  Result := ListenHost;
end;

function GetWebUiUrl(Param: String): String;
begin
  Result := 'http://' + ListenHost + ':' + GetPort('') + '/';
end;

function IsServiceInstalled: Boolean;
begin
  Result := RegKeyExists(HKLM, 'SYSTEM\CurrentControlSet\Services\{#ServiceName}');
end;

{ ---------- running programs ---------- }

function JoinLines(const Lines: TArrayOfString): String;
var
  I: Integer;
begin
  Result := '';
  for I := 0 to GetArrayLength(Lines) - 1 do
    if Trim(Lines[I]) <> '' then
      Result := Result + Trim(Lines[I]) + #13#10;
end;

function RunAndCapture(const FileName, Params: String; var Output: String): Integer;
var
  ExecOutput: TExecOutput;
begin
  Log('Running: ' + FileName + ' ' + Params);
  if not ExecAndCaptureOutput(FileName, Params, '', SW_HIDE, ewWaitUntilTerminated, Result, ExecOutput) then
  begin
    Output := SysErrorMessage(Result);
    Result := -1;
  end
  else
    Output := JoinLines(ExecOutput.StdOut) + JoinLines(ExecOutput.StdErr);
  Log('Exit code ' + IntToStr(Result) + ': ' + Output);
end;

function GetLauncherForChecks: String;
begin
  if not IsLauncherExtracted then
  begin
    ExtractTemporaryFile('databasus-launcher.exe');
    IsLauncherExtracted := True;
  end;
  Result := ExpandConstant('{tmp}\databasus-launcher.exe');
end;

{ ---------- validation ---------- }

function ValidatePort(const Port: String): String;
var
  Value: Integer;
begin
  Result := '';
  Value := StrToIntDef(Port, 0);
  if (Value < 1) or (Value > 65535) or (IntToStr(Value) <> Port) then
    Result := 'The Web UI port must be a number from 1 to 65535.';
end;

{ Asks the launcher, which evaluates the folders' permissions for the LocalSystem service account
  and checks the port is free. Returns the problems, one per line, or '' when there are none. }
function CheckChoices(const CheckFolders, CheckPort: Boolean): String;
var
  Params, Output: String;
  ExitCode: Integer;
begin
  Result := '';
  if CheckPort then
  begin
    Result := ValidatePort(GetPort(''));
    if Result <> '' then
      exit;
  end;

  Params := 'check-install';
  if CheckFolders then
    Params := Params + ' -data-dir "' + GetDataDir('') + '" -backups-dir "' + GetBackupsDir('') + '"';
  { On a reinstall the running service holds its own port, so only a fresh install checks it. }
  if CheckPort and not IsServiceInstalled then
    Params := Params + ' -port ' + GetPort('');

  ExitCode := RunAndCapture(GetLauncherForChecks, Params, Output);
  if ExitCode <> 0 then
  begin
    Result := Trim(Output);
    if Result = '' then
      Result := 'The installation checks failed (exit code ' + IntToStr(ExitCode) + ').';
  end;
end;

{ ---------- wizard ---------- }

procedure InitializeWizard;
var
  DefaultDataDir: String;
begin
  DefaultDataDir := ExpandConstant('{commonappdata}\Databasus');

  FoldersPage := CreateInputDirPage(wpSelectDir,
    'Data folders', 'Where should Databasus keep its data and backups?',
    'The Databasus service runs as LocalSystem and must be able to write to both folders. ' +
    'Both must be on a local disk; network paths are not supported.',
    False, 'New Folder');
  FoldersPage.Add('Program data folder (metadata database, secret key, logs):');
  FoldersPage.Add('Backups folder (local storage):');
  FoldersPage.Values[0] := ExpandConstant('{param:DATADIR|' + GetPreviousSetting('DataDir', DefaultDataDir) + '}');
  FoldersPage.Values[1] := ExpandConstant('{param:BACKUPSDIR|' + GetPreviousSetting('BackupsDir', AddBackslash(DefaultDataDir) + 'backups') + '}');

  PortPage := CreateInputQueryPage(FoldersPage.ID,
    'Web UI port', 'Which port should the Databasus Web UI use?',
    'The Web UI listens on 127.0.0.1 only, so it is reachable from this computer and nowhere else.');
  PortPage.Add('Port:', False);
  PortPage.Values[0] := ExpandConstant('{param:PORT|' + GetPreviousSetting('WebPort', '{#DefaultPort}') + '}');
end;

function NextButtonClick(CurPageID: Integer): Boolean;
var
  Problems: String;
begin
  Result := True;
  if WizardSilent then
    exit; { PrepareToInstall checks everything once in silent mode. }

  Problems := '';
  if CurPageID = FoldersPage.ID then
    Problems := CheckChoices(True, False)
  else if CurPageID = PortPage.ID then
    Problems := CheckChoices(False, True);

  if Problems <> '' then
  begin
    MsgBox(Problems, mbError, MB_OK);
    Result := False;
  end;
end;

function WaitForServiceState(const State: String; TimeoutSeconds: Integer): Boolean;
var
  Output: String;
  Waited: Integer;
begin
  Result := False;
  Waited := 0;
  while Waited < TimeoutSeconds do
  begin
    RunAndCapture(ExpandConstant('{sys}\sc.exe'), 'query {#ServiceName}', Output);
    if Pos(State, Output) > 0 then
    begin
      Result := True;
      exit;
    end;
    Sleep(1000);
    Waited := Waited + 1;
  end;
end;

function PrepareToInstall(var NeedsRestart: Boolean): String;
var
  Output: String;
begin
  Result := '';
  if WizardSilent then
  begin
    Result := CheckChoices(True, True);
    if Result <> '' then
      exit;
  end;

  { A reinstall replaces files the running service holds open. }
  if IsServiceInstalled then
  begin
    RunAndCapture(ExpandConstant('{sys}\sc.exe'), 'stop {#ServiceName}', Output);
    if not WaitForServiceState('STOPPED', ServiceStopTimeoutSeconds) then
      Result := 'The existing Databasus service did not stop. Stop it in the Services console and run Setup again.';
  end;
end;

function IsWebUiResponding: Boolean;
var
  Request: Variant;
begin
  Result := False;
  try
    Request := CreateOleObject('WinHttp.WinHttpRequest.5.1');
    Request.SetTimeouts(3000, 3000, 3000, 3000);
    Request.Open('GET', 'http://' + ListenHost + ':' + GetPort('') + '/api/v1/system/version', False);
    Request.Send('');
    Result := Request.Status = 200;
  except
    Result := False;
  end;
end;

procedure FailInstall(const Message: String);
begin
  Log('Install failed: ' + Message);
  RaiseException(Message);
end;

procedure InstallAndStartService;
var
  Params, Output, LogsHint: String;
  Waited: Integer;
  WasServiceInstalled: Boolean;
begin
  WasServiceInstalled := IsServiceInstalled;
  LogsHint := ' See ' + AddBackslash(GetDataDir('')) + 'logs\launcher.log and the Application event log.';

  WizardForm.StatusLabel.Caption := 'Registering the Databasus service...';
  Params := 'install -data-dir "' + GetDataDir('') + '" -backups-dir "' + GetBackupsDir('') + '"' +
    ' -listen-host ' + ListenHost + ' -port ' + GetPort('');
  if RunAndCapture(ExpandConstant('{app}\databasus-launcher.exe'), Params, Output) <> 0 then
    FailInstall('Registering the Databasus service failed:' + #13#10 + Output);

  WizardForm.StatusLabel.Caption := 'Starting the Databasus service (the first start takes up to a few minutes)...';
  if RunAndCapture(ExpandConstant('{sys}\sc.exe'), 'start {#ServiceName}', Output) <> 0 then
  begin
    if not WasServiceInstalled then
      RunAndCapture(ExpandConstant('{app}\databasus-launcher.exe'), 'uninstall', Output);
    FailInstall('Starting the Databasus service failed:' + #13#10 + Output + LogsHint);
  end;

  Waited := 0;
  while not IsWebUiResponding do
  begin
    if Waited >= ServiceStartTimeoutSeconds then
    begin
      if not WasServiceInstalled then
        RunAndCapture(ExpandConstant('{app}\databasus-launcher.exe'), 'uninstall', Output);
      FailInstall('The Databasus service did not answer on ' + GetWebUiUrl('') + ' within ' +
        IntToStr(ServiceStartTimeoutSeconds) + ' seconds.' + LogsHint);
    end;
    Sleep(2000);
    Waited := Waited + 2;
  end;
  Log('The Web UI answers on ' + GetWebUiUrl(''));
end;

{ ---------- uninstall ---------- }

function HasCommandLineSwitch(const Switch: String): Boolean;
var
  I: Integer;
begin
  Result := False;
  for I := 1 to ParamCount do
    if CompareText(ParamStr(I), Switch) = 0 then
      Result := True;
end;

{ Guards the "delete all data" option against a folder setting that points at a drive root or a
  system folder, which DelTree would otherwise wipe. }
function IsSafeToDelete(const Folder: String): Boolean;
var
  Normalized: String;
begin
  Normalized := Lowercase(RemoveBackslashUnlessRoot(Trim(Folder)));
  Result := (Length(Normalized) > 3) and
    (Normalized <> Lowercase(ExpandConstant('{win}'))) and
    (Normalized <> Lowercase(ExpandConstant('{commonpf}'))) and
    (Normalized <> Lowercase(ExpandConstant('{commonpf32}'))) and
    (Normalized <> Lowercase(ExpandConstant('{commonappdata}'))) and
    (Normalized <> Lowercase(GetEnv('SystemDrive') + '\users'));
end;

function InitializeUninstall: Boolean;
begin
  Result := True;
  RegQueryStringValue(HKLM, '{#SettingsKey}', 'DataDir', UninstallDataDir);
  RegQueryStringValue(HKLM, '{#SettingsKey}', 'BackupsDir', UninstallBackupsDir);

  IsDeleteDataRequested := HasCommandLineSwitch('/DELETEDATA');
  if UninstallSilent then
    exit;

  if not IsDeleteDataRequested then
    IsDeleteDataRequested := MsgBox(
      'Databasus keeps its data and backups when it is uninstalled:' + #13#10 + #13#10 +
      '  ' + UninstallDataDir + #13#10 + '  ' + UninstallBackupsDir + #13#10 + #13#10 +
      'Do you also want to delete ALL data, including every backup in the backups folder?',
      mbConfirmation, MB_YESNO or MB_DEFBUTTON2) = IDYES;

  if IsDeleteDataRequested then
    IsDeleteDataRequested := MsgBox(
      'This permanently deletes the metadata database, the secret key and all backups.' + #13#10 + #13#10 +
      'Without the secret key, encrypted settings and encrypted backups can never be recovered, ' +
      'even if you reinstall Databasus.' + #13#10 + #13#10 +
      'Delete all data?',
      mbCriticalError, MB_YESNO or MB_DEFBUTTON2) = IDYES;
end;

procedure DeleteDataFolder(const Folder: String);
begin
  if Trim(Folder) = '' then
    exit;
  if not IsSafeToDelete(Folder) then
  begin
    Log('Not deleting ' + Folder + ': it is a drive root or a system folder');
    exit;
  end;
  if DirExists(Folder) and not DelTree(Folder, True, True, True) then
    Log('Could not delete all of ' + Folder);
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if (CurUninstallStep = usPostUninstall) and IsDeleteDataRequested then
  begin
    Log('Deleting all data as requested');
    DeleteDataFolder(UninstallBackupsDir);
    DeleteDataFolder(UninstallDataDir);
  end;
end;
