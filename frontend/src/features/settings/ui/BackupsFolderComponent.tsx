import { App, Button, Input } from 'antd';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import { type BackupsFolder, systemApi } from '../../../entity/system';
import { translateApiError } from '../../../shared/i18n';

// eslint-disable-next-line i18next/no-literal-string -- an example Windows path
const NEW_FOLDER_PLACEHOLDER = 'D:\\DatabasusBackups';

export function BackupsFolderComponent() {
  const { t } = useTranslation();
  const { message } = App.useApp();
  const [backupsFolder, setBackupsFolder] = useState<BackupsFolder | undefined>(undefined);
  const [newFolder, setNewFolder] = useState('');
  const [isChanging, setIsChanging] = useState(false);
  const [changeError, setChangeError] = useState<unknown>(undefined);

  const loadBackupsFolder = async () => {
    try {
      setBackupsFolder(await systemApi.getBackupsFolder());
    } catch (error: unknown) {
      message.error(translateApiError(error, t));
    }
  };

  const changeBackupsFolder = async () => {
    setIsChanging(true);
    setChangeError(undefined);

    try {
      const changedBackupsFolder = await systemApi.changeBackupsFolder(newFolder.trim());
      setBackupsFolder(changedBackupsFolder);
      setNewFolder('');
      message.success(
        t('settings.backupsFolder.changed', { folder: changedBackupsFolder.currentFolder }),
      );
    } catch (error: unknown) {
      setChangeError(error);
    } finally {
      setIsChanging(false);
    }
  };

  useEffect(() => {
    loadBackupsFolder();
  }, []);

  if (!backupsFolder?.isChangeable) {
    return null;
  }

  const isNewFolderEntered =
    newFolder.trim() !== '' && newFolder.trim() !== backupsFolder.currentFolder;

  return (
    <section className="my-8 max-w-2xl text-sm">
      <h2 className="mb-3 text-xl font-bold dark:text-white">
        {t('settings.backupsFolder.title')}
      </h2>

      <div className="text-gray-500 dark:text-gray-400">
        {t('settings.backupsFolder.description')}
      </div>

      <div className="mt-3 font-medium text-gray-900 dark:text-white">
        {t('settings.backupsFolder.currentFolder')}
      </div>
      <div className="mt-1 rounded-md border border-gray-300 bg-gray-50 px-3 py-2 !font-mono break-all text-gray-700 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-200">
        {backupsFolder.currentFolder}
      </div>

      {backupsFolder.previousFolders.length > 0 && (
        <div className="mt-3 text-gray-500 dark:text-gray-400">
          <div>{t('settings.backupsFolder.previousFolders')}</div>
          <ul className="mt-1 list-disc pl-5">
            {backupsFolder.previousFolders.map((previousFolder) => (
              <li key={previousFolder} className="!font-mono break-all">
                {previousFolder}
              </li>
            ))}
          </ul>
        </div>
      )}

      <div className="mt-4 font-medium text-gray-900 dark:text-white">
        {t('settings.backupsFolder.newFolder')}
      </div>
      <Input
        className="mt-1 !font-mono"
        value={newFolder}
        placeholder={NEW_FOLDER_PLACEHOLDER}
        status={changeError ? 'error' : undefined}
        onChange={(event) => {
          setNewFolder(event.target.value);
          setChangeError(undefined);
        }}
        onPressEnter={() => isNewFolderEntered && !isChanging && changeBackupsFolder()}
        disabled={isChanging}
      />

      {changeError !== undefined && (
        <div className="mt-1 text-red-600 dark:text-red-400">
          {translateApiError(changeError, t)}
        </div>
      )}

      <div className="mt-1 text-gray-500 dark:text-gray-400">
        {t('settings.backupsFolder.rules')}
      </div>

      {isNewFolderEntered && (
        <Button
          className="mt-3"
          type="primary"
          onClick={changeBackupsFolder}
          loading={isChanging}
          disabled={isChanging}
        >
          {t('settings.backupsFolder.change')}
        </Button>
      )}
    </section>
  );
}
