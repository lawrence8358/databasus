import type { en } from './en';

export const zh: typeof en = {
  language: {
    nativeName: '繁體中文',
  },
  common: {
    actions: {
      add: '新增',
      back: '返回',
      cancel: '取消',
      close: '關閉',
      confirm: '確認',
      continue: '繼續',
      copy: '複製',
      create: '建立',
      delete: '刪除',
      download: '下載',
      edit: '編輯',
      ok: '確定',
      refresh: '重新整理',
      remove: '移除',
      retry: '重試',
      save: '儲存',
    },
    answers: {
      no: '否',
      yes: '是',
    },
    fields: {
      host: '主機',
      name: '名稱',
      password: '密碼',
      port: '連線埠',
      status: '狀態',
      type: '類型',
      username: '使用者名稱',
    },
    states: {
      disabled: '已禁用',
      enabled: '已啟用',
      loading: '載入中…',
    },
    messages: {
      copiedToClipboard: '已複製到剪貼簿',
    },
    // Joins translated words into one list (Chinese uses 、)
    listSeparator: '、',
    confirmation: {
      title: '確認',
    },
  },
  app: {
    documentTitle: 'Databasus - 資料庫備份（PostgreSQL、MySQL、MongoDB）',
    languageControl: {
      title: '語言：{{language}}',
    },
    theme: {
      title: '主題：{{theme}}',
      light: '淺色',
      dark: '深色',
      system: '跟隨系統',
    },
    navigation: {
      databases: '資料庫',
      storages: '儲存空間',
      notifiers: '通知渠道',
      workspaceSettings: '設定',
      profile: '個人資料',
      databasusSettings: 'Databasus 設定',
      users: '使用者',
      docs: '說明文件',
      documentation: '說明文件',
      community: '社群',
      newVersionAvailable: '有新版本可用',
    },
    diskUsage: {
      title: '磁碟用量',
      hint: '在本機建立和還原備份需要足夠的磁碟空間。還原時需要的可用空間與備份大小相同。',
      compactSummary: '磁碟已用 {{used}} / {{total}} GB<lineBreak/>（{{percent}}%）',
      summary: '已用 {{used}} / {{total}} GB（{{percent}}%）',
    },
    workspaceSelection: {
      createShort: '建立',
      create: '建立工作區',
      selected: '目前工作區',
      placeholder: '選擇工作區',
      search: '搜尋工作區',
      notFound: '未找到工作區',
      createNew: '+ 新增工作區',
    },
    sponsorship: {
      link: '贊助',
      hint: 'Databasus 幫你保住了資料？歡迎支援 Databasus，讓它繼續免費、持續維護、保持獨立。',
    },
    starButton: {
      label: '在 GitHub 上加星號',
      ariaLabel: '在 GitHub 上為 lawrence8358/databasus 加星號',
    },
    clipboardPaste: {
      title: '從剪貼簿貼上',
      submit: '提交',
      description: '無法自動讀取剪貼簿，請在下方貼上內容。',
      placeholder: '在此貼上連線字串',
    },
    auth: {
      accountsCheckFailed: '無法檢查該執行個體是否已有帳戶：{{error}}',
    },
    oauthCallback: {
      failedTitle: '認證失敗',
      codeNotFound: '未找到授權碼',
      stateMissing: '缺少 OAuth state 參數',
      invalidProvider: '無效的 OAuth 提供方',
      returnToSignIn: '返回登入',
      completing: '正在完成認證…',
    },
    oauthStorage: {
      configNotFound: '未找到 Google Drive 儲存空間設定',
      exchangeFailed: 'OAuth 授權碼交換失敗',
      unsupportedType: '該儲存空間類型不支援 OAuth',
      invalidState: 'OAuth state 參數無效',
      paramNotFound: '未找到 OAuth 參數。請確認重定向 URL 設定正確。',
      addStorageTitle: '新增儲存空間',
      storageDescription: '儲存空間是存放備份的位置（本機磁碟、S3 等）',
    },
  },
  errors: {
    unknown: '出了點問題，請重試。',
    networkUnreachable: '無法連線 Databasus 伺服器。請檢查網路連線後重試。',
    requestFailed: '伺服器未能完成請求（HTTP {{status}}）。請稍後重試。',
    signInCodeIncorrect: '驗證碼不正確。',
    pendingSignInNotUsable: '這次登入無法繼續，請重新登入。',
    signInCodeNotSent: '執行個體無法傳送登入驗證碼，請稍後重試或聯絡管理員。',
    tooManySignInCodes: '該帳戶在過去一小時內請求的登入驗證碼過多，請稍後重試。',
    signInCodeResentTooSoon: '驗證碼傳送不到一分鐘，請稍等一分鐘後再請求新的驗證碼。',
    rateLimitExceeded: '嘗試次數過多，請稍後重試。',
    emailNotConfigured: '執行個體沒有設定郵件伺服器。',
    adminEmailMissing: '你的帳號沒有可以接收測試郵件的有效電子郵件地址。',
  },
  status: {
    logicalBackup: {
      inProgress: '進行中',
      completed: '成功',
      failed: '失敗',
      deleted: '已刪除',
      canceled: '已取消',
    },
    physicalBackup: {
      inProgress: '進行中',
      completed: '成功',
      error: '錯誤',
      chainBroken: '備份鏈中斷',
      canceled: '已取消',
    },
    restoreVerification: {
      verifiedSuccessful: '已驗證',
      verificationFailed: '驗證失敗',
    },
    restore: {
      inProgress: '進行中',
      completed: '成功',
      failed: '失敗',
      canceled: '已取消',
    },
    verification: {
      pending: '排隊中',
      running: '執行中',
      completed: '成功',
      failed: '失敗',
      canceled: '已取消',
    },
    health: {
      available: '可用',
      unavailable: '不可用',
    },
    agent: {
      neverSeen: '從未連線',
      online: '線上',
      stale: '回應延遲',
      offline: '離線',
    },
  },
  intervals: {
    types: {
      hourly: '每小時',
      daily: '每天',
      weekly: '每週',
      monthly: '每月',
      cron: 'Cron',
    },
    weekdays: {
      mon: '週一',
      tue: '週二',
      wed: '週三',
      thu: '週四',
      fri: '週五',
      sat: '週六',
      sun: '週日',
    },
    periods: {
      day: '1 天',
      week: '1 周',
      month: '1 個月',
      threeMonths: '3 個月',
      sixMonths: '6 個月',
      year: '1 年',
      twoYears: '2 年',
      threeYears: '3 年',
      fourYears: '4 年',
      fiveYears: '5 年',
      forever: '永久',
    },
  },
  auditLogs: {
    title: '稽核記錄',
    columns: {
      user: '使用者',
      message: '事件',
      workspace: '工作區',
      created: '時間',
    },
    systemUser: '系統',
    loadedOfTotal: '已載入：{{loaded}} / {{total}}',
    empty: '沒有稽核記錄。',
    loadingMore: '正在載入更多記錄…',
    allLoaded: '已載入全部記錄（共 {{count}} 條）',
  },
  settings: {
    title: 'Databasus 設定',
    externalRegistrations: {
      title: '允許外部註冊',
      description:
        '開啟後，新使用者可以自行註冊 Databasus 帳號。關閉後，新使用者只能透過邀請註冊。',
    },
    memberInvitations: {
      title: '允許成員邀請',
      description:
        '開啟後，現有成員可以邀請新使用者加入 Databasus。關閉後，只有管理員能邀請使用者。',
    },
    memberWorkspaceCreation: {
      title: '允許成員建立工作區',
      description:
        '開啟後，成員（非管理員使用者）可以建立新的工作區。關閉後，只有管理員能建立工作區。',
    },
    twoFactorAuth: {
      title: '啟用雙因素登入',
      description:
        '啟用後，輸入密碼之後還需要填寫傳送到帳號電子郵件的六位驗證碼。透過 Google 或 GitHub 登入不受此設定影響。',
      mailServerRequired: '這需要郵件伺服器，請先<docsLink>設定 SMTP</docsLink>。',
    },
    saving: '正在儲存…',
    saveChanges: '儲存變更',
    reset: '重置',
    updated: '設定已更新',
    readMore: '關於這些設定的詳細說明，請<docsLink>檢視文件</docsLink>',
    mailServer: {
      title: '郵件伺服器',
      configured: '郵件伺服器<status>已設定</status>。',
      notConfigured: '郵件伺服器<status>未設定</status>。',
      description: '它負責傳送邀請、密碼重置碼和登入驗證碼。郵件通知渠道有自己的 SMTP 設定。',
      docs: '連線郵件伺服器的方法見 <docsLink>SMTP 文件</docsLink>。',
      sendTestEmail: '傳送測試郵件',
      testEmailSent: '測試郵件已傳送至 {{email}}',
    },
    backupsFolder: {
      title: '備份資料夾',
      description:
        '本機儲存會把新的備份寫入這個資料夾。變更時不會搬動任何檔案：先前的備份留在原處，仍然可以還原。',
      currentFolder: '目前的資料夾',
      previousFolders: '以前使用的資料夾，裡面仍有較舊的備份：',
      newFolder: '新的資料夾',
      rules:
        '請選擇這台電腦本機固定磁碟上的資料夾，不能是磁碟根目錄、網路磁碟機、卸除式磁碟或 Windows、程式資料夾。資料夾不存在時 Databasus 會自動建立，並確認服務可以寫入。',
      change: '變更資料夾',
      changed: '新的備份現在會寫入 {{folder}}',
    },
    healthcheck: {
      title: '可用性檢查',
      openInNewTab: '點選在新分頁中開啟',
      copied: '可用性檢查位址已複製到剪貼簿',
      description: '透過這個位址監控 Databasus 系統是否可用',
    },
  },
  databases: {
    types: {
      postgresql: 'PostgreSQL',
      mysql: 'MySQL',
      mariadb: 'MariaDB',
      mongodb: 'MongoDB',
    },
    sslModes: {
      disable: '禁用',
      require: '必需',
      verifyCa: '驗證 CA',
      verifyFull: '完全驗證',
    },
    physicalBackupTypes: {
      full: '僅完整備份',
      fullIncremental: '完整 + 增量',
      fullIncrementalWalStream: '完整 + 增量 + WAL 串流傳輸',
    },
    list: {
      addDatabase: '新增資料庫',
      searchPlaceholder: '搜尋資料庫',
      noSearchResults: '沒有找到與“{{query}}”匹配的資料庫',
      description: '資料庫就是我們要備份的物件',
      backToList: '← 返回資料庫列表',
    },
    card: {
      storage: '儲存空間：<storageName/>',
      lastBackup: '上次備份：{{relativeTime}}',
      hasBackupError: '備份出錯',
    },
    tabs: {
      config: '設定',
      backups: '備份',
      verifications: '還原驗證',
    },
    create: {
      title: '新增要備份的資料庫',
      namePlaceholder: '我最愛的資料庫',
      creating: '正在建立資料庫…',
      complete: '完成',
      backupType: {
        title: '選擇備份類型',
        logical: {
          title: '邏輯備份',
          description: '推薦用於 50 GB 以下的資料庫，設定更簡單。',
        },
        physical: {
          title: '實體備份',
          description: '適用於 50 GB 以上的資料庫。支援時間點還原，RPO/RTO 更好，但需要額外設定。',
        },
      },
    },
    config: {
      namePlaceholder: '輸入名稱',
      removing: '正在刪除資料庫…',
      lastBackupError: {
        title: '上次備份出錯',
        errorLabel: '錯誤：',
        clearHint: '清除此錯誤（任選其一）：',
        clearByTestingConnection: '- 用下方按鈕測試連線（即使你已經修改了設定）；',
        clearByNextBackup: '- 等待下一次備份成功完成；',
      },
      sections: {
        database: '資料庫設定',
        backupConfig: '備份設定',
        healthcheck: '可用性檢查設定',
        restoreVerification: '還原驗證',
        notifiers: '通知渠道設定',
      },
      copyConfirmation: '確定要複製這個資料庫嗎？系統會建立一個設定相同的新資料庫。',
      removeConfirmation: '確定要刪除這個資料庫嗎？此操作無法復原。',
      copied: {
        title: '資料庫已複製',
        description: '已建立“{{name}}”',
      },
      connectionTestSucceeded: {
        title: '連線測試成功',
        description: '資料庫連線測試透過',
      },
    },
    actions: {
      testConnection: '測試連線',
      replace: '替換',
    },
    fields: {
      databaseType: '資料庫類型',
      version: '{{engine}} 版本',
      databaseName: '資料庫名',
      sslMode: 'SSL 模式',
      useHttps: '使用 HTTPS',
      cpuCount: 'CPU 數量',
      backupType: '備份類型',
      clientCertificate: '客戶端憑證',
      clientKey: '客戶端私鑰',
      serverCaCertificate: '伺服器 CA 憑證',
      includeSchemas: '包含的模式',
      excludeTables: '排除的表',
      skipUserMappings: '跳過使用者對應',
      excludeExtensions: '排除擴充套件',
      restoreOwnership: '還原所有者',
      restorePrivileges: '還原許可權',
      excludeEvents: '排除事件',
      galeraReplication: 'Galera 複製',
      useSrvConnection: '使用 SRV 連線',
      directConnection: '直接連線',
      authDatabase: '認證資料庫',
      excludeCollections: '排除的集合',
    },
    edit: {
      parseFromClipboard: '從剪貼簿解析',
      clipboardEmpty: '剪貼簿為空',
      clipboardReadFailed: '無法讀取剪貼簿。請檢查瀏覽器許可權。',
      connectionStringParsed: '已解析連線字串',
      connectionStringParsedWithoutPassword: '已解析連線字串，請手動輸入密碼。',
      connectionTestPassed: {
        title: '連線測試透過',
        description: '可以繼續下一步',
      },
      ipWhitelistHint:
        '如果資料庫啟用了 IP 白名單，請確保已將 Databasus 伺服器的 IP 加入允許列表。',
      localhostHint: '請<docsLink>閱讀這篇文件</docsLink>，了解如何備份本機資料庫',
      advancedSettings: '高階設定',
      placeholders: {
        host: '輸入 {{engine}} 主機',
        port: '輸入 {{engine}} 連線埠',
        username: '輸入 {{engine}} 使用者名稱',
        password: '輸入 {{engine}} 密碼',
        databaseName: '輸入 {{engine}} 資料庫名',
      },
      cpuCountTooltip:
        '備份和還原時使用的 CPU 核心數。數值越大，速度可能越快，但占用的資源也越多。',
      excludeTablesPlaceholder: '未排除任何表',
      excludeTablesTooltip: '不納入備份的表名。可以貼上以逗號或換行分隔的列表。',
      serverCaCertificateTooltip:
        '可選。填寫後，會用這個 CA 驗證伺服器憑證（verify-ca / verify-full）。',
      copyFailed: '複製失敗',
      postgresql: {
        supabaseHint: '請<docsLink>閱讀這篇文件</docsLink>，了解如何備份 Supabase 資料庫',
        includeSchemasPlaceholder: '所有模式（預設）',
        excludeTablesTooltip:
          "不納入備份的表。格式為 'tablename' 或 'schema.tablename'，支援 glob 模式（例如 'logs_*'）。可以貼上以逗號或換行分隔的列表。",
        skipUserMappingsTooltip:
          '還原時跳過使用者對應（CREATE USER MAPPING 語句）。如果備份所用的角色無法讀取對應中的認證資訊，請開啟此項：否則傾印出的對應會遺失選項（OPTIONS），導致 oracle_fdw 等 FDW 還原失敗。',
        excludeExtensionsTooltip:
          '還原時跳過擴充套件（CREATE EXTENSION 語句）。如果要還原到由服務商管理擴充套件的託管 PostgreSQL 服務，請開啟此項。',
        restoreOwnershipTooltip:
          '執行傾印中的 ALTER OWNER 語句，讓還原出的物件保留原來的所有者。連線使用者必須能分配這些角色，通常需要超級使用者。',
        restorePrivilegesTooltip:
          '執行傾印中的 GRANT 和 REVOKE 語句，讓還原出的物件保留原來的 ACL。連線使用者必須能向相關角色授權，通常需要超級使用者。',
        backupTypeTooltips: {
          full: '定期建立獨立的完整備份，每個備份都自成一體。',
          fullIncremental:
            '完整備份加上增量備份。增量備份只儲存自上一次備份以來的變化，體積更小，速度更快。',
          fullIncrementalWalStream:
            '在完整和增量備份之上增加持續的 WAL 串流傳輸。只有這個選項支援時間點還原（PITR），但占用空間更多，還原也更慢。',
        },
      },
      mariadb: {
        skipEvents: '跳過事件',
        skipEventsTooltip:
          '備份時跳過資料庫事件。如果 MariaDB 伺服器上禁用了事件排程器，請開啟此項。',
        skipGaleraDisable: '還原時不禁用',
        skipGaleraDisableTooltip:
          '預設情況下，Databasus 在還原時執行 SET SESSION wsrep_on=OFF，以避免 Galera writeset 大小錯誤。這需要 SUPER 許可權。如果託管服務商不授予 SUPER 許可權，請開啟此項跳過這一步，但大型還原可能會觸及 Galera 的 writeset 限制。',
      },
      mongodb: {
        srvTooltip:
          '連線 MongoDB Atlas 的 SRV 位址（mongodb+srv://）時開啟。SRV 連線不需要連線埠。',
        srvOverTunnelTooltip:
          '透過 SSH 隧道連線時不可用：SRV 會自行解析主機列表，繞過轉送的連線埠。',
        directConnectionTooltip:
          '直接連線單個伺服器，跳過副本集發現。適用於伺服器位於負載均衡器、代理或隧道之後的情況。',
        directConnectionOverTunnelTooltip:
          '啟用 SSH 隧道時強制開啟：隧道只暴露一個位址，因此無法使用副本集發現。',
        excludeCollectionsPlaceholder: '未排除任何集合',
        excludeCollectionsTooltip: '不納入備份的集合名。可以貼上以逗號或換行分隔的列表。',
      },
    },
    show: {
      galeraDisableSkipped: '還原時不禁用',
    },
    sshTunnel: {
      authTypes: {
        password: '密碼',
        privateKey: '私鑰',
      },
      title: 'SSH 隧道',
      tooltip:
        '用於封閉網路中的資料庫。Databasus 連線到下方的 SSH 主機，再由該主機透過上方填寫的主機和連線埠存取資料庫。',
      host: 'SSH 主機',
      hostTooltip:
        '上方的資料庫主機由 SSH 主機解析，而不是由 Databasus 解析。如果資料庫就執行在 SSH 主機上，請填寫 127.0.0.1。',
      port: 'SSH 連線埠',
      username: 'SSH 使用者名稱',
      usernamePlaceholder: '輸入 SSH 使用者名稱',
      authType: 'SSH 認證方式',
      credentials: 'SSH 認證資訊',
      password: 'SSH 密碼',
      passwordPlaceholder: '輸入 SSH 密碼',
      privateKey: 'SSH 私鑰',
      keyPassphrase: '金鑰口令',
      keyPassphrasePlaceholder: '僅用於加密的金鑰',
    },
    notifiers: {
      title: '通知渠道',
      description: '通知渠道是接收通知的地方（郵件、Slack、Telegram 等）',
      multipleHint: '可以選擇多個通知渠道，通知會傳送到所有選中的渠道。',
      selectPlaceholder: '選擇通知渠道',
      createNew: '新增通知渠道',
      addTitle: '新增通知渠道',
      notifyTo: '通知傳送到',
      empty: '未設定通知渠道',
    },
    readOnlyUser: {
      dialogTitle: '建立唯讀使用者',
      readOnly: {
        checking: '正在檢查唯讀使用者…',
        title: '為 Databasus 建立唯讀使用者？',
        description:
          '唯讀使用者是許可權受限的 {{databaseType}} 使用者，只能讀取資料庫中的資料，不能修改。備份推薦使用唯讀使用者，因為：',
        noWriteCredentials:
          '<bold>使用唯讀使用者，就完全不必儲存帶寫入許可權的認證資訊</bold>。即使在被入侵的最壞情況下，也沒有人能破壞你的資料。',
        create: '是，建立唯讀使用者',
        skipTitle: '跳過建立唯讀使用者？',
        skipQuestion: '確定要跳過建立唯讀使用者嗎？',
      },
      replicationOnly: {
        checking: '正在檢查複製專用使用者…',
        title: '為 Databasus 建立複製專用使用者？',
        description:
          '複製專用使用者是許可權受限的 {{databaseType}} 使用者，只能讀取資料庫中的資料，不能修改。備份推薦使用複製專用使用者，因為：',
        noWriteCredentials:
          '<bold>使用複製專用使用者，就完全不必儲存帶寫入許可權的認證資訊</bold>。即使在被入侵的最壞情況下，也沒有人能破壞你的資料。',
        create: '是，建立複製專用使用者',
        skipTitle: '跳過建立複製專用使用者？',
        skipQuestion: '確定要跳過建立複製專用使用者嗎？',
      },
      reasons: {
        preventsModifications: '它能防止備份期間意外修改資料',
        leastPrivilege: '它符合最小許可權原則',
        bestPractice: '這是安全方面的最佳實踐',
      },
      securityNote:
        'Databasus 採用企業級安全措施（<docsLink>詳見這裡</docsLink>），但不可能防範所有風險。',
      noWritePrivileges: '目前使用者<bold>沒有寫入許可權</bold>。',
      noWriteRoles: '目前使用者<bold>沒有帶寫入許可權的角色</bold>。',
      writePrivileges: '目前使用者擁有以下寫入許可權：',
      writeRoles: '目前使用者擁有以下帶寫入許可權的角色：',
      expand: '（展開）',
      collapse: '（收起）',
      skip: '跳過',
      skipRisk:
        '不建議用擁有全部許可權的使用者做備份，這可能帶來安全風險。Databasus 強烈建議你不要跳過這一步。',
      skipAdvice:
        '100% 的防護永遠做不到。哪怕被完全入侵的風險只有 0.01%，也值得防範。所以更穩妥的做法是使用唯讀使用者。',
      acceptRisks: '是，我接受風險',
      continueSecurely: '按安全方式繼續',
      forcedWalRotationUnavailable:
        '來源資料庫不允許向新使用者授予 pg_switch_wal() 的 EXECUTE 許可權，而持續 WAL 串流傳輸需要它來讓還原點儘量接近目前時刻。使用這些認證資訊，完整和增量備份可以正常工作。',
    },
    transfer: {
      title: '將資料庫轉移到其他工作區',
      targetWorkspace: '目標工作區',
      selectWorkspace: '選擇工作區',
      storage: {
        title: '儲存空間',
        transferExisting: '連同現有儲存空間一起轉移',
        selectFromTarget: '從目標工作區選擇儲存空間',
        blocked: '使用此儲存空間的資料庫：{{count}}。其他資料庫依賴該儲存空間，因此無法轉移。',
        canTransfer: '儲存空間可以轉移',
        selectPlaceholder: '選擇儲存空間',
        createNew: '新增儲存空間',
        addTitle: '新增儲存空間',
        description: '儲存空間是存放備份的位置（本機磁碟、S3、Google Drive 等）',
      },
      notifiers: {
        title: '通知渠道（可選）',
        transferWithDatabase: '隨資料庫一起轉移通知渠道',
        selectFromTarget: '從目標工作區選擇通知渠道',
        willTransfer: '將被轉移：',
        willNotTransfer: '不會被轉移（其他資料庫正在使用）：',
        usedByDatabases: '{{name}}（使用它的資料庫：{{count}}）',
        noneTransferred: '不會轉移任何通知渠道。轉移完成後，可以從目標工作區選擇通知渠道。',
        selectPlaceholder: '選擇通知渠道（可選）',
      },
      submit: '轉移',
      transferred: {
        title: '資料庫已轉移',
        description: '“{{name}}”已轉移到新工作區',
      },
    },
    connectionString: {
      errors: {
        empty: '連線字串為空',
        unrecognizedFormat: '無法識別連線字串的格式',
        hostMissing: '連線字串中缺少主機',
        usernameMissing: '連線字串中缺少使用者名稱',
        passwordMissing: '連線字串中缺少密碼',
        databaseMissing: '連線字串中缺少資料庫名',
        parseFailed: '無法解析連線字串',
        jdbc: {
          invalidFormat: 'JDBC 連線字串格式無效。應為：{{expectedFormat}}',
          queryParametersMissing: 'JDBC 連線字串缺少查詢參數（user 和 password）',
          usernameMissing: 'JDBC 連線字串中缺少使用者名稱（user 參數）',
          passwordMissing: 'JDBC 連線字串中缺少 password 參數',
          parseFailed: '無法解析 JDBC 連線字串',
        },
        keyValue: {
          hostMissing: '連線字串中缺少主機。請使用 host=hostname',
          usernameMissing: '連線字串中缺少使用者名稱。請使用 user=username',
          passwordMissing: '連線字串中缺少密碼。請使用 password=yourpassword',
          databaseMissing: '連線字串中缺少資料庫名。請使用 database=database',
          parseFailed: '無法解析鍵值對格式的連線字串',
        },
        libpq: {
          databaseMissing: '連線字串中缺少資料庫名。請使用 dbname=database',
          parseFailed: '無法解析 libpq 連線字串',
        },
      },
    },
    physicalConnectionErrors: {
      grantAsSuperuserNote: '用下面的命令授予該許可權（以超級使用者身份執行）。',
      applyAndRestartNote: '應用下面的修改，然後重啟 PostgreSQL。',
      pgHbaNoEntry: {
        title: '不允許從此主機進行復制',
        summary:
          'pg_hba.conf 中沒有允許 Databasus 建立複製連線的條目。實體備份透過複製連線傳輸，需要一條“host replication”規則，普通的“host all”規則並不涵蓋它。',
        addLineNote:
          '把下面這行加入 pg_hba.conf。第一個“all”匹配任意使用者，第二個匹配任意主機。如需限制存取，可以把使用者縮小到特定的複製角色，和/或把主機縮小到某個 CIDR 位址段。',
        reloadNote: '更新 pg_hba.conf 後，<bold>重新載入 PG 設定</bold>。',
        managedNote:
          '在託管 PostgreSQL（RDS / Azure / GCP）上無法編輯 pg_hba.conf，請改為在服務商控制檯中開啟外部存取或複製存取。',
      },
      badCredentials: {
        title: '使用者名稱或密碼錯誤',
        summary: 'PostgreSQL 拒絕了這些認證資訊。請檢查此資料庫的使用者名稱和密碼。',
      },
      noReplicationPrivilege: {
        title: '使用者無法進行復制',
        summary: '使用者已連線，但缺少實體備份所需的 REPLICATION 許可權。',
        managedNote:
          '在 AWS RDS 上執行 GRANT rds_replication TO {{username}}；在 Azure / GCP 上，請在服務商控制檯中為該角色開啟複製。',
      },
      noWalSwitchPrivilege: {
        title: '使用者無法強制切換 WAL 段',
        summary:
          '持續 WAL 串流傳輸只在來源資料庫關閉一個 WAL 段之後才上傳它。因此對於很少寫入的資料庫，最新的 WAL 會一直留在來源資料庫上，直到 16 MB 的段寫滿。定時強制切換段，才能讓兩次備份之間的還原點保持接近目前時刻，這需要 pg_switch_wal() 的 EXECUTE 許可權。完整和增量備份不需要它：它們還原到各自備份完成的時間點，不會重放歸檔的 WAL。',
        switchBackupTypeNote:
          '或者把備份類型切換為<bold>完整 + 增量</bold>，這種類型不會重放歸檔的 WAL。',
        managedNote:
          '在託管 PostgreSQL（RDS / Azure / GCP）上，該函式屬於平臺的超級使用者，無法向客戶角色授予它的 EXECUTE 許可權，因此無法使用持續 WAL 串流傳輸，請改用完整 + 增量備份。',
      },
      walLevelInvalid: {
        title: 'wal_level 設定過低',
        summary: '實體備份要求 wal_level 為“replica”或“logical”，目前設定低於此要求。',
        applyNote: '應用下面的修改，然後重啟 PostgreSQL（wal_level 只有在重啟後才會生效）。',
        managedNote: '在託管 PostgreSQL 上，請在服務商的參數組中設定 wal_level。',
      },
      noWalSenders: {
        title: '沒有可用的 WAL 傳送程序',
        summary: 'max_wal_senders 為 0，因此 PostgreSQL 無法為備份傳輸 WAL。',
        managedNote: '在託管 PostgreSQL 上，請在服務商的參數組中設定 max_wal_senders。',
      },
      noReplicationSlots: {
        title: '沒有可用的複製槽',
        summary: 'max_replication_slots 為 0，因此 PostgreSQL 無法為備份分配複製槽。',
        managedNote: '在託管 PostgreSQL 上，請在服務商的參數組中設定 max_replication_slots。',
      },
      walSummaryDisabled: {
        title: 'summarize_wal 已關閉',
        summary: '增量備份要求開啟 summarize_wal，但這臺伺服器目前沒有開啟。',
        applyNote: '應用下面的修改（無需重啟）。',
        thenNote: '然後',
        managedNote: '在託管 PostgreSQL 上，請在服務商的參數組中設定 summarize_wal。',
      },
      customTablespaces: {
        title: '不支援自訂表空間',
        summary:
          '該叢集有 pg_default / pg_global 之外的表空間，實體備份無法傳輸它們。請刪除自訂表空間，或把這個資料庫改為邏輯備份。',
      },
      systemIdentifierMismatch: {
        title: '這是另一個叢集',
        summary:
          '此位址上叢集的系統識別符號（system identifier）與最初設定備份時的叢集不同。請把 Databasus 指向原來的叢集，或新增一個資料庫條目。',
      },
      connectionFailed: {
        title: '無法連線',
        summary:
          'Databasus 無法建立複製連線。請檢查主機、連線埠、SSL 模式，並確認沒有防火牆阻止連線。',
      },
    },
  },
  restores: {
    list: {
      empty: '還沒有還原記錄',
    },
    targetDatabase: {
      select: '選擇要還原到的資料庫',
      description:
        '填寫要把備份還原到的資料庫資訊。<underline>還原前需要先建立一個空資料庫</underline>。還原過程中，其中現有的資料都會被清除。',
      notInUseHint:
        '請確保該資料庫目前沒有被使用（你多半不會想把資料還原到建立備份的同一個資料庫）',
      submit: '還原到此資料庫',
    },
    card: {
      startedAt: '開始時間',
      duration: '耗時',
      errorDetailsTooltip: '點選檢視錯誤詳情',
      cancelTooltip: '取消還原',
      expectedDurationHint: '還原耗時通常是備份的 3 到 5 倍（取決於資料類型，有時更快，有時更慢）',
      expectedDurationLimit: '因此預計最多需要 {{duration}}（通常會快得多）',
    },
    errorDetails: {
      title: '還原錯誤詳情',
      copied: '錯誤資訊已複製到剪貼簿',
      excludeExtensionsTip:
        '<bold>💡 提示：</bold>還原到託管 PostgreSQL 服務（如 Yandex Cloud、AWS RDS 等）時通常會出現這個錯誤。請在還原前到高階設定中開啟<bold>“排除擴充套件”</bold>再試。',
    },
    cancelConfirmation: {
      warning:
        '<bold>⚠️ 警告：</bold>取消還原很可能會讓資料庫處於損壞或不完整的狀態。再次嘗試還原前，需要重新增立該資料庫。',
      question: '確定要取消嗎？',
      confirm: '是，取消還原',
    },
    olderTargetWarning: {
      message:
        '警告：此備份由 PostgreSQL {{clientVersion}} 客戶端工具生成，但目標伺服器執行的是 PostgreSQL {{targetVersion}}。較新客戶端生成的傾印在較舊的伺服器上可能無法還原，例如因為舊伺服器不認識某些設定。',
      question: '仍要繼續還原嗎？',
      confirm: '仍然還原',
    },
  },
  healthcheck: {
    attempts: {
      title: '可用性檢查記錄',
      period: '時間範圍',
      periods: {
        today: '今天',
        sevenDays: '7 天',
        thirtyDays: '30 天',
        allTime: '全部',
      },
      empty: '暫無資料',
    },
    config: {
      enable: '啟用可用性檢查',
      isEnabled: '是否啟用可用性檢查',
      notifyWhenUnavailable: '不可用時通知',
      checkIntervalMinutes: '檢查間隔（分鐘）',
      attemptsBeforeDown: '判定不可用前的嘗試次數',
      storeAttemptsDays: '記錄保留時長（天）',
      tooltips: {
        enable: '為此資料庫啟用或禁用可用性檢查',
        notifyWhenUnavailable: '資料庫變為不可用時傳送通知',
        checkIntervalMinutes: '多久檢查一次資料庫的可用性（分鐘）',
        attemptsBeforeDown: '失敗多少次後將資料庫標記為不可用',
        storeAttemptsDays: '可用性檢查記錄保留多少天',
      },
    },
  },
  workspaces: {
    roles: {
      owner: '所有者',
      admin: '管理員',
      member: '成員',
      viewer: '檢視者',
    },
    fields: {
      name: '工作區名稱',
      namePlaceholder: '輸入工作區名稱',
    },
    create: {
      title: '建立工作區',
      submit: '建立工作區',
      defaultName: '我的工作區',
      description:
        '工作區用來集中管理：<lineBreak/>- 你的資料庫；<lineBreak/>- 儲存空間（本機磁碟、S3、Google Drive 等）；<lineBreak/>- 通知渠道（郵件、Slack、Telegram 等）；<lineBreak/>- 存取控制（如果你有團隊）。',
      nameRequired: '請輸入工作區名稱',
      created: '工作區已建立',
      permissionDenied: {
        title: '許可權不足',
        description: '你沒有建立工作區的許可權。請聯絡管理員為你建立工作區。',
      },
    },
    settings: {
      title: '工作區設定',
      readOnly: '你沒有修改這些設定的許可權',
      nameRequired: '工作區名稱不能為空',
      saving: '正在儲存…',
      saveChanges: '儲存變更',
      reset: '重置',
      updated: '基本資訊已更新',
    },
    delete: {
      dangerZone: '危險區域',
      title: '刪除此工作區',
      description: '工作區刪除後無法復原。與該工作區關聯的所有資料和資源都將被永久刪除。',
      submit: '刪除工作區',
      deleting: '正在刪除…',
      notFound: '未找到工作區',
      noPermission: '你沒有刪除此工作區的許可權',
      hasDatabases: '無法刪除工作區，請先刪除其中所有資料庫。目前資料庫數量：{{count}}。',
      deleted: '工作區已刪除',
      confirmation: {
        title: '刪除工作區',
        question: '確定要刪除工作區 <workspaceName/> 嗎？',
        warning: '<bold>此操作無法復原。</bold>所有資料和關聯資源都將被永久刪除。',
        submit: '刪除工作區',
      },
    },
    members: {
      title: '使用者',
      count: '成員：{{count}}',
      empty: '沒有成員',
      emptyHint: '點選“新增成員”開始',
      fields: {
        member: '成員',
        role: '角色',
        joined: '加入時間',
        actions: '操作',
        email: '電子郵件地址',
        emailPlaceholder: '輸入電子郵件地址',
      },
      add: {
        button: '新增成員',
        title: '新增成員',
        submit: '新增成員',
        emailRequired: '請輸入電子郵件',
        hint: '如果該使用者已存在，會被直接新增；否則會向其傳送邀請。',
        added: '成員已新增',
      },
      invited: {
        title: '已邀請使用者',
        sentTo: '邀請已傳送至 {{email}}',
        description:
          '該使用者還沒有註冊，但已受邀加入工作區。用這個電子郵件註冊後，將自動成為工作區成員。',
      },
      roleChanged: '成員角色已更新',
      remove: {
        tooltip: '移除成員',
        title: '移除成員',
        question: '確定要將“{{email}}”從此工作區移除嗎？',
        removed: '已移除成員“{{email}}”',
      },
    },
    transferOwnership: {
      button: '轉讓所有權',
      title: '轉讓工作區所有權',
      submit: '轉讓所有權',
      warning:
        '<bold>警告：</bold>此操作無法復原。你將失去此工作區的所有權，新的所有者將獲得完全控制權。',
      noEligibleMembers:
        '沒有可以轉讓所有權的成員。工作區中至少還要有一位其他成員，才能轉讓所有權。',
      newOwner: '選擇新所有者',
      newOwnerPlaceholder: '選擇要轉讓所有權的成員',
      newOwnerHint: '所選成員將成為工作區所有者',
      memberRequired: '請選擇要轉讓所有權的成員',
      memberNotFound: '未找到所選成員',
      transferred: '所有權已轉讓',
    },
    auditLogs: {
      empty: '此工作區沒有稽核記錄。',
    },
  },
  users: {
    roles: {
      admin: '管理員',
      member: '成員',
    },
    fields: {
      name: '姓名',
      email: '電子郵件',
      yourName: '你的姓名',
      yourEmail: '你的電子郵件',
      confirmPassword: '確認密碼',
      newPassword: '新密碼',
    },
    validation: {
      nameRequired: '請輸入姓名',
      emailRequired: '請輸入電子郵件',
      passwordTooShort: '密碼長度至少為 {{minLength}} 個字元',
      passwordsDoNotMatch: '兩次輸入的密碼不一致',
    },
    oauth: {
      divider: '或使用電子郵件繼續',
      continueWithGithub: '使用 GitHub 繼續',
      continueWithGoogle: '使用 Google 繼續',
      invalidConfiguration: 'OAuth 設定無效',
    },
    signIn: {
      title: '登入',
      submit: '登入',
      noAccount: '還沒有帳號？<signUpLink>註冊</signUpLink>',
      forgotPassword: '忘記密碼？',
    },
    signInCode: {
      title: '輸入驗證碼',
      description: '我們已將六位驗證碼傳送到 {{email}}，它在傳送 10 分鐘後失效。',
      code: '登入驗證碼',
      submit: '登入',
      resend: '重新傳送驗證碼',
      backToSignIn: '返回登入',
      invalidCode: '驗證碼為 6 位數字',
      codeResent: '新的驗證碼已傳送，之前的驗證碼已失效。',
    },
    signUp: {
      title: '註冊',
      namePlaceholder: '張三',
      submit: '註冊',
      administersInstance: '這是該執行個體上的第一個帳戶，它將成為執行個體的管理員。',
      hasAccount: '已有帳號？<signInLink>登入</signInLink>',
    },
    requestPasswordReset: {
      title: '重置密碼',
      description: '輸入你的電子郵件地址，我們會給你傳送重置碼。',
      submit: '傳送重置碼',
      codeSent: '如果該電子郵件已註冊，重置碼已傳送',
      rememberPassword: '想起密碼了？<signInLink>登入</signInLink>',
    },
    resetPassword: {
      title: '重置密碼',
      description: '輸入郵件中的重置碼和新密碼。',
      code: '重置碼',
      confirmPassword: '確認密碼',
      submit: '重置密碼',
      invalidCode: '重置碼必須是 6 位數字',
      succeeded: '密碼已重置，正在跳轉到登入頁…',
      noCode: '沒有收到重置碼？<requestCodeLink>重新取得</requestCodeLink>',
      backToSignIn: '返回登入',
    },
    profile: {
      title: '個人資料',
      information: '個人資訊',
      userId: '使用者 ID',
      namePlaceholder: '輸入你的姓名',
      emailPlaceholder: '輸入你的電子郵件',
      role: '角色',
      saveChanges: '儲存變更',
      noChanges: '沒有需要儲存的變更',
      updated: '個人資料已更新',
      logout: '退出登入',
      changePassword: {
        title: '修改密碼',
        newPasswordPlaceholder: '輸入新密碼',
        confirmPassword: '確認新密碼',
        confirmPasswordPlaceholder: '再次輸入新密碼',
        passwordsDoNotMatch: '兩次輸入的新密碼不一致',
        changing: '正在修改密碼…',
        submit: '修改密碼',
        signedInWithNewPassword: '已使用新密碼登入',
      },
    },
    list: {
      title: 'Databasus 使用者',
      loadedOfTotal: '使用者：{{loaded}} / {{total}}',
      searchPlaceholder: '按電子郵件或姓名搜尋',
      empty: '沒有找到使用者。',
      allLoaded: '已載入全部使用者（共 {{count}} 位）',
      columns: {
        user: '使用者',
        systemRole: '系統角色',
        isActive: '啟用狀態',
        created: '建立時間',
      },
      roleLabel: '角色：',
      activeLabel: '啟用：',
      viewAuditLogs: '檢視稽核記錄',
      activated: '使用者已啟用',
      deactivated: '使用者已停用',
      roleChanged: '使用者角色已變更',
    },
    auditLogs: {
      title: '使用者稽核記錄',
      empty: '該使用者沒有稽核記錄。',
    },
  },
  verification: {
    triggers: {
      manual: '手動',
      scheduled: '定時',
    },
    notificationTypes: {
      verificationSuccess: '驗證成功',
      verificationFailed: '驗證失敗',
    },
    config: {
      scheduledVerification: '定時驗證',
      scheduledVerificationTooltip:
        '定期把最新的備份還原到一次性 Postgres 容器中，驗證它能否還原。<docsLink>了解工作原理</docsLink>',
      interval: '驗證週期',
      afterBackup: '備份後',
      afterBackupHint:
        '每次備份成功後自動執行。如果佇列中還有待處理的驗證，會將其取消，以免備份快於驗證時佇列無限增長。',
      weekday: '每週驗證日',
      dayOfMonth: '每月驗證日',
      timeOfDay: '驗證時間',
      cron: {
        expression: 'Cron 表示式（UTC）',
        format: 'Cron 格式：分鐘 小時 日 月 星期（UTC）',
        examples: '示例：',
        exampleEverySunday: '- {{expression}}：每週日 4:00（UTC）',
        exampleEverySixHours: '- {{expression}}：每 6 小時一次',
        nextRun: '下次執行：{{dateTime}}<lineBreak/>（{{relativeTime}}）',
        invalid: '無效的 Cron 表示式',
      },
      notifications: '通知',
      noNotifications: '無',
    },
    runs: {
      title: '還原驗證',
      description: '每一行是一次嘗試：把此資料庫的一個備份還原到臨時副本中',
      noAgentsWarning:
        '還沒有註冊驗證代理。請在 Databasus 設定分頁的“驗證代理”部分新增。<docsLink>了解更多</docsLink>',
      empty: '還沒有還原驗證記錄',
      allLoaded: '已載入全部還原驗證（共 {{total}} 次）',
      attempt: '第 {{attempt}} 次嘗試',
      restoreDuration: '還原耗時 {{duration}}',
      durationTooltip: '代理報告的還原與驗證總時長。',
      fields: {
        createdAt: '建立時間',
        trigger: '觸發方式',
        duration: '耗時',
        actions: '操作',
      },
      actions: {
        cancel: '取消驗證',
        viewDetails: '檢視詳情',
      },
      details: {
        title: '還原驗證詳情',
        loadFailed: '無法載入還原驗證詳情。',
        cancellationReason: '取消原因',
        failure: '失敗原因',
        sections: {
          status: '狀態',
          timeline: '時間線',
          results: '結果與診斷',
        },
        fields: {
          attempt: '嘗試次數',
          startedAt: '開始時間',
          finishedAt: '完成時間',
          restoreDuration: '還原耗時',
          verifyDuration: '驗證耗時',
          restoredDbSize: '還原後的資料庫大小',
          schemas: '模式',
          tables: '表',
          pgRestoreExitCode: 'pg_restore 退出碼',
        },
        tableStats: {
          title: '各錶行數',
          empty: '沒有上報各表的統計資料。',
          schema: '模式',
          table: '表',
          rows: '行數',
        },
      },
    },
    agents: {
      title: '驗證代理',
      description: '驗證代理負責執行還原驗證，確認備份可以還原（<docsLink>了解更多</docsLink>）',
      empty: '還沒有註冊代理。',
      columns: {
        capacity: '資源',
        created: '建立時間',
      },
      capacity: {
        cpu: '{{count}} 個 CPU',
        ram: '{{count}} GB 記憶體',
        disk: '{{count}} GB 磁碟',
        jobs: '任務數：{{count}}',
        notReported: '尚未上報',
      },
      actions: {
        viewInstallCommands: '檢視安裝命令',
        rotateToken: '輪換令牌',
      },
      deleteConfirmation: '確定要刪除這個代理嗎？',
      deleted: '代理“{{name}}”已刪除',
      copyFailed: '複製失敗',
      agentId: '代理 ID：',
      create: {
        title: '建立驗證代理',
        tokenNotice: '下一個介面會顯示生成的令牌，且只顯示這一次。',
        namePlaceholder: '代理名稱',
        nameRequired: '請輸入名稱',
      },
      rotate: {
        title: '輪換令牌',
        description:
          '輪換 <agentName/> 的令牌會立即讓現有令牌失效。仍在使用舊令牌的代理會在下一次心跳時被拒絕。',
      },
      token: {
        title: '代理令牌',
        tokenFor: '<agentName/> 的令牌：',
        shownOnce: '令牌只顯示一次。請妥善儲存，之後無法再次檢視。',
        confirmSaved: '我已儲存令牌',
      },
      install: {
        title: '安裝命令',
        description:
          '<agentName/> 的安裝命令。請把 <tokenPlaceholder/> 替換為你建立該代理或上次輪換令牌時儲存的令牌。',
        architecture: '架構',
        stepInstall: '第 1 步：安裝',
        stepLaunch: '第 2 步：啟動',
        capacityDefaults:
          '下面的資源值只是初始預設值，請根據執行代理的機器調整 <code>--max-cpu</code>、<code>--max-ram-mb</code>、<code>--max-disk-gb</code> 和 <code>--max-concurrent-jobs</code>。',
        afterInstallation: '安裝之後',
        runsInBackground: '執行 <code>start</code> 後，代理在後臺執行',
        checkStatus: '檢視狀態：<code>./verification-agent status</code>',
        viewLogs: '檢視記錄：工作目錄下的 <code>databasus-verification.log</code>',
        stop: '停止代理：<code>./verification-agent stop</code>',
      },
    },
  },
  notifiers: {
    advancedSettings: '高階設定',
    types: {
      email: '郵件',
      telegram: 'Telegram',
      webhook: 'Webhook',
      slack: 'Slack',
      discord: 'Discord',
      teams: 'Teams',
      mattermost: 'Mattermost',
    },
    notificationTypes: {
      all: '全部',
      backupSuccess: '備份成功',
      backupFailed: '備份失敗',
      healthcheckSuccess: '可用性檢查透過',
      healthcheckFailed: '可用性檢查失敗',
      verificationSuccess: '還原驗證成功',
      verificationFailed: '還原驗證失敗',
    },
    mattermostDeliveryModes: {
      webhook: '傳入 Webhook',
      bot: '機器人帳號',
    },
    list: {
      addNotifier: '新增通知渠道',
      searchPlaceholder: '搜尋通知渠道',
      noSearchResults: '沒有找到與“{{query}}”匹配的通知渠道',
      description: '通知渠道是接收通知的地方（郵件、Slack、Telegram 等）',
      backToList: '← 返回通知渠道列表',
    },
    card: {
      notifyTo: {
        email: '透過郵件通知',
        telegram: '透過 Telegram 通知',
        webhook: '透過 Webhook 通知',
        slack: '透過 Slack 通知',
        discord: '透過 Discord 通知',
        teams: '透過 Teams 通知',
        mattermost: '透過 Mattermost 通知',
      },
      hasSendError: '傳送出錯',
    },
    create: {
      title: '新增通知渠道',
      namePlaceholder: '我的聊天',
    },
    config: {
      namePlaceholder: '輸入名稱',
      settingsTitle: '通知渠道設定',
      sendError: {
        title: '傳送錯誤',
        errorLabel: '錯誤：',
        clearHint: '清除此錯誤（任選其一）：',
        clearBySendingTest: '- 用下方按鈕傳送測試通知（即使你已經修改了設定）；',
        clearByNextNotification: '- 等待下一條通知成功傳送；',
      },
      removeConfirmation: '確定要刪除這個通知渠道嗎？此操作無法復原。',
      removeBlockedByDatabases: '有資料庫正在使用這個通知渠道。請先在這些資料庫中移除它。',
    },
    testNotification: {
      send: '傳送測試通知',
      sent: {
        title: '測試通知已傳送',
        description: '測試通知傳送成功',
      },
    },
    transfer: {
      title: '將通知渠道轉移到其他工作區',
      blockedByDatabases: '有資料庫正在使用這個通知渠道。請先轉移或刪除相關資料庫。',
      description: '選擇要將此通知渠道轉移到哪個工作區。',
      targetWorkspace: '目標工作區',
      selectWorkspace: '選擇工作區',
      submit: '轉移',
    },
    fields: {
      botToken: '機器人令牌',
      targetChatId: '目標聊天 ID',
      webhookUrl: 'Webhook URL',
      channelId: '頻道 ID',
      skipTlsVerify: '跳過 TLS 驗證',
      skipTls: '跳過 TLS 驗證',
    },
    email: {
      targetEmail: '收件電子郵件',
      targetEmailTooltip: '接收通知的電子郵件地址',
      smtpHost: 'SMTP 主機',
      smtpPort: 'SMTP 連線埠',
      smtpUser: 'SMTP 使用者',
      smtpPassword: 'SMTP 密碼',
      smtpPasswordPlaceholder: '密碼',
      from: '發件人',
      fromTooltip:
        '可選。發件人位址，可以帶名稱：Acme Backups <noreply@example.com>。留空時，如果 SMTP 使用者是電子郵件地址就使用它，否則使用 noreply@ 加 SMTP 主機',
      fromAuto: '（自動）',
      skipTlsTooltip:
        '跳過 TLS 憑證驗證。如果 SMTP 伺服器使用自我簽署憑證，請開啟此項。警告：這會降低安全性。',
      security: '連線安全',
      securityTooltip:
        'TLS 從一開始就加密連線，通常用於 465 連線埠。STARTTLS 在連線建立後升級為加密連線，伺服器不支援時傳送失敗。“不加密”會以明文傳送密碼和郵件，只應在不支援 STARTTLS 的可信中繼上使用。',
      securityModes: {
        tls: 'TLS',
        starttls: 'STARTTLS',
        none: '不加密',
      },
      heloName: '問候名稱',
      heloNamePlaceholder: '自動',
      heloNameTooltip:
        'Databasus 向郵件伺服器介紹自己時使用的名稱（EHLO）。留空時使用 DATABASUS_URL 中的主機，或本機的主機名稱。',
    },
    telegram: {
      botTokenHelpLink: '如何取得 Telegram 機器人 API 令牌？',
      targetChatIdTooltip: '接收訊息的聊天（可以是你的私人聊天，也可以是群組）',
      chatIdHelp: {
        toggle: '如何取得 Telegram 聊天 ID？',
        privateChat:
          '要取得你的聊天 ID，請在 Telegram 中給 <botLink>@getmyid_bot</botLink> 發訊息。<underline>請確認你已經和該機器人開始聊天</underline>',
        group:
          '要取得群組的聊天 ID，請把你的機器人和 <botLink>@getmyid_bot</botLink> 一起加入群組，然後傳送 /start（你會看到聊天 ID）',
      },
      botApiUrl: 'Bot API URL',
      botApiUrlTooltip:
        '可選。自建的 Telegram Bot API 伺服器，或 Telegram API 前面的反向代理。留空時使用官方 Telegram API。該伺服器會收到機器人令牌。切換到自建伺服器之前，需要先用 logOut 方法讓機器人從官方 API 登出',
      useProxy: '使用代理',
      useProxyTooltip: '透過代理傳送 Telegram API 請求',
      proxy: '代理',
      proxyUrl: '代理 URL',
      proxyUrlTooltip: '支援 http、https、socks5、socks5h。可以包含使用者名稱和密碼',
      sendToGroupTopic: '傳送到群組話題',
      sendToGroupTopicTooltip: '開啟後，訊息會傳送到群組聊天中的指定話題',
      threadId: '話題 ID',
      threadIdTooltip: '接收訊息的話題 ID',
      topicId: '話題 ID',
      threadIdHelp: {
        steps:
          '要取得話題 ID，請在 Telegram 群組中進入該話題，點選頂部的話題名稱，開啟話題資訊。複製話題連結，取 URL 中的最後一個數字。',
        example:
          '<bold>示例：</bold>如果話題連結是 <code>https://t.me/c/2831948048/3</code>，話題 ID 就是 <code>3</code>',
        note: '<bold>注意：</bold>話題功能只能在群組聊天中使用，私人聊天中不可用。',
      },
    },
    webhook: {
      method: '請求方法',
      sendOn: '傳送時機',
      sendOnPlaceholder: '選擇通知類型',
      customHeaders: '自訂請求標頭',
      customHeadersTooltip:
        '為 Webhook 請求新增自訂 HTTP 請求標頭（例如 Authorization、X-API-Key）',
      savedHeadersHidden: '*出於安全考慮，已儲存的請求標頭不予顯示',
      headers: '請求標頭',
      headerNamePlaceholder: '請求標頭名稱',
      headerValuePlaceholder: '請求標頭值',
      addHeader: '新增請求標頭',
      hiddenHeaderValue: '（已隱藏）',
      bodyTemplate: {
        title: '請求本文範本',
        headingVariable: '<variable/>：通知標題',
        messageVariable: '<variable/>：通知內容',
        invalidJsonFormat: 'JSON 格式無效',
        invalidJson: '無效的 JSON',
      },
      exampleRequest: {
        title: '請求示例',
        headers: '請求標頭：',
      },
    },
    slack: {
      howToConnect: '如何連線 Slack（如何取得機器人令牌和聊天 ID）？',
      testFailedHint:
        '請確認頻道是公開的，或者機器人已被加入私有頻道（透過 @invite）或群組。私人訊息請使用 Slack 個人資料中的使用者 ID。',
    },
    discord: {
      channelWebhookUrl: '頻道 Webhook URL',
      howTo: {
        title: '如何取得 Discord Webhook URL：',
        createChannel: '1. 建立或選擇一個 Discord 頻道',
        openChannelSettings: '2. 開啟頻道設定（齒輪圖示）',
        openIntegrations: '3. 進入整合設定',
        createWebhook: '4. 新增一個 Webhook',
        copyWebhookUrl: '5. 複製 Webhook URL',
        privacyNote: '注意：如有需要，請把頻道設為私密',
      },
    },
    teams: {
      howToConnect: '如何連線 Microsoft Teams？',
      powerAutomateUrl: 'Power Automate URL',
      powerAutomateUrlLabel: 'Power Automate URL：',
      powerAutomateUrlTooltip:
        '你的 Power Automate 流中的 HTTP 端點（When an HTTP request is received 觸發器）',
      showFullUrl: '顯示',
      hideFullUrl: '隱藏',
    },
    mattermost: {
      howToConnect: '如何連線 Mattermost？',
      connectVia: '連線方式',
      incomingWebhookUrl: '傳入 Webhook URL',
      serverUrl: '伺服器 URL',
      channelIdTooltip:
        '26 個字元的頻道 ID，不是頻道名稱。開啟頻道，點選頻道名稱，然後選擇 View Info。',
      channel: '頻道',
      showOptionalSettings: '顯示可選設定',
      hideOptionalSettings: '隱藏可選設定',
      channelOverride: '改用其他頻道',
      channelOverrideTooltip:
        '釋出到這個頻道，而不是建立 Webhook 時指定的頻道。如果 Webhook 已鎖定到某個頻道，Mattermost 會忽略此設定。',
      postAsUsername: '釋出者使用者名稱',
      postAsUsernameTooltip:
        '需要在 Mattermost 系統控制檯中開啟 Enable integrations to override usernames，否則此設定會被忽略。',
      postAsIconUrl: '釋出者頭像 URL',
      postAsIconUrlTooltip:
        '需要在 Mattermost 系統控制檯中開啟 Enable integrations to override profile picture icons，否則此設定會被忽略。',
      skipTlsTooltip:
        '跳過 TLS 憑證驗證。如果 Mattermost 伺服器使用自我簽署憑證，請開啟此項。警告：這會降低安全性。',
      webhookHowTo: {
        title: '如何取得傳入 Webhook URL：',
        openIncomingWebhooks: '1. 主選單 → Integrations → Incoming Webhooks',
        addWebhook: '2. 點選 Add Incoming Webhook，選擇頻道',
        copyUrl: '3. 複製生成的 URL',
      },
      botHowTo: {
        title: '如何取得機器人令牌：',
        addBotAccount: '1. Integrations → Bot Accounts → Add Bot Account',
        copyToken: '2. 複製建立後顯示的令牌（只顯示一次）',
        addBotToChannel: '3. 把機器人加入團隊和頻道',
      },
    },
  },
  backups: {
    encryption: {
      none: '無',
      encrypted: '已啟用',
    },
    encryptionOptions: {
      none: '不加密',
      encrypted: '加密備份檔案',
    },
    list: {
      title: '備份',
      scheduledBackupsDisabled: '定時備份已禁用（可以在備份設定中重新啟用）',
      empty: '還沒有備份',
      allLoaded: '已載入全部備份（共 {{count}} 個）',
      clickToSeeErrorDetails: '點選檢視錯誤詳情',
      errorDetailsTitle: '備份錯誤詳情',
      restoreTitle: '從備份還原',
      columns: {
        createdAt: '建立時間',
        created: '建立時間',
        size: '大小',
        duration: '耗時',
        actions: '操作',
      },
      actions: {
        cancelBackup: '取消備份',
        deleteBackup: '刪除備份',
      },
    },
    filters: {
      allTypes: '全部類型',
      allStatuses: '全部狀態',
      before: '早於',
    },
    config: {
      backupsEnabled: '啟用備份',
      storage: '儲存空間',
      selectStorage: '選擇儲存空間',
      createNewStorage: '新增儲存空間',
      addStorageTitle: '新增儲存空間',
      storageDescription: '儲存空間是存放備份的位置（本機磁碟、S3、Google Drive 等）',
      storageChangeWarning: '如果更換儲存空間，該儲存空間中的所有備份都會被刪除。',
      storageChangeAcknowledge: '我已了解',
      encryption: '加密',
      notifications: '通知',
      noNotifications: '無',
      schedule: {
        weekday: '星期',
        dayOfMonth: '每月日期',
        cronExpression: 'Cron 表示式（UTC）',
        timeOfDay: '時間',
        nextRun: '下次執行：{{time}}',
        invalidCron: '無效的 Cron 表示式',
        cronHelp: {
          format: 'Cron 格式：分鐘 小時 日 月 星期（UTC）',
          examples: '示例：',
          dailyAt2am: '每天 2:00（UTC）',
          every6Hours: '每 6 小時一次',
          mondaysAt3am: '每週一 3:00（UTC）',
          twiceAMonth: '每月 1 日和 15 日 4:30（UTC）',
        },
      },
      gfs: {
        hourly: '每小時：{{count}}',
        daily: '每天：{{count}}',
        weekly: '每週：{{count}}',
        monthly: '每月：{{count}}',
        yearly: '每年：{{count}}',
        notConfigured: '未設定',
      },
    },
    logical: {
      retentionPolicyTypes: {
        gfs: 'GFS（保留最近 N 個每小時、每天、每週、每月和每年的備份）',
        timePeriod: '按時間範圍（最近 N 天）',
        count: '按數量（最近 N 個備份）',
      },
      notificationTypes: {
        backupSuccess: '備份成功',
        backupFailed: '備份失敗',
      },
      list: {
        backupNow: '立即備份',
        makeBackupNow: '立即建立備份',
        encrypted: '已加密',
        sizeTooltip:
          '上方：實際存放在儲存空間（本機、S3、Google Drive 等）中的壓縮後備份檔案大小。下方（灰色）：備份時資料庫未壓縮的原始大小。備份通常能壓縮到原來的五分之一左右。',
        databaseSize: '{{size}}（資料庫大小）',
        deleteConfirmation: '確定要刪除這個備份嗎？',
        actions: {
          restore: '從備份還原',
          verifyRestore: '驗證還原：安排代理把這個備份還原到臨時資料庫，並報告行數',
          download: {
            postgresql: '下載備份檔案。可以用 pg_restore 手動還原（custom 格式）',
            mysql: '下載備份檔案。可以用 mysql 客戶端手動還原（SQL 傾印）',
            mariadb: '下載備份檔案。可以用 mariadb 客戶端手動還原（SQL 傾印）',
            mongodb: '下載備份檔案。可以用 mongorestore 手動還原（歸檔格式）',
            other: '下載備份檔案',
          },
        },
        verifyRestore: {
          title: '驗證還原？',
          description:
            '代理會把這個備份還原到臨時資料庫並報告行數。視備份大小，這可能需要一些時間。<docsLink>了解工作原理</docsLink>',
          queue: '加入佇列',
          queued: '還原驗證已加入佇列',
        },
      },
      config: {
        interval: '備份週期',
        weekday: '每週備份日',
        dayOfMonth: '每月備份日',
        timeOfDay: '備份時間',
        encryptionTooltip:
          '備份加密後，儲存空間（S3、本機等）中的備份檔案無法直接使用。你可以透過 Databasus 還原備份，或用“下載”按鈕下載解密後的檔案。',
        retentionPolicy: '保留原則',
        retention: {
          timePeriodTooltip: '備份保留多長時間。超過這個期限的備份會被自動刪除。',
          count: '最近的備份',
          countTooltip: '只保留指定數量的最近備份，更早的備份會被自動刪除。',
          countValue: '保留最近 {{count}} 個備份',
          showGfsHint: '什麼是 GFS（祖父-父-子）？',
          hideGfsHint: '隱藏',
          gfsHint:
            'GFS（祖父-父-子）輪換：保留最近 N 個每小時、每天、每週、每月和每年的備份。這樣可以在合理的儲存空間空間內長期保留備份。',
          gfsTooltip: 'GFS（祖父-父-子）輪換：保留最近 N 個每小時、每天、每週、每月和每年的備份。',
          gfsFields: {
            hourly: '每小時備份',
            daily: '每天備份',
            weekly: '每週備份',
            monthly: '每月備份',
            yearly: '每年備份',
          },
        },
        advancedSettings: '高階設定',
        retryIfFailed: '備份失敗時重試',
        retryIfFailedShort: '失敗時重試',
        retryIfFailedTooltip:
          '自動重試失敗的備份。網路故障、儲存空間問題或資料庫暫時不可用都可能導致備份失敗。',
        maxFailedTries: '最大重試次數',
        maxFailedTriesTooltip: '失敗備份的最大重試次數。所有重試都失敗後，你會收到通知。',
      },
    },
    physical: {
      types: {
        full: '完整',
        incremental: '增量',
        wal: 'WAL',
      },
      retentions: {
        chains: '備份鏈',
        fullBackups: '完整備份',
        chainsAndFullBackups: '備份鏈和完整備份',
      },
      fullBackupsPolicies: {
        lastN: '數量',
        gfs: 'GFS',
      },
      notificationTypes: {
        backupSuccess: '備份成功',
        backupFailed: '備份失敗',
        chainBroken: '備份鏈中斷',
        walGap: 'WAL 缺口',
      },
      list: {
        totalUsage: '總占用：{{size}}',
        backUpNow: '立即備份',
        fullBackup: '完整備份',
        incrementalBackup: '增量備份',
        restore: '還原',
        restoreFromThisBackup: '從此備份還原',
        pointInTimeRestoreTitle: '時間點還原',
        walRange: {
          from: '開始',
          to: '結束',
          segments: '段數',
          gap: '缺口',
          gapBeforeTooltip: '此範圍與上一個範圍之間缺少 WAL 段，時間點還原無法跨越該缺口。',
        },
        deleteFullConfirmation:
          '刪除這個完整備份會移除它的整條備份鏈（所有依賴它的增量備份）。此操作無法復原。確定繼續嗎？',
        deleteIncrementalConfirmation:
          '刪除這個增量備份會移除所有依賴它的後續增量備份。此操作無法復原。確定繼續嗎？',
      },
      config: {
        fullBackupCadence: '完整備份頻率',
        incrementalBackupCadence: '增量備份頻率',
        incrementalMoreFrequent: '增量備份的頻率必須高於完整備份。',
        encryptionTooltip:
          '備份加密後，儲存空間中的檔案無法直接使用。你可以透過 Databasus 還原，或下載解密後的檔案。',
        retention: '保留',
        fullBackupsOnlyRetention: '這個資料庫只做完整備份，因此保留規則作用於完整備份。',
        retentionOptions: {
          chains: '保留最近 N 條備份鏈',
          chainsAndFullBackups: '保留最近 N 條備份鏈和完整備份',
        },
        retentionHelp: {
          chain:
            '一條備份鏈由一個完整備份和依賴它的增量備份（以及串流傳輸的 WAL）組成。還原時需要完整備份和鏈中的每一個增量備份。',
          chains:
            '只保留最近 N 條備份鏈。一條鏈被淘汰時，它的完整備份和所有增量備份都會被刪除，只能在保留下來的鏈範圍內還原。',
          chainsAndFullBackups:
            '保留最近 N 條備份鏈，用於還原到近期的時間點。此外，還會按下面的原則（GFS 或最近 N 個）保留更早的獨立完整備份：它們的增量備份刪除後，這些完整備份仍可作為長期還原點。',
        },
        chainsCount: '備份鏈數',
        chainsCountTooltip: '要保留的最近備份鏈數量。一條鏈由一個完整備份及其增量備份組成。',
        fullBackups: '完整備份',
        fullBackupsCount: '最近的完整備份',
        gfsFields: {
          hourly: '每小時',
          daily: '每天',
          weekly: '每週',
          monthly: '每月',
          yearly: '每年',
        },
        chainsKept: '保留的備份鏈',
        fullBackupsKept: '保留的完整備份',
        fullBackupsCountValue: '最近 {{count}} 個',
        gfsTooltip:
          'GFS（祖父-父-子）輪換：保留最近 N 個每小時、每天、每週、每月和每年的完整備份。',
      },
      restore: {
        errorHints: {
          downloadInProgress: '這個資料庫已有一個還原下載正在進行。請等待它完成後重試。',
          targetTimeUnreachable:
            '無法還原到所選的目標時間：WAL 存在缺口，或者時間超出了可用範圍。請選擇其他時間，或還原到最新的可用時間點。',
        },
        backupIntro:
          '這個備份的還原命令。完整備份只需要它自己的檔案；增量備份還需要它的完整備份和之前的所有增量備份。',
        pointInTimeIntro: '時間點還原。選擇目標時間，或留空以還原到最新的可用時間點。',
        targetTime: '目標時間',
        latestAvailable: '最新可用時間點',
        preparing: '正在準備還原命令…',
        linkExpiry:
          '下載連結只能使用一次，15 分鐘後失效。如果命令因 401 失敗，重新開啟此對話方塊即可生成新連結。',
        copied: '已複製',
        methods: {
          script: '使用指令碼',
          manual: '手動',
        },
        environments: {
          host: '主機上的 PostgreSQL',
        },
        fields: {
          restoreDirectory: '還原目錄',
          pgBinPath: 'PostgreSQL bin 路徑',
          pgBinPathPlaceholder: '{{path}}（可選）',
          pgImage: 'PostgreSQL 映象',
        },
        scriptTitles: {
          host: '在還原目標主機上執行',
          docker: '在可以使用 Docker 的機器上執行',
        },
        manualSteps: {
          download: '下載備份包',
          extract: '解壓',
          verify: '校驗傳輸完整性',
          reconstruct: '重建資料目錄',
          checkConfig: '檢查叢集的設定檔案',
          wireUpRecovery: '解壓 WAL 並設定還原',
          recoveryParameters: '如果還原因參數設定而停止',
        },
        beforeYouRun: {
          title: '執行前',
          host: {
            clientTools:
              '安裝 PostgreSQL {{version}} 客戶端工具。如果它們不在 PATH 中，請在上方設定 bin 路徑。',
            zstd: '解壓備份需要 zstd。',
            zstdWithWal: '解壓備份及其 WAL 需要 zstd。',
            emptyDirectory: '還原目錄必須為空，且不能是正在使用的叢集目錄。',
          },
          docker: {
            image: '使用 postgres:{{version}} 映象，主版本必須與來源資料庫一致。',
            hostTools:
              '主機上需要 zstd、tar 和 curl：下載和所有解壓（基礎備份）都在主機上進行，容器中只執行 pg_combinebackup，因此 postgres 映象不需要額外工具。',
            hostToolsWithWal:
              '主機上需要 zstd、tar 和 curl：下載和所有解壓（基礎備份和 WAL）都在主機上進行，容器中只執行 pg_combinebackup，因此 postgres 映象不需要額外工具。',
          },
        },
        afterItFinishes: {
          title: '完成之後',
          host: {
            clusterLocation: '叢集位於 <dataDir/>。',
            changeOwner: '以 root 身份變更所有者：<command/>。',
            start: '以 postgres 身份啟動：<command/>，或者把 data_directory 指向它。',
            walReplay: 'PostgreSQL 會重放 WAL 並提升為主要伺服器，請在記錄中檢視還原是否完成。',
          },
          docker: {
            clusterLocation: '主機上的 <dataDir/> 就是還原出的叢集。',
            volumeMount:
              'PostgreSQL {{version}} 把資料目錄放在卷 <volumeDir/> 下的 <dataDir/>。請把輸出目錄掛載為該卷的根目錄，這正是 docker-compose 中 <code>./pgdata:/var/lib/postgresql</code> 所期望的：',
            dataDirectoryMount: '把叢集繫結掛載到映象的資料目錄 <dataDir/>：',
            ownership: '所有者必須與映象內 postgres 使用者的 uid 一致（官方映象中為 999）。',
          },
          missingConfigFiles:
            '基礎備份只複製資料目錄。如果源 PostgreSQL 把設定放在資料目錄之外（Debian/Ubuntu 的佈局，<configDir/>），那麼 <code>postgresql.conf</code>、<code>pg_hba.conf</code> 和 <code>pg_ident.conf</code> 不在備份中，伺服器將無法啟動。在源伺服器上用 <code>psql -Atc "SHOW config_file"</code> 查出它們的實際路徑，把三個檔案都複製過來，然後註釋掉 <code>data_directory</code>、<code>hba_file</code>、<code>ident_file</code>、<code>external_pid_file</code> 以及 <code>ssl</code> / <code>ssl_*</code> 這些行，建立 <code>conf.d</code>；如果在容器中執行，還要設定 <code>listen_addresses = \'*\'</code>。如果這些檔案都在，卻仍然出現同樣的錯誤，那麼出錯的是掛載的目錄，而不是設定。',
        },
      },
    },
  },
  storages: {
    types: {
      local: '本機儲存空間',
      s3: 'S3',
      googleDrive: 'Google Drive',
      nas: 'NAS',
      azureBlob: 'Azure Blob Storage',
      ftp: 'FTP',
      sftp: 'SFTP',
      rclone: 'Rclone',
    },
    s3StorageClasses: {
      default: '預設（Standard）',
      standard: 'Standard',
      standardIa: 'Standard - Infrequent Access',
      onezoneIa: 'One Zone - Infrequent Access',
      intelligentTiering: 'Intelligent Tiering',
      reducedRedundancy: 'Reduced Redundancy',
      glacierIr: 'Glacier Instant Retrieval',
    },
    azureBlobAuthMethods: {
      accountKey: '帳戶金鑰',
      connectionString: '連線字串',
    },
    sftpAuthMethods: {
      password: '密碼',
      privateKey: '私鑰',
    },
    list: {
      addStorage: '新增儲存空間',
      description: '儲存空間是存放備份的位置（本機磁碟、S3 等）',
      backToList: '← 返回儲存空間列表',
    },
    card: {
      type: {
        local: '類型：本機儲存空間',
        s3: '類型：S3',
        googleDrive: '類型：Google Drive',
        nas: '類型：NAS',
        azureBlob: '類型：Azure Blob Storage',
        ftp: '類型：FTP',
        sftp: '類型：SFTP',
        rclone: '類型：Rclone',
      },
      hasSaveError: '儲存出錯',
    },
    config: {
      namePlaceholder: '輸入名稱',
      settingsTitle: '儲存空間設定',
      lastSaveError: {
        title: '儲存錯誤',
        errorLabel: '錯誤：',
        clearHint: '清除此錯誤（任選其一）：',
        clearByTestingConnection: '- 用下方按鈕測試連線（即使你已經修改了設定）；',
        clearByNextSave: '- 等待下一次儲存成功完成；',
      },
      removeConfirmation:
        '確定要刪除這個儲存空間嗎？此操作無法復原。使用該儲存空間的備份也會一併刪除。',
      removeBlockedByDatabases: '有資料庫正在使用這個儲存空間。請先在這些資料庫中移除它。',
    },
    actions: {
      testConnection: '測試連線',
    },
    connectionTestSucceeded: {
      title: '連線測試成功',
      description: '儲存空間連線測試透過',
    },
    transfer: {
      title: '將儲存空間轉移到其他工作區',
      blockedByDatabases: '有資料庫正在使用這個儲存空間。請先轉移或刪除相關資料庫。',
      description: '選擇要將此儲存空間轉移到哪個工作區。',
      targetWorkspace: '目標工作區',
      selectWorkspace: '選擇工作區',
      submit: '轉移',
    },
    fields: {
      path: '路徑',
      remotePath: '遠端路徑',
      share: '共享名',
      domain: '域',
      useSsl: '使用 SSL',
      useSslTls: '使用 SSL/TLS',
      skipTls: '跳過 TLS 驗證',
      skipTlsVerify: '跳過 TLS 驗證',
      skipHostKey: '跳過主機金鑰驗證',
      authMethod: '認證方式',
      privateKey: '私鑰',
      credentials: '認證資訊',
      s3Bucket: 'S3 儲存貯體',
      region: '區域',
      accessKey: '存取金鑰',
      secretKey: '秘密存取金鑰',
      endpoint: '端點',
      folderPrefix: '資料夾字首',
      prefix: '字首',
      blobPrefix: 'Blob 字首',
      virtualHost: '虛擬主機',
      storageClass: '儲存類別',
      connection: '連線字串',
      connectionString: '連線字串',
      accountName: '帳戶名',
      accountKey: '帳戶金鑰',
      containerName: '容器名稱',
      clientId: '客戶端 ID',
      clientSecret: '客戶端金鑰',
      userToken: '使用者令牌',
      config: '設定',
    },
    edit: {
      namePlaceholder: '我的儲存空間',
      advancedSettings: '高階設定',
      optionalPlaceholder: '{{example}}（可選）',
      usernamePlaceholder: '使用者名稱',
      passwordPlaceholder: '密碼',
      remotePathTooltip: '存放備份的遠端目錄路徑（可選）',
      local: {
        diskSpaceWarning:
          '注意：使用本機儲存空間可能會耗盡磁碟空間。建議使用 S3 或不限容量的儲存空間',
      },
      s3: {
        cloudflareR2Guide: '如何配合 Cloudflare R2 使用？',
        endpointTooltip: '自訂 S3 相容端點的 URL（可選，使用 AWS S3 時留空）',
        prefixTooltip:
          "所有物件鍵的可選字首（例如 'backups/' 或 'my_team/'）。部分 S3 相容儲存服務可能不支援。建立後無法修改（否則會遺失備份）。",
        useVirtualHostedStyle: '使用虛擬主機式域名',
        virtualHostTooltip:
          '使用虛擬主機式 URL（bucket.s3.region.amazonaws.com），而不是路徑式 URL（s3.region.amazonaws.com/bucket）。如果遇到 COS 錯誤，可能需要開啟此項。',
        skipTls: '跳過 TLS 驗證',
        skipTlsVerifyTooltip:
          '跳過 TLS 憑證驗證。如果 S3 相容儲存服務使用自我簽署憑證，請開啟此項。警告：這會降低安全性。',
        storageClassTooltip:
          '上傳物件使用的 S3 儲存類別。保持預設即為 Standard。部分服務商提供更便宜的類別，例如 One Zone IA。不要使用 Glacier/Deep Archive：還原時檔案必須能立即存取。',
      },
      googleDrive: {
        guideLink: '如何連線 Google Drive？',
        authorize: '授權',
      },
      nas: {
        shareWithSlashWarning:
          '共享名只能填寫單個共享的名稱。子目錄請填在“路徑”欄位中（例如共享名：Databasus，路徑：DB1）',
        useSslTooltip: '啟用 SSL/TLS 加密，確保連線安全',
        domainTooltip: 'Windows 域名（可選，不使用域認證時留空）',
        pathPlaceholder: '{{example}}（可選，開頭不加斜槓）',
        pathTooltip: '共享內的子目錄路徑（可選）',
      },
      azureBlob: {
        connectionStringTooltip: '從 Azure Portal 取得的 Azure Storage 連線字串',
        endpointTooltip: '自訂端點 URL（可選，使用標準 Azure 時留空）',
        prefixTooltip: "所有 Blob 名稱的可選字首（例如 'backups/' 或 'my_team/'）",
      },
      ftp: {
        enableFtps: '啟用 FTPS',
        useSslTooltip: '使用顯式 TLS 加密（FTPS）安全傳輸檔案',
        skipCertificateVerification: '跳過憑證驗證',
        skipTlsVerifyTooltip:
          '跳過 TLS 憑證驗證。如果 FTP 伺服器使用自我簽署憑證，請開啟此項。警告：這會降低安全性。',
      },
      sftp: {
        privateKeyTooltip: '貼上你的 SSH 私鑰（PEM 格式）。支援 RSA、DSA、ECDSA 和 Ed25519 金鑰。',
        skipHostKeyVerification: '跳過主機金鑰驗證',
        skipHostKeyTooltip:
          '跳過 SSH 主機金鑰驗證。如果你信任該伺服器，可以開啟此項。警告：這會降低安全性。',
      },
      rclone: {
        configTooltip:
          "在此貼上 rclone.conf 的內容。執行 'rclone config file' 可以找到該檔案，然後複製其中的內容。這份設定支援 70 多家雲儲存空間服務商。",
        configHiddenNote: '*內容已隱藏，以免洩露敏感資料。如需更新現有設定，請在此填入新的設定',
        docsLink: 'Rclone 文件',
        remotePathTooltip:
          "遠端儲存空間上存放備份的可選路徑字首（例如 '/backups' 或 'my-folder/backups'）",
      },
    },
  },
};
