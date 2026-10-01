import { createWebKit } from '@novelia/web-kit';

export const webKit = createWebKit({
  auth: {
    app: 'f',
    url: __AUTH_URL__,
    storageKey: 'f-session',
  },
  brand: '论坛',
  repository: {
    url: 'https://github.com/auto-novel/forum',
    buildTime: __BUILD_TIME__,
    commitSha: __COMMIT_SHA__,
  },
  themeStorageKey: 'forum:theme',
});
