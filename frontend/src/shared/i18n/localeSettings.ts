/* eslint-disable i18next/no-literal-string -- locale codes */
import type { Locale as AntdLocale } from 'antd/lib/locale';
import enUS from 'antd/locale/en_US';
import esES from 'antd/locale/es_ES';
import frFR from 'antd/locale/fr_FR';
import ptBR from 'antd/locale/pt_BR';
import ruRU from 'antd/locale/ru_RU';
import zhTW from 'antd/locale/zh_TW';
import 'dayjs/locale/es';
import 'dayjs/locale/fr';
import 'dayjs/locale/pt-br';
import 'dayjs/locale/ru';
import 'dayjs/locale/zh-tw';

import type { Locale } from './dictionaries';

interface LocaleSettings {
  // BCP 47 tag for the document's lang attribute and for Intl number formatting
  languageTag: string;
  dayjsLocale: string;
  antdLocale: AntdLocale;
}

export const LOCALE_SETTINGS: Record<Locale, LocaleSettings> = {
  en: { languageTag: 'en', dayjsLocale: 'en', antdLocale: enUS },
  ru: { languageTag: 'ru', dayjsLocale: 'ru', antdLocale: ruRU },
  es: { languageTag: 'es', dayjsLocale: 'es', antdLocale: esES },
  // Brazilian Portuguese is the one Portuguese variant shipped
  pt: { languageTag: 'pt-BR', dayjsLocale: 'pt-br', antdLocale: ptBR },
  // Traditional Chinese (Taiwan) is the one Chinese variant shipped
  zh: { languageTag: 'zh-TW', dayjsLocale: 'zh-tw', antdLocale: zhTW },
  fr: { languageTag: 'fr', dayjsLocale: 'fr', antdLocale: frFR },
};
