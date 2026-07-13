import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const version = process.env.BUILD_VERSION || 'dev';

const config: Config = {
  title: 'iCal Filter Proxy',
  tagline: 'iCal proxy with support for user-defined filtering rules',
  favicon: 'img/favicon.ico',

  // Future flags, see https://docusaurus.io/docs/api/docusaurus-config#future
  future: {
    v4: true, // Improve compatibility with the upcoming Docusaurus v4
  },

  url: 'https://yungwood.github.io',
  baseUrl: '/ical-filter-proxy',
  organizationName: 'yungwood',
  projectName: 'ical-filter-proxy',

  onBrokenLinks: 'throw',

  customFields: {
    buildVersion: process.env.BUILD_VERSION || 'dev',
  },

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          routeBasePath: '/',
          sidebarPath: './sidebars.ts',
          editUrl:
            'https://github.com/yungwood/ical-filter-proxy/tree/main/docs/',
        },
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    image: 'img/logo.png',
    colorMode: {
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'iCal Filter Proxy',
      logo: {
        alt: 'iCal Filter Proxy Logo',
        src: 'img/logo.png',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docs',
          position: 'left',
          label: 'Docs',
        },
        {
          type: 'html',
          position: 'right',
          value: `<span class="navbar-version">Release ${version}</span>`,
        },
        {
          href: 'https://github.com/yungwood/ical-filter-proxy',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Docs',
          items: [
            {
              label: 'Quick Start',
              to: '/getting-started',
            },
            {
              label: 'Configuration',
              to: '/configuration',
            },
            {
              label: 'Recipes',
              to: '/recipes/remove-cancelled-events',
            },
          ],
        },
        {
          title: 'Project',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/yungwood/ical-filter-proxy',
            },
            {
              label: 'Docker Hub',
              href: 'https://hub.docker.com/r/yungwood/ical-filter-proxy',
            },
            {
              label: 'Helm Chart',
              href: 'https://github.com/yungwood/helm-charts/tree/main/charts/ical-filter-proxy',
            },
          ],
        },
      ],
      copyright: `Release ${version}`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
