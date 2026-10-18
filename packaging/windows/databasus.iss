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
;                   [/LANACCESS [/PFX=<file> [/PFXPASSWORDFILE=<file>] | /SELFSIGNED]] [/LOCALONLY]
;                   /LANACCESS opens the Web UI to the local network over HTTPS only and allows its port
;                   in Windows Firewall. The certificate comes from /PFX (its password is read from
;                   /PFXPASSWORDFILE, never from the command line), or a new self-signed one with
;                   /SELFSIGNED; with neither, an earlier valid certificate is kept or a self-signed one
;                   is generated. /LOCALONLY switches back to 127.0.0.1. An upgrade keeps the earlier
;                   choice and certificate unless one of them is given.
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
#define FirewallRuleName "Databasus Web UI"

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

[Registry]
; Read back by the uninstaller and by later installers (upgrade keeps these choices).
Root: HKLM; Subkey: "{#SettingsKey}"; Flags: uninsdeletekey
Root: HKLM; Subkey: "{#SettingsKey}"; ValueType: string; ValueName: "InstallDir"; ValueData: "{app}"
Root: HKLM; Subkey: "{#SettingsKey}"; ValueType: string; ValueName: "DataDir"; ValueData: "{code:GetDataDir}"
Root: HKLM; Subkey: "{#SettingsKey}"; ValueType: string; ValueName: "BackupsDir"; ValueData: "{code:GetBackupsDir}"
Root: HKLM; Subkey: "{#SettingsKey}"; ValueType: string; ValueName: "WebPort"; ValueData: "{code:GetPort}"
Root: HKLM; Subkey: "{#SettingsKey}"; ValueType: string; ValueName: "WebListenHost"; ValueData: "{code:GetListenHost}"
Root: HKLM; Subkey: "{#SettingsKey}"; ValueType: string; ValueName: "WebLanAccess"; ValueData: "{code:GetLanAccessFlag}"
Root: HKLM; Subkey: "{#SettingsKey}"; ValueType: string; ValueName: "ExternalMetadataDb"; ValueData: "{code:GetExternalDbFlag}"
Root: HKLM; Subkey: "{#SettingsKey}"; ValueType: string; ValueName: "Version"; ValueData: "{#AppVersion}"

[Run]
Filename: "{code:GetWebUiUrl}"; Description: "Open the Databasus Web UI"; Flags: shellexec postinstall skipifsilent nowait runasoriginaluser; Check: IsServiceStarted

[UninstallRun]
Filename: "{app}\databasus-launcher.exe"; Parameters: "uninstall"; Flags: runhidden waituntilterminated; RunOnceId: "RemoveService"
; Added by Setup when LAN access is on; deleting a rule that does not exist only fails quietly.
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""{#FirewallRuleName}"""; Flags: runhidden waituntilterminated; RunOnceId: "RemoveFirewallRule"

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

{ ---------- LAN access: the choice the settings below depend on ---------- }

var
  LanAccessCheckBox: TNewCheckBox;

function IsLanAccessChosen: Boolean;
begin
  Result := LanAccessCheckBox.Checked;
end;

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

{ An empty host makes the service listen on every interface (IPv4 and IPv6). }
function GetListenHost(Param: String): String;
begin
  if IsLanAccessChosen then
    Result := ''
  else
    Result := ListenHost;
end;

{ With LAN access the Web UI speaks HTTPS only; localhost is one of the self-signed certificate's names. }
function GetWebUiUrl(Param: String): String;
begin
  if IsLanAccessChosen then
    Result := 'https://localhost:' + GetPort('') + '/'
  else
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

{ ---------- LAN access and HTTPS ---------- }

{ With LAN access the service listens on every interface and serves HTTPS only. The launcher
  stores the certificate under the data folder (tls-setup), restricted to the service account,
  SYSTEM and Administrators. A PFX password never goes on a command line (visible to other users,
  written to the Setup log): it reaches the launcher through an environment variable of this
  process. The firewall rule is Setup's own, found by name, and the uninstaller deletes it. }
const
  PfxPasswordEnvVariable = 'DATABASUS_PFX_PASSWORD';

var
  LanAccessPage: TWizardPage;
  CertKeepRadio: TNewRadioButton;
  CertSelfSignedRadio: TNewRadioButton;
  CertPfxRadio: TNewRadioButton;
  PfxFileEdit: TNewEdit;
  PfxBrowseButton: TNewButton;
  PfxPasswordEdit: TPasswordEdit;
  { Summary line of a usable certificate already in the data folder, '' when there is none. }
  ExistingCertificate: String;
  PfxPasswordFileProblem: String;

function IsLanSwitchGiven(const Switch: String): Boolean;
var
  I: Integer;
begin
  Result := False;
  for I := 1 to ParamCount do
    if CompareText(ParamStr(I), Switch) = 0 then
      Result := True;
end;

function GetLanAccessFlag(Param: String): String;
begin
  if IsLanAccessChosen then
    Result := '1'
  else
    Result := '0';
end;

function RunTlsSetup(const Launcher, Params: String; var Output: String): Integer;
begin
  if CertPfxRadio.Checked then
    SetEnvironmentVariable(PfxPasswordEnvVariable, PfxPasswordEdit.Text);
  try
    Result := RunAndCapture(Launcher, 'tls-setup -data-dir "' + GetDataDir('') + '" ' + Params, Output);
  finally
    SetEnvironmentVariable(PfxPasswordEnvVariable, '');
  end;
end;

procedure FindExistingCertificate;
var
  Output: String;
begin
  ExistingCertificate := '';
  if RunTlsSetup(GetLauncherForChecks, '-source existing', Output) = 0 then
    ExistingCertificate := Trim(Output);
end;

procedure UpdateLanAccessControls;
var
  IsOn: Boolean;
begin
  IsOn := IsLanAccessChosen;
  CertKeepRadio.Enabled := IsOn and (ExistingCertificate <> '');
  CertSelfSignedRadio.Enabled := IsOn;
  CertPfxRadio.Enabled := IsOn;
  PfxFileEdit.Enabled := IsOn and CertPfxRadio.Checked;
  PfxBrowseButton.Enabled := PfxFileEdit.Enabled;
  PfxPasswordEdit.Enabled := PfxFileEdit.Enabled;
end;

{ Keeps a usable stored certificate by default, and otherwise generates a self-signed one. }
procedure SelectDefaultCertificateSource;
begin
  if CertKeepRadio.Checked and (ExistingCertificate = '') then
    CertKeepRadio.Checked := False;
  if not (CertKeepRadio.Checked or CertSelfSignedRadio.Checked or CertPfxRadio.Checked) then
  begin
    if ExistingCertificate <> '' then
      CertKeepRadio.Checked := True
    else
      CertSelfSignedRadio.Checked := True;
  end;
end;

{ The data folder is chosen on an earlier page, so the stored certificate is looked up here. }
procedure LanAccessPageActivate(Sender: TWizardPage);
begin
  FindExistingCertificate;
  if ExistingCertificate <> '' then
    CertKeepRadio.Caption := 'Keep the current certificate (the Setup log shows its details)'
  else
    CertKeepRadio.Caption := 'Keep the current certificate (none is stored yet)';
  SelectDefaultCertificateSource;
  UpdateLanAccessControls;
end;

procedure LanAccessControlClick(Sender: TObject);
begin
  UpdateLanAccessControls;
end;

procedure PfxBrowseButtonClick(Sender: TObject);
var
  FileName: String;
begin
  FileName := PfxFileEdit.Text;
  if GetOpenFileName('Select the PFX certificate file', FileName, '',
    'PFX certificates (*.pfx;*.p12)|*.pfx;*.p12|All files|*.*', 'pfx') then
    PfxFileEdit.Text := FileName;
end;

{ Silent mode has no page to activate: /PFX imports a file, /SELFSIGNED generates a new
  certificate, and otherwise a usable stored certificate is kept. }
procedure ChooseSilentCertificateSource;
var
  PasswordFile: String;
  Lines: TArrayOfString;
begin
  if not IsLanAccessChosen then
    exit;
  FindExistingCertificate;
  PfxFileEdit.Text := ExpandConstant('{param:PFX|}');
  if PfxFileEdit.Text <> '' then
  begin
    CertPfxRadio.Checked := True;
    PasswordFile := ExpandConstant('{param:PFXPASSWORDFILE|}');
    if PasswordFile <> '' then
    begin
      if LoadStringsFromFile(PasswordFile, Lines) then
      begin
        if GetArrayLength(Lines) > 0 then
          PfxPasswordEdit.Text := Lines[0];
      end
      else
        PfxPasswordFileProblem := 'The PFX password file ' + PasswordFile + ' cannot be read.';
    end;
  end
  else if IsLanSwitchGiven('/SELFSIGNED') then
    CertSelfSignedRadio.Checked := True;
  SelectDefaultCertificateSource;
end;

{ Returns the problem in words for the person running Setup, or '' when LAN access is off or the
  chosen certificate can be used. Stores nothing. }
function CheckLanAccess: String;
var
  Output: String;
begin
  Result := '';
  if not IsLanAccessChosen then
    exit;

  if CertKeepRadio.Checked and (ExistingCertificate = '') then
    Result := 'No usable HTTPS certificate is stored in the data folder. Generate or import one.'
  else if CertPfxRadio.Checked then
  begin
    if PfxPasswordFileProblem <> '' then
      Result := PfxPasswordFileProblem
    else if Trim(PfxFileEdit.Text) = '' then
      Result := 'Choose the PFX certificate file to import.'
    else if not FileExists(Trim(PfxFileEdit.Text)) then
      Result := 'The PFX certificate file ' + Trim(PfxFileEdit.Text) + ' does not exist.'
    else if RunTlsSetup(GetLauncherForChecks,
      '-source pfx -check-only -pfx "' + Trim(PfxFileEdit.Text) + '"', Output) <> 0 then
    begin
      Result := Trim(Output);
      if Result = '' then
        Result := 'The PFX certificate file cannot be used.';
    end;
  end;
  if Result <> '' then
    Result := Result + #13#10 + #13#10 + 'Nothing has been installed.';
end;

function RunNetsh(const Params: String; var Output: String): Integer;
begin
  Result := RunAndCapture(ExpandConstant('{sys}\netsh.exe'), 'advfirewall firewall ' + Params, Output);
end;

{ Stores the certificate and opens the port in Windows Firewall, or, with LAN access off, removes
  the rule an earlier install added. A stored certificate stays in the data folder either way. }
function ApplyLanAccess(var Failure: String): Boolean;
var
  Source, Output: String;
begin
  Result := True;
  Failure := '';
  WizardForm.StatusLabel.Caption := 'Configuring network access...';
  { Fails when there is no rule; it is added again below with the current port. }
  RunNetsh('delete rule name="{#FirewallRuleName}"', Output);
  if not IsLanAccessChosen then
    exit;

  if CertPfxRadio.Checked then
    Source := '-source pfx -pfx "' + Trim(PfxFileEdit.Text) + '"'
  else if CertSelfSignedRadio.Checked then
    Source := '-source self-signed'
  else
    Source := '-source existing';
  if RunTlsSetup(ExpandConstant('{app}\databasus-launcher.exe'), Source, Output) <> 0 then
  begin
    Failure := 'Storing the HTTPS certificate failed:' + #13#10 + Output;
    Result := False;
    exit;
  end;
  Log('HTTPS certificate: ' + Trim(Output));

  if RunNetsh('add rule name="{#FirewallRuleName}" dir=in action=allow protocol=TCP localport=' +
    GetPort('') + ' program="' + ExpandConstant('{app}\databasus.exe') + '" profile=domain,private enable=yes',
    Output) <> 0 then
  begin
    Failure := 'Allowing port ' + GetPort('') + ' in Windows Firewall failed:' + #13#10 + Output;
    Result := False;
  end;
end;

procedure CreateLanAccessPage(const AfterPageId: Integer);
var
  Description, CertificateLabel, PasswordLabel: TNewStaticText;
begin
  LanAccessPage := CreateCustomPage(AfterPageId,
    'Network access', 'Who should be able to open the Databasus Web UI?');
  LanAccessPage.OnActivate := @LanAccessPageActivate;

  LanAccessCheckBox := TNewCheckBox.Create(LanAccessPage);
  LanAccessCheckBox.Parent := LanAccessPage.Surface;
  LanAccessCheckBox.Width := LanAccessPage.SurfaceWidth;
  LanAccessCheckBox.Caption := 'Allow other computers on the local network (HTTPS only)';
  LanAccessCheckBox.OnClick := @LanAccessControlClick;

  Description := TNewStaticText.Create(LanAccessPage);
  Description.Parent := LanAccessPage.Surface;
  Description.Top := LanAccessCheckBox.Top + LanAccessCheckBox.Height + ScaleY(4);
  Description.WordWrap := True;
  Description.Width := LanAccessPage.SurfaceWidth;
  Description.Caption :=
    'The Web UI then listens on every network adapter without plain HTTP, and Setup allows its port in ' +
    'Windows Firewall for domain and private networks. Unchecked, only this computer can open it.';

  CertificateLabel := TNewStaticText.Create(LanAccessPage);
  CertificateLabel.Parent := LanAccessPage.Surface;
  CertificateLabel.Top := Description.Top + Description.Height + ScaleY(10);
  CertificateLabel.Caption := 'HTTPS certificate:';

  CertKeepRadio := TNewRadioButton.Create(LanAccessPage);
  CertKeepRadio.Parent := LanAccessPage.Surface;
  CertKeepRadio.Top := CertificateLabel.Top + CertificateLabel.Height + ScaleY(4);
  CertKeepRadio.Width := LanAccessPage.SurfaceWidth;
  CertKeepRadio.Caption := 'Keep the current certificate';
  CertKeepRadio.OnClick := @LanAccessControlClick;

  CertSelfSignedRadio := TNewRadioButton.Create(LanAccessPage);
  CertSelfSignedRadio.Parent := LanAccessPage.Surface;
  CertSelfSignedRadio.Top := CertKeepRadio.Top + CertKeepRadio.Height + ScaleY(3);
  CertSelfSignedRadio.Width := LanAccessPage.SurfaceWidth;
  CertSelfSignedRadio.Caption := 'Generate a self-signed certificate (browsers warn until it is trusted)';
  CertSelfSignedRadio.OnClick := @LanAccessControlClick;

  CertPfxRadio := TNewRadioButton.Create(LanAccessPage);
  CertPfxRadio.Parent := LanAccessPage.Surface;
  CertPfxRadio.Top := CertSelfSignedRadio.Top + CertSelfSignedRadio.Height + ScaleY(3);
  CertPfxRadio.Width := LanAccessPage.SurfaceWidth;
  CertPfxRadio.Caption := 'Import a PFX file, for example one issued by your organization:';
  CertPfxRadio.OnClick := @LanAccessControlClick;

  PfxBrowseButton := TNewButton.Create(LanAccessPage);
  PfxBrowseButton.Parent := LanAccessPage.Surface;
  PfxBrowseButton.Width := ScaleX(75);
  PfxBrowseButton.Height := ScaleY(23);
  PfxBrowseButton.Top := CertPfxRadio.Top + CertPfxRadio.Height + ScaleY(4);
  PfxBrowseButton.Left := LanAccessPage.SurfaceWidth - PfxBrowseButton.Width;
  PfxBrowseButton.Caption := 'Browse...';
  PfxBrowseButton.OnClick := @PfxBrowseButtonClick;

  PfxFileEdit := TNewEdit.Create(LanAccessPage);
  PfxFileEdit.Parent := LanAccessPage.Surface;
  PfxFileEdit.Left := ScaleX(16);
  PfxFileEdit.Top := PfxBrowseButton.Top + ScaleY(1);
  PfxFileEdit.Width := PfxBrowseButton.Left - PfxFileEdit.Left - ScaleX(8);

  PasswordLabel := TNewStaticText.Create(LanAccessPage);
  PasswordLabel.Parent := LanAccessPage.Surface;
  PasswordLabel.Left := ScaleX(16);
  PasswordLabel.Top := PfxBrowseButton.Top + PfxBrowseButton.Height + ScaleY(9);
  PasswordLabel.Caption := 'PFX password:';

  PfxPasswordEdit := TPasswordEdit.Create(LanAccessPage);
  PfxPasswordEdit.Parent := LanAccessPage.Surface;
  PfxPasswordEdit.Left := PfxFileEdit.Left + ScaleX(90);
  PfxPasswordEdit.Top := PasswordLabel.Top - ScaleY(3);
  PfxPasswordEdit.Width := PfxFileEdit.Width - ScaleX(90);

  LanAccessCheckBox.Checked := IsLanSwitchGiven('/LANACCESS') or
    ((GetPreviousSetting('WebLanAccess', '0') = '1') and not IsLanSwitchGiven('/LOCALONLY'));
  UpdateLanAccessControls;
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
    'By default the Web UI listens on 127.0.0.1 only, so it is reachable from this computer and nowhere else. ' +
    'The next page can open it to the local network over HTTPS.');
  PortPage.Add('Port:', False);
  PortPage.Values[0] := ExpandConstant('{param:PORT|' + GetPreviousSetting('WebPort', '{#DefaultPort}') + '}');

  CreateExternalDbPage(PortPage.ID);
  CreateLanAccessPage(PortPage.ID);
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
    Problems := CheckExternalDb
  else if CurPageID = LanAccessPage.ID then
    Problems := CheckLanAccess;

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
    if Result = '' then
    begin
      ChooseSilentCertificateSource;
      Result := CheckLanAccess;
    end;
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
    Request.Open('GET', GetWebUiUrl('') + 'api/v1/system/version', False);
    { This only asks whether the local service is up; a self-signed certificate is expected. }
    if IsLanAccessChosen then
      Request.Option[4] := $3300; { WinHttpRequestOption_SslErrorIgnoreFlags: ignore all }
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

  if not ApplyLanAccess(Output) then
  begin
    FailInstall(Output);
    exit;
  end;

  WizardForm.StatusLabel.Caption := 'Registering the Databasus service...';
  Params := 'install -data-dir "' + GetDataDir('') + '" -backups-dir "' + GetBackupsDir('') + '"' +
    ' -listen-host "' + GetListenHost('') + '" -port ' + GetPort('');
  if IsLanAccessChosen then
    Params := Params + ' -https';
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
