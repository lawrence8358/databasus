import { Spin } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';

import { type Database, DatabaseType, databaseApi } from '../../../entity/databases';
import { AvailabilityFeature, useAvailability } from '../../../entity/system';
import { LogicalBackupsComponent } from '../../backups/logical';
import { PhysicalBackupsComponent } from '../../backups/physical';
import { HealthckeckAttemptsComponent } from '../../healthcheck';
import { VerificationsComponent } from '../../verification/runs';
import { DatabaseConfigComponent } from './DatabaseConfigComponent';

interface Props {
  contentHeight: number;
  databaseId: string;
  onDatabaseChanged: (database: Database) => void;
  onDatabaseDeleted: () => void;
  isCanManageDBs: boolean;
}

export const DatabaseComponent = ({
  contentHeight,
  databaseId,
  onDatabaseChanged,
  onDatabaseDeleted,
  isCanManageDBs,
}: Props) => {
  const { t } = useTranslation();
  const [currentTab, setCurrentTab] = useState<'config' | 'backups' | 'verifications'>('backups');

  const [database, setDatabase] = useState<Database | undefined>();
  const [editDatabase, setEditDatabase] = useState<Database | undefined>();

  const scrollContainerRef = useRef<HTMLDivElement>(null);

  const [isHealthcheckVisible, setIsHealthcheckVisible] = useState(false);

  const handleHealthcheckVisibilityChange = (isVisible: boolean) => {
    setIsHealthcheckVisible(isVisible);
  };

  const isLogicalDatabase = database?.type === DatabaseType.POSTGRES_LOGICAL;
  const isPhysicalDatabase = database?.type === DatabaseType.POSTGRES_PHYSICAL;

  const loadSettings = () => {
    setDatabase(undefined);
    setEditDatabase(undefined);
    databaseApi.getDatabase(databaseId).then(setDatabase);
  };

  const { isFeatureAvailable } = useAvailability();

  useEffect(() => {
    loadSettings();
  }, [databaseId]);

  if (!database) {
    return <Spin />;
  }

  const isVerificationShown =
    isLogicalDatabase && isFeatureAvailable(AvailabilityFeature.VERIFICATION_AGENT);

  return (
    <div
      className="w-full overflow-y-auto"
      style={{ maxHeight: contentHeight }}
      ref={scrollContainerRef}
    >
      <div className="flex">
        <div
          className={`mr-2 cursor-pointer rounded-tl-md rounded-tr-md px-6 py-2 ${currentTab === 'config' ? 'bg-white dark:bg-gray-800' : 'bg-gray-200 dark:bg-gray-700'}`}
          onClick={() => setCurrentTab('config')}
        >
          {t('databases.tabs.config')}
        </div>

        <div
          className={`mr-2 cursor-pointer rounded-tl-md rounded-tr-md px-6 py-2 ${currentTab === 'backups' ? 'bg-white dark:bg-gray-800' : 'bg-gray-200 dark:bg-gray-700'}`}
          onClick={() => setCurrentTab('backups')}
        >
          {t('databases.tabs.backups')}
        </div>

        {isVerificationShown && (
          <div
            className={`mr-2 cursor-pointer rounded-tl-md rounded-tr-md px-6 py-2 ${currentTab === 'verifications' ? 'bg-white dark:bg-gray-800' : 'bg-gray-200 dark:bg-gray-700'}`}
            onClick={() => setCurrentTab('verifications')}
          >
            {t('databases.tabs.verifications')}
          </div>
        )}
      </div>

      {currentTab === 'config' && (
        <DatabaseConfigComponent
          database={database}
          setDatabase={setDatabase}
          onDatabaseChanged={onDatabaseChanged}
          onDatabaseDeleted={onDatabaseDeleted}
          editDatabase={editDatabase}
          setEditDatabase={setEditDatabase}
          isCanManageDBs={isCanManageDBs}
        />
      )}

      {currentTab === 'backups' && (
        <>
          <HealthckeckAttemptsComponent
            database={database}
            onVisibilityChange={handleHealthcheckVisibilityChange}
          />

          {isPhysicalDatabase ? (
            <PhysicalBackupsComponent
              database={database}
              isCanManageDBs={isCanManageDBs}
              isDirectlyUnderTab={!isHealthcheckVisible}
              scrollContainerRef={scrollContainerRef}
            />
          ) : (
            <LogicalBackupsComponent
              database={database}
              isCanManageDBs={isCanManageDBs}
              isDirectlyUnderTab={!isHealthcheckVisible}
              scrollContainerRef={scrollContainerRef}
              onNavigateToVerifications={() => setCurrentTab('verifications')}
            />
          )}
        </>
      )}

      {currentTab === 'verifications' && isVerificationShown && (
        <VerificationsComponent
          database={database}
          isCanManageDBs={isCanManageDBs}
          isDirectlyUnderTab={true}
          scrollContainerRef={scrollContainerRef}
        />
      )}
    </div>
  );
};
