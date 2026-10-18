; Offline installer for Databasus on Windows x64. Built by build.ps1, which stages the files and
; passes AppVersion and StageDir. Installs the files, registers the Databasus service through
; databasus-launcher.exe, starts it and waits until the Web UI answers.
;
; Silent install:   databasus-setup-<version>-x64.exe /VERYSILENT /SUPPRESSMSGBOXES
;                     [/DIR=<install dir>] [/DATADIR=<dir>] [/BACKUPSDIR=<dir>] [/PORT=<port>] [/LOG=<file>]
;                   Exit code 0: installed and the Web UI answers. 7: a check failed before anything was
;                   installed (the reason is in the log). 10: files are installed but the service did
;                   not register or start (the reason is in the log).
;                   [/EXTERNALDBFILE=<file>] [/BUNDLEDDB]
;                   /EXTERNALDBFILE names a file holding the connection string of an external PostgreSQL
;                   metadata database (the string itself is not accepted on the command line, because
;                   Setup logs its command line); /BUNDLEDDB switches back to the bundled one. An upgrade
;                   keeps the earlier choice unless one of them is given.
; Upgrade:          run a newer setup over the installed one; it keeps the earlier data folder,
;                   backups folder and port (/DATADIR, /BACKUPSDIR and /PORT are ignored) and refuses,
;                   before stopping the service, a metadata database of another PostgreSQL major.
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
; The supported Windows version and architecture are checked in InitializeSetup, which can say
; what this computer has; the MinVersion and ArchitecturesAllowed directives show only a generic
; "does not support the version of Windows" message for both.
ArchitecturesInstallIn64BitMode=x64os
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

[Files]
Source: "{#StageDir}\*"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "{#StageDir}\databasus-launcher.exe"; DestDir: "{app}"; Flags: ignoreversion

[InstallDelete]
; An upgrade replaces these folders as a whole, so files the new version no longer ships (an old
; migration, old Web UI chunks, binaries of the previous PostgreSQL minor) do not linger.
Type: filesandordirs; Name: "{app}\migrations"
Type: filesandordirs; Name: "{app}\ui\build"
Type: filesandordirs; Name: "{app}\assets\tools\win-x64\postgresql"

[Registry]
; Read back by the uninstaller and by later installers (upgrade keeps these choices).
Root: HKLM; Subkey: "{#SettingsKey}"; Flags: uninsdeletekey
Root: HKLM; Subkey: "{#SettingsKey}"; ValueType: string; ValueName: "InstallDir"; ValueData: "{app}"
Root: HKLM; Subkey: "{#SettingsKey}"; ValueType: string; ValueName: "DataDir"; ValueData: "{code:GetDataDir}"
Root: HKLM; Subkey: "{#SettingsKey}"; ValueType: string; ValueName: "BackupsDir"; ValueData: "{code:GetBackupsDir}"
Root: HKLM; Subkey: "{#SettingsKey}"; ValueType: string; ValueName: "WebPort"; ValueData: "{code:GetPort}"
Root: HKLM; Subkey: "{#SettingsKey}"; ValueType: string; ValueName: "WebListenHost"; ValueData: "{code:GetListenHost}"
Root: HKLM; Subkey: "{#SettingsKey}"; ValueType: string; ValueName: "ExternalMetadataDb"; ValueData: "{code:GetExternalDbFlag}"
Root: HKLM; Subkey: "{#SettingsKey}"; ValueType: string; ValueName: "Version"; ValueData: "{#AppVersion}"

[Run]
Filename: "{code:GetWebUiUrl}"; Description: "Open the Databasus Web UI"; Flags: shellexec postinstall skipifsilent nowait runasoriginaluser; Check: IsServiceStarted

[UninstallRun]
Filename: "{app}\databasus-launcher.exe"; Parameters: "uninstall"; Flags: runhidden waituntilterminated; RunOnceId: "RemoveService"

[UninstallDelete]
; Written by the launcher at every start.
Type: files; Name: "{app}\ui\build\runtime-config.js"

[Code]
const
  ListenHost = '127.0.0.1';
  { Windows Server 2016 and Windows 10 version 1607 }
  MinimumWindowsBuild = 14393;
  ServiceStartTimeoutSeconds = 300;
  ServiceStopTimeoutSeconds = 180;
  ServiceFailedExitCode = 10;

var
  ServiceFailure: String;
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

{ ---------- external metadata database ---------- }

{ The connection string holds a password, so it never goes on a command line (visible to other
  users, written to the Setup log): a typed one reaches the launcher through an environment
  variable of this process, a file one by the path of a file, and the launcher stores it in the
  data folder with a restricted ACL. }
const
  ExternalDbEnvVariable = 'DATABASUS_INSTALLER_EXTERNAL_DB_DSN';
  ExternalDbFileName = 'external-metadata-db.dsn';

var
  ExternalDbPage: TWizardPage;
  ExternalDbCheckBox: TNewCheckBox;
  ExternalDbEdit: TPasswordEdit;
  ExternalDbFileParam: String;
  WasExternalDbChosen: Boolean;

function SetEnvironmentVariable(Name, Value: String): Boolean;
  external 'SetEnvironmentVariableW@kernel32.dll stdcall';

function IsBundledDbRequested: Boolean;
var
  I: Integer;
begin
  Result := False;
  for I := 1 to ParamCount do
    if CompareText(ParamStr(I), '/BUNDLEDDB') = 0 then
      Result := True;
end;

function IsExternalDbChosen: Boolean;
begin
  Result := ExternalDbCheckBox.Checked;
end;

function GetExternalDbFlag(Param: String): String;
begin
  if IsExternalDbChosen then
    Result := '1'
  else
    Result := '0';
end;

function GetExternalDbTypedValue: String;
begin
  Result := Trim(ExternalDbEdit.Text);
end;

{ The stored string of an earlier install, kept when the administrator types nothing new. }
function GetPreviousExternalDbFile: String;
begin
  Result := AddBackslash(GetPreviousSetting('DataDir', '')) + ExternalDbFileName;
end;

{ The file the launcher reads the string from, or '' when the administrator typed it. }
function GetExternalDbSourceFile: String;
begin
  Result := '';
  if GetExternalDbTypedValue <> '' then
    exit;
  if ExternalDbFileParam <> '' then
    Result := ExternalDbFileParam
  else
    Result := GetPreviousExternalDbFile;
end;

function GetExternalDbTargetFile: String;
begin
  Result := AddBackslash(GetDataDir('')) + ExternalDbFileName;
end;

function RunExternalDbCommand(const Params: String; var Output: String): Integer;
var
  TypedValue: String;
begin
  TypedValue := GetExternalDbTypedValue;
  if TypedValue <> '' then
    SetEnvironmentVariable(ExternalDbEnvVariable, TypedValue);
  try
    Result := RunAndCapture(GetLauncherForChecks, Params, Output);
  finally
    SetEnvironmentVariable(ExternalDbEnvVariable, '');
  end;
end;

{ Returns the problem in words for the person running Setup, or '' when the external database
  is not chosen or accepts the connection string. }
function CheckExternalDb: String;
var
  SourceFile, Params, Output: String;
  ExitCode: Integer;
begin
  Result := '';
  if not IsExternalDbChosen then
    exit;

  SourceFile := GetExternalDbSourceFile;
  Params := 'test-external-db';
  if SourceFile <> '' then
  begin
    if not FileExists(SourceFile) then
    begin
      if ExternalDbFileParam <> '' then
        Result := 'The connection string file ' + SourceFile + ' does not exist.'
      else
        Result := 'Enter the connection string of the external PostgreSQL database.';
      exit;
    end;
    Params := Params + ' -file "' + SourceFile + '"';
  end;

  ExitCode := RunExternalDbCommand(Params, Output);
  if ExitCode <> 0 then
  begin
    Result := Trim(Output);
    if Result = '' then
      Result := 'The external metadata database check failed (exit code ' + IntToStr(ExitCode) + ').';
    Result := Result + #13#10 + #13#10 + 'Nothing has been installed.';
  end;
end;

{ Stores the string for the service, or removes a string left by an earlier install that used one. }
function StoreExternalDb(var Failure: String): Boolean;
var
  SourceFile, Params: String;
begin
  Result := True;
  if not IsExternalDbChosen then
  begin
    if FileExists(GetExternalDbTargetFile) then
      DeleteFile(GetExternalDbTargetFile);
    exit;
  end;

  SourceFile := GetExternalDbSourceFile;
  Params := 'save-external-db -data-dir "' + GetDataDir('') + '"';
  if SourceFile <> '' then
    Params := Params + ' -file "' + SourceFile + '"';
  Result := RunExternalDbCommand(Params, Failure) = 0;
end;

procedure ExternalDbCheckBoxClick(Sender: TObject);
begin
  ExternalDbEdit.Enabled := ExternalDbCheckBox.Checked;
end;

procedure CreateExternalDbPage(const AfterPageId: Integer);
var
  Description, Prompt: TNewStaticText;
  PreviousFile: String;
begin
  ExternalDbFileParam := ExpandConstant('{param:EXTERNALDBFILE|}');
  PreviousFile := GetPreviousExternalDbFile;
  WasExternalDbChosen := (GetPreviousSetting('ExternalMetadataDb', '0') = '1') and FileExists(PreviousFile);

  ExternalDbPage := CreateCustomPage(AfterPageId,
    'Metadata database (advanced)', 'Where should Databasus keep its own settings and records?');

  Description := TNewStaticText.Create(ExternalDbPage);
  Description.Parent := ExternalDbPage.Surface;
  Description.WordWrap := True;
  Description.Width := ExternalDbPage.SurfaceWidth;
  Description.Caption :=
    'By default Databasus runs its own PostgreSQL for this. If your organization already runs and backs up ' +
    'a PostgreSQL server, Databasus can keep it there instead; the bundled PostgreSQL is then neither ' +
    'initialized nor started. The database must be empty or already used by Databasus.';

  ExternalDbCheckBox := TNewCheckBox.Create(ExternalDbPage);
  ExternalDbCheckBox.Parent := ExternalDbPage.Surface;
  ExternalDbCheckBox.Top := Description.Top + Description.Height + ScaleY(12);
  ExternalDbCheckBox.Width := ExternalDbPage.SurfaceWidth;
  ExternalDbCheckBox.Caption := 'Use an external PostgreSQL server';
  ExternalDbCheckBox.OnClick := @ExternalDbCheckBoxClick;

  Prompt := TNewStaticText.Create(ExternalDbPage);
  Prompt.Parent := ExternalDbPage.Surface;
  Prompt.Top := ExternalDbCheckBox.Top + ExternalDbCheckBox.Height + ScaleY(12);
  Prompt.WordWrap := True;
  Prompt.Width := ExternalDbPage.SurfaceWidth;
  Prompt.Caption := 'Connection string, for example ' +
    'postgres://user:password@host:5432/databasus?sslmode=require';
  if WasExternalDbChosen then
    Prompt.Caption := Prompt.Caption + #13#10 +
      'An external database is already configured. Leave this empty to keep its stored connection string.';

  ExternalDbEdit := TPasswordEdit.Create(ExternalDbPage);
  ExternalDbEdit.Parent := ExternalDbPage.Surface;
  ExternalDbEdit.Top := Prompt.Top + Prompt.Height + ScaleY(6);
  ExternalDbEdit.Width := ExternalDbPage.SurfaceWidth;

  ExternalDbCheckBox.Checked := (ExternalDbFileParam <> '') or (WasExternalDbChosen and not IsBundledDbRequested);
  ExternalDbCheckBoxClick(nil);
end;

{ ---------- supported systems ---------- }

function GetArchitectureName: String;
begin
  case ProcessorArchitecture of
    paX86: Result := '32-bit (x86)';
    paX64: Result := 'x64';
    paArm32: Result := 'ARM';
    paArm64: Result := 'ARM64';
  else
    Result := 'unknown';
  end;
end;

function GetUnsupportedSystemReason: String;
var
  Version: TWindowsVersion;
begin
  Result := '';
  GetWindowsVersionEx(Version);

  if not IsX64OS then
    Result := 'Databasus can only be installed on 64-bit Windows for x64 processors. ' +
      'This computer runs Windows for ' + GetArchitectureName + ' processors.'
  else if (Version.Major < 10) or ((Version.Major = 10) and (Version.Build < MinimumWindowsBuild)) then
    Result := 'Databasus requires Windows Server 2016, Windows 10 version 1607 or newer. ' +
      Format('This computer runs Windows %d.%d (build %d).', [Version.Major, Version.Minor, Version.Build]);
end;

function InitializeSetup: Boolean;
var
  Reason: String;
begin
  Reason := GetUnsupportedSystemReason;
  Result := Reason = '';
  if not Result then
  begin
    Log('Unsupported system: ' + Reason);
    SuppressibleMsgBox(Reason + #13#10 + #13#10 + 'Setup cannot continue.', mbCriticalError, MB_OK, IDOK);
  end;
end;

{ ---------- upgrade ---------- }

{ Running Setup over an installed Databasus keeps the program data folder, the backups folder, the
  port and the metadata database choice from the settings key the previous Setup wrote: the
  metadata database, secret key and backups stay where the service expects them, and the new
  service start runs the migrations. The external database choice is kept by its own page's
  defaults (/EXTERNALDBFILE and /BUNDLEDDB still change it); the page is only skipped here.
  The hooks below use event attributes so they run beside the main event functions. }

var
  IsUpgradeChecked: Boolean;
  IsUpgradeCached: Boolean;

{ Decided once, before [Registry] writes the settings key for a fresh install too. }
function IsUpgrade: Boolean;
begin
  if not IsUpgradeChecked then
  begin
    IsUpgradeCached := RegValueExists(HKLM, '{#SettingsKey}', 'DataDir');
    IsUpgradeChecked := True;
    if IsUpgradeCached then
      Log('Upgrading Databasus ' + GetPreviousSetting('Version', '(unknown version)') + ' to {#AppVersion}');
  end;
  Result := IsUpgradeCached;
end;

{ Moving the folders during an upgrade would start the service on an empty data folder, so the
  previous choices win over /DATADIR, /BACKUPSDIR and /PORT. }
procedure KeepPreviousSettingsOnUpgrade;
var
  Previous: String;
begin
  if not IsUpgrade then
    exit;

  Previous := GetPreviousSetting('DataDir', FoldersPage.Values[0]);
  if CompareText(RemoveBackslashUnlessRoot(Trim(FoldersPage.Values[0])), Previous) <> 0 then
    Log('Upgrade: keeping the program data folder ' + Previous + ' instead of ' + FoldersPage.Values[0]);
  FoldersPage.Values[0] := Previous;

  Previous := GetPreviousSetting('BackupsDir', FoldersPage.Values[1]);
  if CompareText(RemoveBackslashUnlessRoot(Trim(FoldersPage.Values[1])), Previous) <> 0 then
    Log('Upgrade: keeping the backups folder ' + Previous + ' instead of ' + FoldersPage.Values[1]);
  FoldersPage.Values[1] := Previous;

  Previous := GetPreviousSetting('WebPort', PortPage.Values[0]);
  if Trim(PortPage.Values[0]) <> Previous then
    Log('Upgrade: keeping the Web UI port ' + Previous + ' instead of ' + PortPage.Values[0]);
  PortPage.Values[0] := Previous;
end;

<event('ShouldSkipPage')>
function UpgradeShouldSkipPage(PageID: Integer): Boolean;
begin
  Result := IsUpgrade and
    ((PageID = FoldersPage.ID) or (PageID = PortPage.ID) or (PageID = ExternalDbPage.ID));
end;

<event('CurPageChanged')>
procedure UpgradeCurPageChanged(CurPageID: Integer);
begin
  if (CurPageID = wpReady) and IsUpgrade then
  begin
    WizardForm.ReadyMemo.Lines.Add('');
    WizardForm.ReadyMemo.Lines.Add('Upgrade from Databasus ' + GetPreviousSetting('Version', '(unknown version)') +
      '. These stay as they are, with all their data:');
    WizardForm.ReadyMemo.Lines.Add('      Program data folder: ' + GetDataDir(''));
    WizardForm.ReadyMemo.Lines.Add('      Backups folder: ' + GetBackupsDir(''));
    WizardForm.ReadyMemo.Lines.Add('      Web UI port: ' + GetPort(''));
    if IsExternalDbChosen then
      WizardForm.ReadyMemo.Lines.Add('      Metadata database: external PostgreSQL (stored connection string)')
    else
      WizardForm.ReadyMemo.Lines.Add('      Metadata database: bundled PostgreSQL');
  end;
end;

{ Runs before the main PrepareToInstall stops the service, so a refused upgrade leaves the running
  version and its data untouched. The skipped pages' checks run here; a silent install checks
  everything in the main implementation. }
<event('PrepareToInstall')>
function UpgradePrepareToInstall(var NeedsRestart: Boolean): String;
begin
  Result := '';
  if IsUpgrade and not WizardSilent then
  begin
    Result := CheckChoices(True, False);
    if Result = '' then
      Result := CheckExternalDb;
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

  CreateExternalDbPage(PortPage.ID);

  KeepPreviousSettingsOnUpgrade;
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
    Problems := CheckChoices(False, True)
  else if CurPageID = ExternalDbPage.ID then
    Problems := CheckExternalDb;

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
    if Result = '' then
      Result := CheckExternalDb;
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

{ Inno Setup cannot roll back from the post-install step, so a failure keeps the files (the admin
  can read the logs and fix the cause), tells why, and makes Setup exit with ServiceFailedExitCode. }
procedure FailInstall(const Message: String);
begin
  ServiceFailure := Message;
  Log('Service setup failed: ' + Message);
  SuppressibleMsgBox(Message, mbCriticalError, MB_OK, IDOK);
end;

function IsServiceStarted: Boolean;
begin
  Result := ServiceFailure = '';
end;

function GetCustomSetupExitCode: Integer;
begin
  Result := 0;
  if ServiceFailure <> '' then
    Result := ServiceFailedExitCode;
end;

procedure InstallAndStartService;
var
  Params, Output, LogsHint: String;
  Waited: Integer;
  WasServiceInstalled: Boolean;
begin
  WasServiceInstalled := IsServiceInstalled;
  LogsHint := ' See ' + AddBackslash(GetDataDir('')) + 'logs\launcher.log and the Application event log.';

  if not StoreExternalDb(Output) then
  begin
    FailInstall('Storing the external metadata database connection string failed:' + #13#10 + Output);
    exit;
  end;

  WizardForm.StatusLabel.Caption := 'Registering the Databasus service...';
  Params := 'install -data-dir "' + GetDataDir('') + '" -backups-dir "' + GetBackupsDir('') + '"' +
    ' -listen-host ' + ListenHost + ' -port ' + GetPort('');
  if IsExternalDbChosen then
    Params := Params + ' -external-metadata-db-file "' + GetExternalDbTargetFile + '"';
  if RunAndCapture(ExpandConstant('{app}\databasus-launcher.exe'), Params, Output) <> 0 then
  begin
    FailInstall('Registering the Databasus service failed:' + #13#10 + Output);
    exit;
  end;

  WizardForm.StatusLabel.Caption := 'Starting the Databasus service (the first start takes up to a few minutes)...';
  if RunAndCapture(ExpandConstant('{sys}\sc.exe'), 'start {#ServiceName}', Output) <> 0 then
  begin
    if not WasServiceInstalled then
      RunAndCapture(ExpandConstant('{app}\databasus-launcher.exe'), 'uninstall', Output);
    FailInstall('Starting the Databasus service failed:' + #13#10 + Output + LogsHint);
    exit;
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
      exit;
    end;
    Sleep(2000);
    Waited := Waited + 2;
  end;
  Log('The Web UI answers on ' + GetWebUiUrl(''));
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssPostInstall then
    InstallAndStartService;
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

{ Runs after the "remove Databasus?" confirmation and before anything is removed, while the
  settings key still exists. }
procedure AskWhetherToDeleteData;
begin
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

procedure InitializeUninstallProgressForm;
begin
  AskWhetherToDeleteData;
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
