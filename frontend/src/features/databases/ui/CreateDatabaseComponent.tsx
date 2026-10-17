import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import {
  type LogicalBackupConfig,
  logicalBackupConfigApi,
  logicalBackupsApi,
} from '../../../entity/backups/logical';
import {
  type PhysicalBackupConfig,
  physicalBackupConfigApi,
  physicalBackupsApi,
} from '../../../entity/backups/physical';
import {
  type Database,
  DatabaseType,
  Period,
  databaseApi,
  initializeDatabaseTypeData,
  isPostgresType,
} from '../../../entity/databases';
import { AvailabilityFeature, useAvailability } from '../../../entity/system';
import { translateApiError } from '../../../shared/i18n';
import { EditLogicalBackupConfigComponent } from '../../backups/logical';
import { EditPhysicalBackupConfigComponent } from '../../backups/physical';
import { ChoosePostgresBackupTypeComponent } from './edit/ChoosePostgresBackupTypeComponent';
import { CreateReadOnlyComponent } from './edit/CreateReadOnlyComponent';
import { EditDatabaseBaseInfoComponent } from './edit/EditDatabaseBaseInfoComponent';
import { EditDatabaseNotifiersComponent } from './edit/EditDatabaseNotifiersComponent';
import { EditDatabaseSpecificDataComponent } from './edit/EditDatabaseSpecificDataComponent';

interface Props {
  workspaceId: string;
  onCreated: (databaseId: string) => void;
  onClose: () => void;
  onConnectionErrorChange?: (hasConnectionError: boolean) => void;
  onWideStepChange?: (isWideStep: boolean) => void;
}

const createInitialDatabase = (workspaceId: string): Database =>
  ({
    id: undefined as unknown as string,
    name: '',
    workspaceId,
    storePeriod: Period.MONTH,

    type: DatabaseType.POSTGRES_LOGICAL,

    storage: {} as unknown as Storage,

    notifiers: [],
    sendNotificationsOn: [],
  }) as Database;

export const CreateDatabaseComponent = ({
  workspaceId,
  onCreated,
  onClose,
  onConnectionErrorChange,
  onWideStepChange,
}: Props) => {
  const { t } = useTranslation();
  const [isCreating, setIsCreating] = useState(false);
  const [backupConfig, setBackupConfig] = useState<LogicalBackupConfig | undefined>();
  const [physicalBackupConfig, setPhysicalBackupConfig] = useState<
    PhysicalBackupConfig | undefined
  >();
  const [database, setDatabase] = useState<Database>(createInitialDatabase(workspaceId));

  const [step, setStep] = useState<
    | 'base-info'
    | 'postgres-backup-type'
    | 'db-settings'
    | 'create-readonly-user'
    | 'backup-config'
    | 'notifiers'
  >('base-info');

  const isPhysical = database.type === DatabaseType.POSTGRES_PHYSICAL;
  const isPostgres = isPostgresType(database.type);

  const { isFeatureAvailable } = useAvailability();

  // The read-only user step explains its choice in long sentences and puts three buttons in one
  // row, which the narrow wizard width cannot hold.
  useEffect(() => {
    onWideStepChange?.(step === 'create-readonly-user');
  }, [step]);

  const createDatabase = async (database: Database) => {
    if (isPhysical ? !physicalBackupConfig : !backupConfig) return;

    setIsCreating(true);

    try {
      const createdDatabase = await databaseApi.createDatabase(database);
      setDatabase({ ...createdDatabase });

      if (isPhysical && physicalBackupConfig) {
        physicalBackupConfig.databaseId = createdDatabase.id;
        await physicalBackupConfigApi.savePhysicalBackupConfig(physicalBackupConfig);

        if (physicalBackupConfig.isBackupsEnabled) {
          await physicalBackupsApi.triggerPhysicalBackup(createdDatabase.id, 'auto');
        }
      } else if (backupConfig) {
        backupConfig.databaseId = createdDatabase.id;
        await logicalBackupConfigApi.saveBackupConfig(backupConfig);

        if (backupConfig.isBackupsEnabled) {
          await logicalBackupsApi.makeBackup(createdDatabase.id);
        }
      }

      onCreated(createdDatabase.id);
      onClose();
    } catch (error) {
      alert(translateApiError(error, t));
    }

    setIsCreating(false);
  };

  const isPostgresBackupTypeChoosable = isFeatureAvailable(AvailabilityFeature.PHYSICAL_BACKUP);

  if (step === 'base-info') {
    return (
      <div>
        <EditDatabaseBaseInfoComponent
          database={database}
          isShowName
          isShowEngine
          isSaveToApi={false}
          saveButtonText={t('common.actions.continue')}
          onCancel={() => onClose()}
          onSaved={(db) => {
            const initializedDb = initializeDatabaseTypeData(db);
            setDatabase({ ...initializedDb });
            setStep(
              isPostgresType(db.type) && isPostgresBackupTypeChoosable
                ? 'postgres-backup-type'
                : 'db-settings',
            );
          }}
        />
      </div>
    );
  }

  if (step === 'postgres-backup-type') {
    return (
      <ChoosePostgresBackupTypeComponent
        database={database}
        saveButtonText={t('common.actions.continue')}
        onBack={() => setStep('base-info')}
        onSelected={(type) => {
          const initializedDb = initializeDatabaseTypeData({ ...database, type });
          setDatabase({ ...initializedDb });
          setStep('db-settings');
        }}
      />
    );
  }

  if (step === 'db-settings') {
    return (
      <EditDatabaseSpecificDataComponent
        database={database}
        isShowCancelButton={false}
        onCancel={() => onClose()}
        isShowBackButton
        onBack={() =>
          setStep(
            isPostgres && isPostgresBackupTypeChoosable ? 'postgres-backup-type' : 'base-info',
          )
        }
        saveButtonText={t('common.actions.continue')}
        isSaveToApi={false}
        onConnectionErrorChange={onConnectionErrorChange}
        onSaved={(database) => {
          setDatabase({ ...database });
          setStep('create-readonly-user');
        }}
      />
    );
  }

  if (step === 'create-readonly-user') {
    return (
      <CreateReadOnlyComponent
        database={database}
        onReadOnlyUserUpdated={(database) => {
          setDatabase({ ...database });
          setStep('backup-config');
        }}
        onGoBack={() => setStep('db-settings')}
        onSkipped={() => setStep('backup-config')}
        onReadOnlyUserNotSuggested={() => setStep('backup-config')}
      />
    );
  }

  if (step === 'backup-config') {
    if (isPhysical) {
      return (
        <EditPhysicalBackupConfigComponent
          database={database}
          initialConfig={physicalBackupConfig}
          isShowCancelButton={false}
          onCancel={() => onClose()}
          isShowBackButton
          onBack={() => setStep('db-settings')}
          saveButtonText={t('common.actions.continue')}
          isSaveToApi={false}
          onSaved={(physicalBackupConfig) => {
            setPhysicalBackupConfig(physicalBackupConfig);
            setStep('notifiers');
          }}
        />
      );
    }

    return (
      <EditLogicalBackupConfigComponent
        database={database}
        isShowCancelButton={false}
        onCancel={() => onClose()}
        isShowBackButton
        onBack={() => setStep('db-settings')}
        saveButtonText={t('common.actions.continue')}
        isSaveToApi={false}
        onSaved={(backupConfig) => {
          setBackupConfig(backupConfig);
          setStep('notifiers');
        }}
      />
    );
  }

  if (step === 'notifiers') {
    if (isCreating) {
      return <div>{t('databases.create.creating')}</div>;
    }

    return (
      <EditDatabaseNotifiersComponent
        database={database}
        isShowCancelButton={false}
        workspaceId={workspaceId}
        onCancel={() => onClose()}
        isShowBackButton
        onBack={() => setStep('backup-config')}
        isShowSaveOnlyForUnsaved={false}
        saveButtonText={t('databases.create.complete')}
        isSaveToApi={false}
        onSaved={(database) => {
          if (isCreating) return;

          setDatabase({ ...database });
          createDatabase(database);
        }}
      />
    );
  }
};
